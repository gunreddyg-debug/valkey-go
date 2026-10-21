package valkeygcp

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"testing"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/memorystore/apiv1/memorystorepb"
	"github.com/valkey-io/valkey-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	vpcA = "projects/my-proj/global/networks/vpc-a"
	vpcB = "projects/my-proj/global/networks/vpc-b"
)

func clusterInstance(auth memorystorepb.Instance_AuthorizationMode) *memorystorepb.Instance {
	return &memorystorepb.Instance{
		Name:              testInstance,
		Mode:              memorystorepb.Instance_CLUSTER,
		AuthorizationMode: auth,
		Endpoints: []*memorystorepb.Instance_InstanceEndpoint{endpointGroup(
			autoConn("10.0.0.15", 6379, memorystorepb.ConnectionType_CONNECTION_TYPE_DISCOVERY, vpcA),
		)},
	}
}

// T-DISC-1
func TestResolve_DiscoverClusterIAMPerInstanceCA(t *testing.T) {
	ca1, ca2 := newTestCA(t, "ca-1"), newTestCA(t, "ca-2")
	inst := clusterInstance(memorystorepb.Instance_IAM_AUTH)
	inst.TransitEncryptionMode = memorystorepb.Instance_SERVER_AUTHENTICATION
	inst.ServerCaMode = memorystorepb.Instance_GOOGLE_MANAGED_PER_INSTANCE_CA.Enum()
	fake := &fakeReader{inst: inst, perCA: []string{string(ca1.pem), string(ca2.pem)}}

	r, err := resolve(context.Background(), Config{Instance: testInstance, Auth: AuthOptions{IAM: staticIAM()}}, fake.factory())
	if err != nil {
		t.Fatal(err)
	}
	if r.auth != AuthIAM || r.tls != TLSPerInstanceCA || r.topology != TopologyCluster {
		t.Fatalf("resolved %v/%v/%v", r.auth, r.tls, r.topology)
	}
	if fake.gotName != testInstance || fake.perCalls != 1 || fake.perName != testInstance || fake.sharedCalls != 0 {
		t.Fatalf("calls: get=%q per=%d(%q) shared=%d", fake.gotName, fake.perCalls, fake.perName, fake.sharedCalls)
	}
	if !fake.closed {
		t.Error("control-plane client not closed after resolution")
	}
	opt, err := r.clientOption(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if opt.InitAddress[0] != "10.0.0.15:6379" || opt.TLSConfig == nil || opt.AuthCredentialsFn == nil {
		t.Fatalf("option: %v tls=%v auth=%v", opt.InitAddress, opt.TLSConfig != nil, opt.AuthCredentialsFn != nil)
	}
	if fake.perCalls != 1 {
		t.Errorf("PerInstanceCA called %d times", fake.perCalls)
	}
}

// T-DISC-2
func TestResolve_DiscoverClusterDisabledTokenSharedCA(t *testing.T) {
	ca := newTestCA(t, "regional")
	inst := &memorystorepb.Instance{
		Mode:                  memorystorepb.Instance_CLUSTER_DISABLED,
		AuthorizationMode:     memorystorepb.Instance_TOKEN_AUTH,
		TransitEncryptionMode: memorystorepb.Instance_SERVER_AUTHENTICATION,
		ServerCaMode:          memorystorepb.Instance_GOOGLE_MANAGED_SHARED_CA.Enum(),
		Endpoints: []*memorystorepb.Instance_InstanceEndpoint{endpointGroup(
			autoConn("10.0.0.2", 6379, memorystorepb.ConnectionType_CONNECTION_TYPE_PRIMARY, vpcA),
			autoConn("10.0.0.3", 6379, memorystorepb.ConnectionType_CONNECTION_TYPE_READER, vpcA),
		)},
	}
	fake := &fakeReader{inst: inst, sharedCA: []string{string(ca.pem)}}
	cfg := Config{Instance: testInstance, Auth: AuthOptions{
		IAM:   staticIAM(), // ignored: the instance decides
		Token: TokenAuthOptions{Token: "tok"},
	}}
	r, err := resolve(context.Background(), cfg, fake.factory())
	if err != nil {
		t.Fatal(err)
	}
	if r.auth != AuthToken || r.tls != TLSSharedCA || r.topology != TopologyClusterDisabled {
		t.Fatalf("resolved %v/%v/%v", r.auth, r.tls, r.topology)
	}
	if fake.sharedCalls != 1 || fake.sharedProject != "my-proj" || fake.sharedRegion != "us-west4" || fake.perCalls != 0 {
		t.Fatalf("shared CA call: %d %q %q (per=%d)", fake.sharedCalls, fake.sharedProject, fake.sharedRegion, fake.perCalls)
	}
	opt, err := r.clientOption(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if opt.InitAddress[0] != "10.0.0.2:6379" || opt.Standalone.ReplicaAddress[0] != "10.0.0.3:6379" || opt.SendToReplicas == nil {
		t.Fatalf("endpoints: %v %v", opt.InitAddress, opt.Standalone.ReplicaAddress)
	}
	creds, err := opt.AuthCredentialsFn(valkey.AuthCredentialsContext{})
	if err != nil || creds.Username != "" || creds.Password != "tok" {
		t.Fatalf("token creds = %+v, %v", creds, err)
	}
}

// T-DISC-3
func TestResolve_DiscoverNoAuthNoTLS(t *testing.T) {
	inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
	inst.TransitEncryptionMode = memorystorepb.Instance_TRANSIT_ENCRYPTION_DISABLED
	fake := &fakeReader{inst: inst}
	opt, err := resolveOption(t, Config{Instance: testInstance, Auth: AuthOptions{IAM: staticIAM()}}, fake.factory())
	if err != nil {
		t.Fatal(err)
	}
	if opt.AuthCredentialsFn != nil || opt.TLSConfig != nil || fake.perCalls+fake.sharedCalls != 0 {
		t.Fatalf("auth=%v tls=%v caCalls=%d", opt.AuthCredentialsFn != nil, opt.TLSConfig != nil, fake.perCalls+fake.sharedCalls)
	}
}

// T-DISC-4
func TestResolve_DiscoverContradictions(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{"auth", Config{Auth: AuthOptions{Mode: AuthToken, Token: TokenAuthOptions{Token: "t"}}}, "instance uses IAM_AUTH but Auth.Mode is AuthToken"},
		{"tls", Config{TLS: TLSOptions{Mode: TLSSharedCA}}, "instance uses TRANSIT_ENCRYPTION_DISABLED but TLS.Mode is TLSSharedCA"},
		{"topology", Config{Topology: TopologyClusterDisabled}, "instance mode is CLUSTER but Topology is TopologyClusterDisabled"},
		{"endpoint", Config{PrimaryEndpoint: "10.9.9.9:6379"}, "PrimaryEndpoint and ReaderEndpoint are for cluster-disabled"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inst := clusterInstance(memorystorepb.Instance_IAM_AUTH)
			inst.TransitEncryptionMode = memorystorepb.Instance_TRANSIT_ENCRYPTION_DISABLED
			c.cfg.Instance = testInstance
			_, err := resolve(context.Background(), c.cfg, (&fakeReader{inst: inst}).factory())
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want %q", err, c.want)
			}
		})
	}
}

