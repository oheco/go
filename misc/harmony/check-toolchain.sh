#!/bin/sh
# Native upstream suites. Keep separate logs and preserve every exit status.
set -eu
root=$(CDPATH= cd "${1:-$(dirname "$0")/../..}" && pwd -P)
suite=${2:-all}
: "${TMPDIR:?Set TMPDIR to an application-private writable directory on OHOS}"
export GOROOT=$root PATH="$root/bin:$PATH" GOENV=off GOWORK=off GOTOOLCHAIN=local
export CGO_ENABLED=1 GOMAXPROCS=${GOMAXPROCS:-4}
export CC="$root/misc/harmony/cc-sign.sh" CXX="$root/misc/harmony/cxx-sign.sh"
unset GOOS GOARCH GOFLAGS GOEXPERIMENT
command -v binary-sign-tool >/dev/null || {
    printf 'binary-sign-tool not found; check the LLVM tool directory in PATH.\n' >&2
    exit 1
}
output=${TEST_OUTPUT:-$(mktemp -d "$TMPDIR/go-toolchain.XXXXXX")}
mkdir -p "$output"
output=$(CDPATH= cd "$output" && pwd -P)
printf 'Upstream test logs: %s\n' "$output"
go version > "$output/environment.txt"
go env GOOS GOARCH CGO_ENABLED CC CXX >> "$output/environment.txt"
: > "$output/status.tsv"
case "$suite" in
    all) suites='paths cgo tools std' ;;
    paths|cgo|tools|std) suites=$suite ;;
    *) printf 'Unknown suite: %s\n' "$suite" >&2; exit 2 ;;
esac
cd "$root/src"
result=0
for name in $suites; do
    code=0
    case "$name" in
        paths) set -- os path/filepath ;;
        cgo) set -- cmd/cgo/internal/... ;;
        tools) set -- -short cmd/... ;;
        std) set -- -short std ;;
    esac
    go test -v -count=1 -p=2 -timeout=10m "$@" > "$output/$name.log" 2>&1 || code=$?
    printf '%s\t%s\n' "$name" "$code" >> "$output/status.tsv"
    printf '%s: exit %s\n' "$name" "$code"
    [ "$code" = 0 ] || result=1
done
exit "$result"
