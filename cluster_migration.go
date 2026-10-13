package valkey

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valkey-io/valkey-go/internal/cmds"
	"github.com/valkey-io/valkey-go/internal/util"
)

// Atomic Slot Migration (ASM) awareness for Valkey 9.0+.
//
// During an atomic slot migration the source primary keeps owning and serving the slots while
// they are replicated to the target. The job then enters its finalization phase: the source
// pauses writes, the target catches up and takes ownership, and both sides report success.
// Commands sent to the source after that point are answered with MOVED.
//
// slotMigrationMonitor follows those jobs through CLUSTER GETSLOTMIGRATIONS on the primaries that
// own slots. Each primary is polled at a cadence that matches the phase of its jobs, with jitter.
// Both sides of a job in its finalization phase are polled at the fastest cadence: the target
// takes ownership first and reports success before the source does. A MOVED reply also makes the
// monitor poll the node it points to, the new owner, without waiting for its next scheduled poll,
// so a finalization phase shorter than the polling interval is not missed under traffic.
// When a job reports success, the topology is refreshed from the node that reported it: the source
// (EXPORT) or the target (IMPORT). Both already know the new owner, so the refresh does not depend
// on gossip propagation, and the IMPORT side still reports success if the source primary fails
// right after the handover. Commands are never delayed: a MOVED reply only notifies the monitor.

const (
	defaultSlotMigrationIdleInterval       = 30 * time.Second
	defaultSlotMigrationActiveInterval     = 250 * time.Millisecond
	defaultSlotMigrationFinalizingInterval = 20 * time.Millisecond

	// slotMigrationCooldown keeps a primary at the active cadence after its last active job, because
	// rebalancing usually runs many jobs back to back.
	slotMigrationCooldown = time.Minute
	// slotMigrationJitter spreads polls of many clients by ±20% of the interval.
	slotMigrationJitter = 5 // interval / 5 = 20%

	slotMigrationPollTimeout        = time.Second
	slotMigrationMaxConcurrentPolls = 16
	slotMigrationRefreshTimeout     = 3 * time.Second
	slotMigrationRefreshAttempts    = 3
	slotMigrationHandledRetention   = 10 * time.Minute
	slotMigrationMinVersion         = 9
	// slotMigrationWakeBuffer bounds the MOVED targets queued for the monitor; more are dropped, because a
	// burst of MOVED replies names the same few nodes.
	slotMigrationWakeBuffer = 16
)

// CLUSTER GETSLOTMIGRATIONS values, see https://valkey.io/commands/cluster-getslotmigrations/
const (
	slotMigrationExport = "EXPORT"
	slotMigrationImport = "IMPORT"

	slotMigrationStateSuccess   = "success"
	slotMigrationStateFailed    = "failed"
	slotMigrationStateCancelled = "cancelled"

	// finalization phase on the source (EXPORT)
	slotMigrationStateWaitingToPause  = "waiting-to-pause"
	slotMigrationStateFailoverPaused  = "failover-paused"
	slotMigrationStateFailoverGranted = "failover-granted"
	// finalization phase on the target (IMPORT)
	slotMigrationStateWaitingForPaused  = "waiting-for-paused"
	slotMigrationStateFailoverRequested = "failover-requested"
)

type slotMigrationPhase int

const (
	slotMigrationPhaseIdle       slotMigrationPhase = iota // no running job
	slotMigrationPhaseActive                               // a job is running
	slotMigrationPhaseFinalizing                           // a job is about to hand over slot ownership
)

// slotMigration is an entry of CLUSTER GETSLOTMIGRATIONS.
type slotMigration struct {
	name      string
	operation string // EXPORT on the source primary, IMPORT on the target primary
	state     string
	slots     [][2]int64
}