// T-DISC-5
func TestResolve_DiscoverTokenAuthWithoutSource(t *testing.T) {
	fake := &fakeReader{inst: clusterInstance(memorystorepb.Instance_TOKEN_AUTH)}
	_, err := resolve(context.Background(), Config{Instance: testInstance, Auth: AuthOptions{IAM: staticIAM()}}, fake.factory())
	if !errors.Is(err, errNoTokenSource) || !strings.Contains(err.Error(), "Auth.Token.SecretVersion") {
		t.Fatalf("got %v", err)
	}
}

// T-DISC-6
func TestResolve_DiscoverPermissionDenied(t *testing.T) {
	denied := status.Error(codes.PermissionDenied, "Permission 'memorystore.instances.get' denied")
	_, err := resolve(context.Background(), Config{Instance: testInstance}, (&fakeReader{getErr: denied}).factory())
	if err == nil || !strings.Contains(err.Error(), "memorystore.instances.get") || !strings.Contains(err.Error(), "roles/memorystore.viewer") {
		t.Fatalf("got %v", err)
	}
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("gRPC status lost in wrapping: %v", status.Code(err))
	}

	// CA download needs the same permission.
	inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
	inst.TransitEncryptionMode = memorystorepb.Instance_SERVER_AUTHENTICATION
	_, err = resolve(context.Background(), Config{Instance: testInstance}, (&fakeReader{inst: inst, caErr: denied}).factory())
	if err == nil || !strings.Contains(err.Error(), "roles/memorystore.viewer") || !strings.Contains(err.Error(), "server CA") {
		t.Fatalf("CA download: got %v", err)
	}

	// Other errors pass through without the permission hint.
	_, err = resolve(context.Background(), Config{Instance: testInstance}, (&fakeReader{getErr: status.Error(codes.NotFound, "no such instance")}).factory())
	if err == nil || strings.Contains(err.Error(), "roles/memorystore.viewer") {
		t.Fatalf("NotFound: got %v", err)
	}

	// The reader factory's own error is returned as is.
	boom := errors.New("no ADC")
	_, err = resolve(context.Background(), Config{Instance: testInstance}, func(context.Context) (instanceReader, error) { return nil, boom })
	if !errors.Is(err, boom) {
		t.Fatalf("factory error: got %v", err)
	}
}

