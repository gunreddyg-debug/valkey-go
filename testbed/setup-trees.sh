#!/usr/bin/env bash
# Makes the valkey-go trees that the bench is built against, under src/. Needs git and network access.
#
#   src/m2          upstream main, the base of the PR
#   src/m2vd        main + ClusterOption.MaxRefreshDelay (Vardhaman's draft), only to emulate main at 1000 nodes
#                   with -max-refresh-delay 30s, the bound main uses there
#   src/r2          the PR after its second review (pull/209/head): no deadline refresh, no wait for a slot
#   src/r2big       the same with nodeCount() fixed at 1000
#   src/final8      the PR as reviewed on 8 November (fork branch pr209-before-second-review)
#   src/final8big   the PR, with nodeCount() fixed at 1000, so that every default that depends on the size of
#                   the cluster takes its 1000-node value on the 6-node test cluster
#   src/finalboth   the earlier version of the PR with fixed settings, for the fixed 50 ms / 250 ms comparison
#
# Idempotent: a tree that exists is left alone. Pushing is disabled in every tree.
set -euo pipefail
cd "$(dirname "$0")"
REPO=${REPO:-https://github.com/valkey-io/valkey-go}
MAIN=${MAIN:-a23e13016ef029df8f7c092bf7e3202dfd5b0192}
PR_REF=${PR_REF:-pull/209/head}
FORK=${FORK:-https://github.com/gunreddyg-debug/valkey-go}
PR_V1_REF=${PR_V1_REF:-pr209-before-second-review}
GIT_ID=(-c user.name=testbed -c user.email=testbed@localhost)
mkdir -p src
tree() { # tree NAME REF [PATCH]; a REF of fork:BRANCH is fetched from $FORK
  local t=$1 ref=$2 patch=${3:-} src=origin
  [[ $ref == fork:* ]] && { src=$FORK; ref=${ref#fork:}; }
  [[ -d src/$t/.git ]] && { echo "src/$t exists, kept"; return; }
  git clone -q "$REPO" "src/$t"
  git -C "src/$t" remote set-url --push origin no-push
  git -C "src/$t" fetch -q "$src" "$ref"
  git -C "src/$t" checkout -q --detach FETCH_HEAD
  case $patch in
    *.mbox) git "${GIT_ID[@]}" -C "src/$t" am -q "$PWD/patches/$patch" ;;
    ?*) git -C "src/$t" apply "$PWD/patches/$patch" && git "${GIT_ID[@]}" -C "src/$t" commit -q -am "testbed: $patch" ;;
  esac
  echo "src/$t: $(git -C "src/$t" log --oneline -1)"
}
tree m2 "$MAIN"
tree m2vd "$MAIN" vardhaman-pr9-maxrefreshdelay.diff
tree r2 "$PR_REF"
tree r2big "$PR_REF" emulate-1000-nodes.patch
tree final8 "fork:$PR_V1_REF"
tree final8big "fork:$PR_V1_REF" emulate-1000-nodes.patch
tree finalboth "$MAIN" fixed-settings-split-both.mbox