func (s slotMigration) phase() slotMigrationPhase {
	switch s.state {
	case slotMigrationStateSuccess, slotMigrationStateFailed, slotMigrationStateCancelled:
		return slotMigrationPhaseIdle
	case slotMigrationStateWaitingToPause, slotMigrationStateFailoverPaused, slotMigrationStateFailoverGranted:
		if s.operation == slotMigrationExport {
			return slotMigrationPhaseFinalizing
		}
	case slotMigrationStateWaitingForPaused, slotMigrationStateFailoverRequested:
		if s.operation == slotMigrationImport {
			return slotMigrationPhaseFinalizing
		}
	}
	return slotMigrationPhaseActive // includes unknown future states
}

// parseSlotMigrations extracts EXPORT and IMPORT entries from a RESP2 or RESP3 CLUSTER GETSLOTMIGRATIONS reply.
func parseSlotMigrations(reply ValkeyMessage) ([]slotMigration, error) {
	entries, err := reply.ToArray()
	if err != nil {
		return nil, err
	}
	migrations := make([]slotMigration, 0, len(entries))
	for i := range entries {
		fields, err := entries[i].AsMap() // AsMap accepts both RESP3 maps and RESP2 flat arrays
		if err != nil {
			continue
		}
		m := slotMigration{
			name:      slotMigrationField(fields, "name"),
			operation: slotMigrationField(fields, "operation"),
			state:     slotMigrationField(fields, "state"),
			slots:     parseSlotMigrationRanges(slotMigrationField(fields, "slot_ranges")),
		}
		if m.name == "" || (m.operation != slotMigrationExport && m.operation != slotMigrationImport) {
			continue
		}
		migrations = append(migrations, m)
	}
	return migrations, nil
}

func slotMigrationField(fields map[string]ValkeyMessage, key string) string {
	if v, ok := fields[key]; ok {
		s, _ := v.ToString()
		return s
	}
	return ""
}

// parseSlotMigrationRanges parses "start-end" ranges joined by spaces, for example "0-10 20-30".
func parseSlotMigrationRanges(s string) (ranges [][2]int64) {
	for _, part := range strings.Fields(s) {
		lo, hi, found := strings.Cut(part, "-")
		if !found {
			hi = lo
		}
		start, err1 := strconv.ParseInt(lo, 10, 64)
		end, err2 := strconv.ParseInt(hi, 10, 64)
		if err1 != nil || err2 != nil || start < 0 || end >= 16384 || start > end {
			continue
		}
		ranges = append(ranges, [2]int64{start, end})
	}
	return ranges
}

// serverMajorVersion returns the major server version reported by HELLO. conn.Version() is fixed to 5 on
// RESP2 connections (AlwaysRESP2 or servers without RESP3), so the HELLO reply is preferred when present.
// It returns 0 when the version is not known, for example while the connection is broken.
func serverMajorVersion(cc conn) int {
	if v, ok := cc.Info()["version"]; ok {
		if s, err := v.ToString(); err == nil {
			major, _, _ := strings.Cut(s, ".")
			if n, err := strconv.Atoi(major); err == nil {
				return n
			}
		}
	}
	return cc.Version()
}

// isSlotMigrationUnsupported reports whether an error means the node will never serve CLUSTER GETSLOTMIGRATIONS:
// servers without the command, or managed services that block it.
func isSlotMigrationUnsupported(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "unknown subcommand") ||
		strings.Contains(lower, "unknown command") ||
		strings.Contains(lower, "not allowed")
}

// slotMigrationNode is the monitor state of one primary.
type slotMigrationNode struct {
	conn        conn
	next        time.Time
	last        time.Time // last poll
	lastActive  time.Time
	seen        map[string]string // job name -> last observed state
	phase       slotMigrationPhase
	baselined   bool
	unsupported bool
}

// observe records the jobs of one poll. It returns the jobs that reached success since the previous poll and
// the most urgent phase among the jobs that are still running. Jobs found on the first poll of a node are only
// recorded, because CLUSTER GETSLOTMIGRATIONS keeps finished jobs in its log.
func (n *slotMigrationNode) observe(migrations []slotMigration) (completed []slotMigration, phase slotMigrationPhase) {
	seen := make(map[string]string, len(migrations))
	for _, m := range migrations {
		prev, known := n.seen[m.name]
		seen[m.name] = m.state
		if m.state == slotMigrationStateSuccess && n.baselined && (!known || prev != slotMigrationStateSuccess) {
			completed = append(completed, m)
		}
		if p := m.phase(); p > phase {
			phase = p
		}
	}
	n.seen = seen // jobs that dropped out of the server log are forgotten
	n.baselined = true
	n.phase = phase
	return completed, phase
}