// T-DISC-7: Network selects among the networks without the metadata server. Without
// Network, the first endpoint is no longer assumed: the VM's network decides, or it fails.
func TestResolve_DiscoverNetworkSelection(t *testing.T) {
	inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
	inst.Endpoints = []*memorystorepb.Instance_InstanceEndpoint{
		endpointGroup(autoConn("10.0.0.15", 6379, memorystorepb.ConnectionType_CONNECTION_TYPE_DISCOVERY, vpcA)),
		endpointGroup(userConn("10.1.0.15", 6379, memorystorepb.ConnectionType_CONNECTION_TYPE_DISCOVERY, vpcB)),
	}
	discovered := func(cfg Config) (string, error) {
		cfg.Instance = testInstance
		r, err := resolve(context.Background(), cfg, (&fakeReader{inst: inst}).factory())
		if err != nil {
			return "", err
		}
		return r.discovery, nil
	}
	useVM(t, nil)
	for network, want := range map[string]string{
		vpcA: "10.0.0.15:6379",
		vpcB: "10.1.0.15:6379",
		"https://www.googleapis.com/compute/v1/" + vpcB: "10.1.0.15:6379",
		"vpc-b": "10.1.0.15:6379",
	} {
		if got, err := discovered(Config{Network: network}); err != nil || got != want {
			t.Errorf("network %q: got %q, %v; want %q", network, got, err, want)
		}
	}
	if _, err := discovered(Config{Network: "vpc-c"}); err == nil || !strings.Contains(err.Error(), `in network "vpc-c"`) {
		t.Fatalf("unknown network: got %v", err)
	}

	useVM(t, &fakeVM{network: "projects/111/networks/vpc-b", projectNum: "111", projectID: "my-proj"})
	if got, err := discovered(Config{}); err != nil || got != "10.1.0.15:6379" {
		t.Fatalf("VM in vpc-b: got %q, %v", got, err)
	}
	useVM(t, &fakeVM{offGCE: true})
	if _, err := discovered(Config{}); err == nil || !strings.Contains(err.Error(), "set Config.Network") {
		t.Fatalf("off GCE: got %v", err)
	}
}

