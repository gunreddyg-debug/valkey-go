module localfailover/bench

go 1.25.0

require github.com/valkey-io/valkey-go v1.0.0

require golang.org/x/sys v0.47.0 // indirect

// The upstream build. go.pr1.mod is the same file with the replace pointing at ../src/pr1;
// build.sh builds the pr1 binary with `-modfile=go.pr1.mod -tags pr1`.
replace github.com/valkey-io/valkey-go => ../src/upstream