// cadence is the phase that decides the next poll interval, including the cooldown after recent activity.
func (n *slotMigrationNode) cadence(now time.Time) slotMigrationPhase {
	if n.phase == slotMigrationPhaseIdle && !n.lastActive.IsZero() && now.Sub(n.lastActive) < slotMigrationCooldown {
		return slotMigrationPhaseActive
	}
	return n.phase
}

// slotMigrationCompletion is a job that reached success, with the node that reported it.
type slotMigrationCompletion struct {
	reporter  conn
	migration slotMigration
}

type slotMigrationMonitor struct {
	c          *clusterClient
	ctx        context.Context
	nodes      map[conn]*slotMigrationNode
	handled    map[string]time.Time // job name -> last time it was seen, for jobs already acted on
	moved      <-chan string        // addresses named by MOVED replies
	now        func() time.Time
	jitter     func(time.Duration) time.Duration
	refresh    func(completed []slotMigrationCompletion)
	idle       time.Duration
	active     time.Duration
	finalizing time.Duration
}

func newSlotMigrationMonitor(c *clusterClient) *slotMigrationMonitor {
	opt := c.opt.ClusterOption
	m := &slotMigrationMonitor{
		c:          c,
		ctx:        context.Background(),
		nodes:      make(map[conn]*slotMigrationNode),
		handled:    make(map[string]time.Time),
		moved:      c.migMoved,
		now:        time.Now,
		jitter:     jitterSlotMigrationInterval,
		idle:       opt.SlotMigrationIdleInterval,
		active:     opt.SlotMigrationActiveInterval,
		finalizing: opt.SlotMigrationFinalizingInterval,
	}
	if m.idle <= 0 {
		m.idle = defaultSlotMigrationIdleInterval
	}
	if m.active <= 0 {
		m.active = defaultSlotMigrationActiveInterval
	}
	if m.finalizing <= 0 {
		m.finalizing = defaultSlotMigrationFinalizingInterval
	}
	m.refresh = m.refreshFromReporters
	return m
}

// jitterSlotMigrationInterval returns d ± 20%, so that polls of many clients do not line up.
func jitterSlotMigrationInterval(d time.Duration) time.Duration {
	spread := d / slotMigrationJitter
	if spread <= 0 {
		return d
	}
	return d - spread + time.Duration(util.FastRand(int(2*spread)+1))
}

func (m *slotMigrationMonitor) interval(p slotMigrationPhase) time.Duration {
	switch p {
	case slotMigrationPhaseFinalizing:
		return m.finalizing
	case slotMigrationPhaseActive:
		return m.active
	default:
		return m.idle
	}
}

// run polls until the client is closed or the monitor must stop, then closes done.
// Closing the client cancels in-flight polls and refreshes.
func (m *slotMigrationMonitor) run(done chan struct{}) {
	defer close(done)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-m.c.stopCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	m.ctx = ctx

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case addr := <-m.moved:
			if m.wake(addr) {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(m.delay(m.now()))
			}
			continue
		case <-timer.C:
		}
		delay, ok := m.tick()
		if !ok {
			return
		}
		timer.Reset(delay)
	}
}