// T-DISC-8
func TestResolve_DiscoverDeprecatedDiscoveryEndpoints(t *testing.T) {
	inst := &memorystorepb.Instance{
		Mode:               memorystorepb.Instance_CLUSTER,
		DiscoveryEndpoints: []*memorystorepb.DiscoveryEndpoint{{Address: "10.0.0.20", Port: 6380, Network: vpcA}}, //nolint:staticcheck // SA1019: covers the fallback for the deprecated field
	}
	r, err := resolve(context.Background(), Config{Instance: testInstance}, (&fakeReader{inst: inst}).factory())
	if err != nil || r.discovery != "10.0.0.20:6380" {
		t.Fatalf("got %q, %v", r.discovery, err)
	}
}

// T-DISC-9
func TestResolve_DiscoverCustomerManagedCANeedsPEM(t *testing.T) {
	inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
	inst.TransitEncryptionMode = memorystorepb.Instance_SERVER_AUTHENTICATION
	inst.ServerCaMode = memorystorepb.Instance_CUSTOMER_MANAGED_CAS_CA.Enum()
	fake := &fakeReader{inst: inst}
	_, err := resolve(context.Background(), Config{Instance: testInstance}, fake.factory())
	if err == nil || !strings.Contains(err.Error(), "needs CA certificates") || strings.Contains(err.Error(), "Config.Instance") {
		t.Fatalf("got %v", err)
	}
	if fake.perCalls+fake.sharedCalls != 0 {
		t.Error("customer-managed CAs can't be downloaded; no CA call expected")
	}
	ca := newTestCA(t, "cas-root")
	opt, err := resolveOption(t, Config{Instance: testInstance, TLS: TLSOptions{CAPEM: ca.pem}}, (&fakeReader{inst: inst}).factory())
	if err != nil || opt.TLSConfig == nil {
		t.Fatalf("with CAPEM: %v", err)
	}
}

// T-DISC-10
func TestResolve_DiscoverUnspecifiedCAModeIsPerInstance(t *testing.T) {
	ca := newTestCA(t, "legacy")
	inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
	inst.TransitEncryptionMode = memorystorepb.Instance_SERVER_AUTHENTICATION
	inst.ServerCaMode = memorystorepb.Instance_SERVER_CA_MODE_UNSPECIFIED.Enum()
	fake := &fakeReader{inst: inst, perCA: []string{string(ca.pem)}}
	r, err := resolve(context.Background(), Config{Instance: testInstance}, fake.factory())
	if err != nil || r.tls != TLSPerInstanceCA || fake.perCalls != 1 {
		t.Fatalf("got %v, %v (per=%d)", r.tls, err, fake.perCalls)
	}
}

// T-DISC-11
func TestResolve_DiscoverPortZeroDefaults(t *testing.T) {
	inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
	inst.Endpoints = []*memorystorepb.Instance_InstanceEndpoint{endpointGroup(
		autoConn("10.0.0.15", 0, memorystorepb.ConnectionType_CONNECTION_TYPE_DISCOVERY, vpcA),
	)}
	r, err := resolve(context.Background(), Config{Instance: testInstance}, (&fakeReader{inst: inst}).factory())
	if err != nil || r.discovery != "10.0.0.15:6379" {
		t.Fatalf("got %q, %v", r.discovery, err)
	}
}

func TestResolve_DiscoverEdgeCases(t *testing.T) {
	t.Run("malformed instance name", func(t *testing.T) {
		for _, name := range []string{"my-valkey", "projects/p/locations/r/clusters/c", "projects//locations/r/instances/i", testInstance + "/x"} {
			if _, err := resolve(context.Background(), Config{Instance: name}, noReader(t)); err == nil {
				t.Errorf("%q accepted", name)
			}
		}
	})
	t.Run("explicit endpoint overrides discovery", func(t *testing.T) {
		r, err := resolve(context.Background(), Config{Instance: testInstance, DiscoveryEndpoint: "proxy:6379"},
			(&fakeReader{inst: clusterInstance(memorystorepb.Instance_AUTH_DISABLED)}).factory())
		if err != nil || r.discovery != "proxy:6379" {
			t.Fatalf("got %q, %v", r.discovery, err)
		}
	})
	t.Run("no PSC endpoint yet", func(t *testing.T) {
		_, err := resolve(context.Background(), Config{Instance: testInstance},
			(&fakeReader{inst: &memorystorepb.Instance{Mode: memorystorepb.Instance_CLUSTER}}).factory())
		if err == nil || !strings.Contains(err.Error(), "CONNECTION_TYPE_DISCOVERY") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("explicit TLSCustomCA on a Google-managed CA", func(t *testing.T) {
		ca := newTestCA(t, "offline-bundle")
		inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
		inst.TransitEncryptionMode = memorystorepb.Instance_SERVER_AUTHENTICATION
		fake := &fakeReader{inst: inst}
		r, err := resolve(context.Background(), Config{Instance: testInstance, TLS: TLSOptions{Mode: TLSCustomCA, CAPEM: ca.pem}}, fake.factory())
		if err != nil || r.tls != TLSCustomCA || fake.perCalls != 0 {
			t.Fatalf("got %v, %v (per=%d)", r.tls, err, fake.perCalls)
		}
	})
	t.Run("caller TLSConfig on a plain-text instance", func(t *testing.T) {
		inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
		inst.TransitEncryptionMode = memorystorepb.Instance_TRANSIT_ENCRYPTION_DISABLED
		cfg := Config{Instance: testInstance}
		cfg.ClientOption.TLSConfig = &tls.Config{}
		_, err := resolve(context.Background(), cfg, (&fakeReader{inst: inst}).factory())
		if err == nil || !strings.Contains(err.Error(), "in-transit encryption disabled") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("API returns no CA certificate", func(t *testing.T) {
		inst := clusterInstance(memorystorepb.Instance_AUTH_DISABLED)
		inst.TransitEncryptionMode = memorystorepb.Instance_SERVER_AUTHENTICATION
		_, err := resolve(context.Background(), Config{Instance: testInstance}, (&fakeReader{inst: inst}).factory())
		if err == nil || !strings.Contains(err.Error(), "no server CA certificate") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("legacy STANDALONE mode", func(t *testing.T) {
		inst := &memorystorepb.Instance{
			Mode: memorystorepb.Instance_Mode(1), // STANDALONE, the deprecated name of CLUSTER_DISABLED
			Endpoints: []*memorystorepb.Instance_InstanceEndpoint{endpointGroup(
				autoConn("10.0.0.2", 6379, memorystorepb.ConnectionType_CONNECTION_TYPE_PRIMARY, ""),
			)},
		}
		r, err := resolve(context.Background(), Config{Instance: testInstance}, (&fakeReader{inst: inst}).factory())
		if err != nil || r.topology != TopologyClusterDisabled || r.primary != "10.0.0.2:6379" || r.reader != "" {
			t.Fatalf("got %+v, %v", r, err)
		}
	})
	t.Run("unknown enum values", func(t *testing.T) {
		for _, inst := range []*memorystorepb.Instance{
			{Mode: memorystorepb.Instance_Mode(99)},
			{AuthorizationMode: memorystorepb.Instance_AuthorizationMode(99)},
			{TransitEncryptionMode: memorystorepb.Instance_TransitEncryptionMode(99)},
			{TransitEncryptionMode: memorystorepb.Instance_SERVER_AUTHENTICATION, ServerCaMode: memorystorepb.Instance_ServerCaMode(99).Enum()},
		} {
			if _, err := resolve(context.Background(), Config{Instance: testInstance}, (&fakeReader{inst: inst}).factory()); err == nil ||
				!strings.Contains(err.Error(), "unsupported") {
				t.Errorf("%v accepted: %v", inst, err)
			}
		}
	})
}

// fakeVM stands in for the metadata server in network detection.
type fakeVM struct {
	offGCE                         bool
	network, projectNum, projectID string // network as "projects/<number>/networks/<name>"
	fail                           string // metadata path that returns an error
}

// useVM points the metadata seams at vm until the test ends. With a nil vm, any metadata
// call fails the test.
func useVM(t *testing.T, vm *fakeVM) {
	t.Helper()
	onGCE, get, num, id := metadataOnGCE, metadataGet, metadataNumericProjectID, metadataProjectID
	t.Cleanup(func() { metadataOnGCE, metadataGet, metadataNumericProjectID, metadataProjectID = onGCE, get, num, id })
	unexpected := vm == nil
	if unexpected {
		vm = &fakeVM{offGCE: true}
	}
	read := func(path, v string) (string, error) {
		if unexpected {
			t.Errorf("metadata server read %s; want no network detection", path)
		}
		if path == vm.fail {
			return "", metadata.NotDefinedError(path)
		}
		return v, nil
	}
	metadataOnGCE = func(context.Context) bool {
		if unexpected {
			t.Error("metadata server probed; want no network detection")
		}
		return !vm.offGCE
	}
	metadataGet = func(_ context.Context, path string) (string, error) {
		if path != "instance/network-interfaces/0/network" {
			t.Errorf("unexpected metadata read %s", path)
		}
		return read(path, vm.network)
	}
	metadataNumericProjectID = func(context.Context) (string, error) { return read("project/numeric-project-id", vm.projectNum) }
	metadataProjectID = func(context.Context) (string, error) { return read("project/project-id", vm.projectID) }
}

// T-DISC-12: the endpoint of each connection type is the one in the client's VPC network.
func TestResolve_DiscoverEndpointSelection(t *testing.T) {
	const (
		disc    = memorystorepb.ConnectionType_CONNECTION_TYPE_DISCOVERY
		primary = memorystorepb.ConnectionType_CONNECTION_TYPE_PRIMARY
		reader  = memorystorepb.ConnectionType_CONNECTION_TYPE_READER

		ownDefault  = "projects/my-proj/global/networks/default"
		hostDefault = "projects/host-proj/global/networks/default" // a Shared VPC host project
		hostShared  = "projects/host-proj/global/networks/shared"
	)
	cluster := func(groups ...*memorystorepb.Instance_InstanceEndpoint) *memorystorepb.Instance {
		return &memorystorepb.Instance{Mode: memorystorepb.Instance_CLUSTER, Endpoints: groups}
	}
	discIn := func(ip, network string) *memorystorepb.Instance_InstanceEndpoint { // one group per network, like the API
		return endpointGroup(autoConn(ip, 6379, disc, network))
	}
	twoVPCs := cluster(discIn("10.0.0.15", vpcA), discIn("10.1.0.15", vpcB))
	sameName := cluster(discIn("10.2.0.15", hostDefault), discIn("10.0.0.15", ownDefault))
	deprecated := &memorystorepb.Instance{
		Mode: memorystorepb.Instance_CLUSTER,
		DiscoveryEndpoints: []*memorystorepb.DiscoveryEndpoint{ //nolint:staticcheck // SA1019: covers the fallback for the deprecated field
			{Address: "10.0.0.20", Network: vpcA}, {Address: "10.1.0.20", Network: vpcB},
		},
	}
	clusterDisabled := &memorystorepb.Instance{
		Mode: memorystorepb.Instance_CLUSTER_DISABLED,
		Endpoints: []*memorystorepb.Instance_InstanceEndpoint{
			endpointGroup(autoConn("10.0.0.2", 6379, primary, vpcA), autoConn("10.0.0.3", 6379, reader, vpcA)),
			endpointGroup(autoConn("10.1.0.2", 6379, primary, vpcB), autoConn("10.1.0.3", 6379, reader, vpcB)),
		},
	}
	vmIn := func(network string) *fakeVM { // a VM in project my-proj (number 111)
		return &fakeVM{network: network, projectNum: "111", projectID: "my-proj"}
	}
	// several is what an ambiguity error names: instance, connection type, candidates, fix.
	several := func(typ memorystorepb.ConnectionType, nets ...string) []string {
		return append([]string{"instance " + testInstance, typ.String() + " endpoints in several networks",
			"set Config.Network to one of them"}, nets...)
	}

	cases := []struct {
		name   string
		inst   *memorystorepb.Instance
		cfg    Config
		vm     *fakeVM // nil: the metadata server must not be used
		want   string  // discovery or primary endpoint
		reader string
		err    []string
	}{
		// One network: the first endpoint, as before; Network only has to name it.
		{name: "one network", inst: cluster(discIn("10.0.0.15", vpcA), discIn("10.0.0.16", vpcA)), want: "10.0.0.15:6379"},
		{name: "one network, Network path", inst: cluster(discIn("10.0.0.15", vpcA)), cfg: Config{Network: vpcA}, want: "10.0.0.15:6379"},
		{name: "one network, Network name", inst: cluster(discIn("10.0.0.15", vpcA)), cfg: Config{Network: "vpc-a"}, want: "10.0.0.15:6379"},
		{name: "one network, other Network", inst: cluster(discIn("10.0.0.15", vpcA)), cfg: Config{Network: "vpc-b"},
			err: []string{`instance ` + testInstance + ` has no CONNECTION_TYPE_DISCOVERY endpoint in network "vpc-b"`}},
		{name: "no network reported", inst: cluster(discIn("10.0.0.15", ""), discIn("10.1.0.15", "")), cfg: Config{Network: "vpc-b"}, want: "10.0.0.15:6379"},

		// Several networks and Network set.
		{name: "Network path", inst: twoVPCs, cfg: Config{Network: vpcB}, want: "10.1.0.15:6379"},
		{name: "Network URL", inst: twoVPCs, cfg: Config{Network: "https://www.googleapis.com/compute/v1/" + vpcB}, want: "10.1.0.15:6379"},
		{name: "Network name", inst: twoVPCs, cfg: Config{Network: "vpc-b"}, want: "10.1.0.15:6379"},
		{name: "Network name in two projects", inst: sameName, cfg: Config{Network: "default"},
			err: append(several(disc, hostDefault, ownDefault), `the network name "default" matches more than one`)},
		{name: "Network path, name in two projects", inst: sameName, cfg: Config{Network: hostDefault}, want: "10.2.0.15:6379"},
		{name: "Network path in a third project", inst: sameName, cfg: Config{Network: "projects/other/global/networks/default"},
			err: []string{`no CONNECTION_TYPE_DISCOVERY endpoint in network "projects/other/global/networks/default"`}},

		// Several networks and no Network: the VM's network decides.
		{name: "off GCE", inst: twoVPCs, vm: &fakeVM{offGCE: true}, err: append(several(disc, vpcA, vpcB), "can't be detected off GCE")},
		{name: "VM network detected", inst: twoVPCs, vm: vmIn("projects/111/networks/vpc-b"), want: "10.1.0.15:6379"},
		{name: "VM in neither network", inst: twoVPCs, vm: vmIn("projects/111/networks/vpc-c"),
			err: append(several(disc, vpcA, vpcB), "the VM's network projects/111/networks/vpc-c is none of them")},
		{name: "VM network unreadable", inst: twoVPCs, vm: &fakeVM{fail: "instance/network-interfaces/0/network"},
			err: append(several(disc, vpcA, vpcB), "can't read the VM's network", "not defined")},
		{name: "name in two projects, VM's own project", inst: sameName, vm: vmIn("projects/111/networks/default"), want: "10.0.0.15:6379"},
		{name: "name in two projects, Shared VPC", inst: sameName, vm: vmIn("projects/222/networks/default"),
			err: append(several(disc, hostDefault, ownDefault), "the VM's network projects/222/networks/default is in another project")},
		{name: "Shared VPC, unique name", inst: cluster(discIn("10.0.0.15", vpcA), discIn("10.2.0.15", hostShared)),
			vm: vmIn("projects/222/networks/shared"), want: "10.2.0.15:6379"},
		{name: "own project not among them", inst: cluster(discIn("10.2.0.15", hostDefault), discIn("10.3.0.15", "projects/third/global/networks/default")),
			vm: vmIn("projects/111/networks/default"), err: append(several(disc), "(project my-proj) is none of them")},
		{name: "project number unreadable", inst: sameName,
			vm: &fakeVM{network: "projects/111/networks/default", fail: "project/numeric-project-id"}, err: append(several(disc), "project number")},
		{name: "project ID unreadable", inst: sameName,
			vm: &fakeVM{network: "projects/111/networks/default", projectNum: "111", fail: "project/project-id"}, err: append(several(disc), "project ID")},

		// Other endpoint sources follow the same rules.
		{name: "user-registered PSC connections", vm: vmIn("projects/111/networks/vpc-b"), want: "10.1.0.15:6379", inst: cluster(
			endpointGroup(userConn("10.0.0.15", 6379, disc, vpcA)), endpointGroup(userConn("10.1.0.15", 6379, disc, vpcB)))},
		{name: "user-registered PSC connections, Network", cfg: Config{Network: "vpc-a"}, want: "10.0.0.15:6379", inst: cluster(
			endpointGroup(userConn("10.0.0.15", 6379, disc, vpcA)), endpointGroup(userConn("10.1.0.15", 6379, disc, vpcB)))},
		{name: "deprecated discovery_endpoints, Network", inst: deprecated, cfg: Config{Network: "vpc-b"}, want: "10.1.0.20:6379"},
		{name: "deprecated discovery_endpoints, detected", inst: deprecated, vm: vmIn("projects/111/networks/vpc-a"), want: "10.0.0.20:6379"},
		{name: "deprecated discovery_endpoints, off GCE", inst: deprecated, vm: &fakeVM{offGCE: true}, err: several(disc, vpcA, vpcB)},

		// Endpoints that replace the discovered one skip selection.
		{name: "DiscoveryEndpoint set", inst: twoVPCs, cfg: Config{DiscoveryEndpoint: "proxy:6379"}, want: "proxy:6379"},
		{name: "ClientOption.InitAddress set", inst: twoVPCs, cfg: Config{ClientOption: valkey.ClientOption{InitAddress: []string{"proxy:6379"}}}},

		// Cluster-disabled: PRIMARY and READER.
		{name: "cluster-disabled, Network", inst: clusterDisabled, cfg: Config{Network: "vpc-b"}, want: "10.1.0.2:6379", reader: "10.1.0.3:6379"},
		{name: "cluster-disabled, detected", inst: clusterDisabled, vm: vmIn("projects/111/networks/vpc-a"), want: "10.0.0.2:6379", reader: "10.0.0.3:6379"},
		{name: "cluster-disabled, off GCE", inst: clusterDisabled, vm: &fakeVM{offGCE: true}, err: several(primary, vpcA, vpcB)},
		{name: "cluster-disabled, PrimaryEndpoint set, off GCE", inst: clusterDisabled, cfg: Config{PrimaryEndpoint: "p:6379"},
			vm: &fakeVM{offGCE: true}, err: several(reader, vpcA, vpcB)},
		{name: "cluster-disabled, ReaderEndpoint set", inst: clusterDisabled, cfg: Config{Network: vpcA, ReaderEndpoint: "r:6379"},
			want: "10.0.0.2:6379", reader: "r:6379"},
		{name: "cluster-disabled, caller replicas", inst: clusterDisabled, want: "p:6379", cfg: Config{PrimaryEndpoint: "p:6379",
			ClientOption: valkey.ClientOption{Standalone: valkey.StandaloneOption{ReplicaAddress: []string{"r:6379"}}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			useVM(t, c.vm)
			c.cfg.Instance = testInstance
			r, err := resolve(context.Background(), c.cfg, (&fakeReader{inst: c.inst}).factory())
			if len(c.err) > 0 {
				for _, want := range c.err {
					if err == nil || !strings.Contains(err.Error(), want) {
						t.Fatalf("got error %v; want it to contain %q", err, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := r.discovery
			if r.topology == TopologyClusterDisabled {
				got = r.primary
			}
			if got != c.want || r.reader != c.reader {
				t.Fatalf("got %q (reader %q); want %q (reader %q)", got, r.reader, c.want, c.reader)
			}
		})
	}
}