// wake brings forward the next poll of the primary at addr, after a MOVED reply pointed to it. That node is
// the new owner of the slot: when the MOVED comes from an atomic slot migration, it already reports the job's
// success, usually before the source does. It is polled no sooner than the finalizing interval after its
// previous poll while a migration is known to run, or the active interval otherwise, so a burst of MOVED
// replies, including ones from failovers or legacy migrations, costs at most a few polls. It reports
// whether the schedule changed.
func (m *slotMigrationMonitor) wake(addr string) bool {
	m.c.mu.RLock()
	cc := m.c.conns[addr].conn
	m.c.mu.RUnlock()
	n, ok := m.nodes[cc]
	if cc == nil || !ok || n.unsupported {
		return false // not a monitored primary yet; the scheduled polls cover it once it owns slots
	}
	gap := m.active
	for _, o := range m.nodes {
		if o.phase != slotMigrationPhaseIdle {
			gap = m.finalizing
			break
		}
	}
	at := n.last.Add(gap)
	if now := m.now(); at.Before(now) {
		at = now
	}
	if !at.Before(n.next) {
		return false
	}
	n.next = at
	return true
}

// primaries returns the distinct connections that own slots. With ReplicaOnly, the slot map points to
// replicas, which do not report export jobs, so every known node is polled instead.
func (m *slotMigrationMonitor) primaries() []conn {
	m.c.mu.RLock()
	defer m.c.mu.RUnlock()
	var out []conn
	set := make(map[conn]struct{})
	if m.c.opt.ReplicaOnly {
		for _, cc := range m.c.conns {
			if _, ok := set[cc.conn]; !ok && !cc.hidden && cc.conn != nil {
				set[cc.conn] = struct{}{}
				out = append(out, cc.conn)
			}
		}
		return out
	}
	var last conn
	for _, cc := range m.c.wslots {
		if cc == nil || cc == last {
			continue
		}
		last = cc
		if _, ok := set[cc]; !ok {
			set[cc] = struct{}{}
			out = append(out, cc)
		}
	}
	return out
}

// poll sends CLUSTER GETSLOTMIGRATIONS to the due nodes, at most slotMigrationMaxConcurrentPolls at a time.
// skipped[i] is true for nodes older than Valkey 9.0 or whose version is not known yet.
func (m *slotMigrationMonitor) poll(due []*slotMigrationNode) (results []ValkeyResult, skipped []bool) {
	results = make([]ValkeyResult, len(due))
	skipped = make([]bool, len(due))
	sem := make(chan struct{}, slotMigrationMaxConcurrentPolls)
	var wg sync.WaitGroup
	for i, n := range due {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, cc conn) {
			defer func() {
				<-sem
				wg.Done()
			}()
			// checked here, concurrently, because reading the version may have to dial a broken connection
			if serverMajorVersion(cc) < slotMigrationMinVersion {
				skipped[i] = true
				return
			}
			ctx, cancel := context.WithTimeout(m.ctx, slotMigrationPollTimeout)
			results[i] = cc.Do(ctx, cmds.GetSlotMigrationsCmd)
			cancel()
		}(i, n.conn)
	}
	wg.Wait()
	return results, skipped
}

// tick polls the primaries that are due and returns the delay until the next poll.
// It returns false when the monitor must stop.
func (m *slotMigrationMonitor) tick() (time.Duration, bool) {
	now := m.now()

	current := m.primaries()
	alive := make(map[conn]struct{}, len(current))
	var due []*slotMigrationNode
	for _, cc := range current {
		alive[cc] = struct{}{}
		n, ok := m.nodes[cc]
		if !ok {
			n = &slotMigrationNode{conn: cc}
			m.nodes[cc] = n
		}
		if n.unsupported || n.next.After(now) {
			continue
		}
		due = append(due, n)
	}
	for cc := range m.nodes {
		if _, ok := alive[cc]; !ok {
			delete(m.nodes, cc)
		}
	}

	results, skipped := m.poll(due)
	if m.ctx.Err() != nil {
		return 0, false // the client was closed during the poll
	}

	var completed []slotMigrationCompletion
	for i, n := range due {
		n.last = now
		if skipped[i] {
			n.next = now.Add(m.jitter(m.idle))
			continue
		}
		if err := results[i].Error(); err != nil {
			if verr, ok := IsValkeyErr(err); ok {
				msg := verr.Error()
				if strings.HasPrefix(msg, "NOPERM") {
					return 0, false // the ACL user is not allowed to run the command on this cluster
				}
				if isSlotMigrationUnsupported(msg) {
					n.unsupported = true
					continue
				}
			}
			n.next = now.Add(m.jitter(m.interval(n.cadence(now)))) // transient error, keep the cadence
			continue
		}
		migrations, err := parseSlotMigrations(results[i].val)
		if err != nil {
			n.next = now.Add(m.jitter(m.interval(n.cadence(now))))
			continue
		}
		done, phase := n.observe(migrations)
		for _, mg := range done {
			completed = append(completed, slotMigrationCompletion{reporter: n.conn, migration: mg})
		}
		for _, mg := range migrations {
			if _, ok := m.handled[mg.name]; ok {
				m.handled[mg.name] = now
			}
		}
		if phase != slotMigrationPhaseIdle {
			n.lastActive = now
		}
		n.next = now.Add(m.jitter(m.interval(n.cadence(now))))
	}

	if completed = m.claim(completed, now); len(completed) > 0 {
		m.refresh(completed)
	}
	for name, seen := range m.handled {
		if now.Sub(seen) > slotMigrationHandledRetention {
			delete(m.handled, name)
		}
	}

	return m.delay(m.now()), true
}

// delay returns the time until the next due poll, at most the idle interval.
func (m *slotMigrationMonitor) delay(now time.Time) time.Duration {
	delay := m.idle
	for _, n := range m.nodes {
		if n.unsupported {
			continue
		}
		if d := n.next.Sub(now); d < delay {
			delay = d
		}
	}
	if delay < 0 {
		delay = 0
	}
	return delay
}

// claim keeps the first completion of each job. The source (EXPORT) and the target (IMPORT) both report the
// same job; whichever is seen first triggers the refresh.
func (m *slotMigrationMonitor) claim(completed []slotMigrationCompletion, now time.Time) []slotMigrationCompletion {
	out := completed[:0]
	for _, c := range completed {
		if _, ok := m.handled[c.migration.name]; ok {
			continue
		}
		m.handled[c.migration.name] = now
		out = append(out, c)
	}
	return out
}

// refreshFromReporters refreshes the topology from the nodes that reported the completed jobs. A jittered
// lazyRefresh started by MOVED replies at cutover is expedited instead of being waited for, and the refresh
// is repeated if the migrated slots are not routed to their new owner afterwards.
func (m *slotMigrationMonitor) refreshFromReporters(completed []slotMigrationCompletion) {
	var reporters []conn
	set := make(map[conn]struct{}, len(completed))
	for _, c := range completed {
		if _, ok := set[c.reporter]; !ok {
			set[c.reporter] = struct{}{}
			reporters = append(reporters, c.reporter)
		}
	}
	ctx, cancel := context.WithTimeout(m.ctx, slotMigrationRefreshTimeout)
	defer cancel()
	for attempt := 0; attempt < slotMigrationRefreshAttempts; attempt++ {
		m.c.sc.Expedite()
		err := m.c.sc.Do(ctx, func() error { return m.c._refreshFrom(reporters) })
		if err == nil && m.routed(completed) {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(m.finalizing):
		}
	}
	if m.ctx.Err() == nil {
		m.c.lazyRefresh() // leave it to the regular refresh path
	}
}

// routed reports whether the slots of the completed jobs point to their new owner: away from the source for
// EXPORT reports, and to the target for IMPORT reports. With ReplicaOnly the slot map points to replicas, so
// it cannot be checked this way.
func (m *slotMigrationMonitor) routed(completed []slotMigrationCompletion) bool {
	if m.c.opt.ReplicaOnly {
		return true
	}
	m.c.mu.RLock()
	defer m.c.mu.RUnlock()
	for _, c := range completed {
		for _, r := range c.migration.slots {
			for i := r[0]; i <= r[1]; i++ {
				owner := m.c.wslots[i]
				if c.migration.operation == slotMigrationExport && owner == c.reporter {
					return false
				}
				if c.migration.operation == slotMigrationImport && owner != c.reporter {
					return false
				}
			}
		}
	}
	return true
}
