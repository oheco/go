#!/bin/sh
# Native functional regressions. The optional full std run preserves all failures.
set -eu
root=$(CDPATH= cd "${1:-$(dirname "$0")/../..}" && pwd -P)
: "${TMPDIR:?Set TMPDIR to an application-private writable directory on OHOS}"
export GOROOT=$root PATH="$root/bin:$PATH" GOENV=off GOWORK=off GOTOOLCHAIN=local
export CC=${CC:-clang} CXX=${CXX:-clang++} CGO_ENABLED=1 GOMAXPROCS=${GOMAXPROCS:-4}
unset GOOS GOARCH GOFLAGS
signer=$(command -v binary-sign-tool)
work=$(mktemp -d "$TMPDIR/go-full.XXXXXX")
output=${TEST_OUTPUT:-$work/results}
mkdir -p "$output"
output=$(CDPATH= cd "$output" && pwd -P)
printf 'Regression artifacts: %s\n' "$output"
trap 'rm -rf "$work"' 0
# Preserve results by default.
if [ "$output" = "$work/results" ]; then
    trap 'printf "Results retained in %s\n" "$work"' 0
fi
sign() {
    "$signer" sign -inFile "$1" -outFile "$1.new" -selfSign 1 > "$output/sign.log" 2>&1
    chmod +x "$1.new"
    mv "$1.new" "$1"
}
cp -R "$root/misc/harmony/testdata/interop" "$work/interop"
cd "$work/interop"
for mode in exe pie c-archive c-shared; do
    go build -buildmode="$mode" -o "$output/interop-$mode" .
done
"$output/interop-exe"
"$output/interop-pie"
"$CC" caller/main.c -o "$output/caller-shared" -ldl -lpthread
sign "$output/caller-shared"
OHOS_GO_INTEROP=inherited "$output/caller-shared" "$output/interop-c-shared"
long_arg=$(printf '%05000d' 0)
OHOS_GO_INTEROP=inherited "$output/caller-shared" "$output/interop-c-shared" '' 'two words' '鸿蒙' "$long_arg"
"$CC" -DSTATIC_GO caller/main.c "$output/interop-c-archive" -o "$output/caller-archive" -ldl -lpthread
sign "$output/caller-archive"
OHOS_GO_INTEROP=inherited "$output/caller-archive"
OHOS_GO_INTEROP=inherited "$output/caller-archive" '' 'two words' '鸿蒙' "$long_arg"
cp -R "$root/misc/harmony/testdata/plugins" "$work/plugins"
cd "$work/plugins"
go build -buildmode=plugin -o "$output/extension.so" ./extension
go run ./host "$output/extension.so"
cp -R "$root/misc/harmony/testdata/diagnostics" "$work/diagnostics"
cd "$work/diagnostics"
go test -coverprofile="$output/coverage.out" -cpuprofile="$output/cpu.pprof" -trace="$output/trace.out" -bench=. -benchtime=1s
go tool cover -func="$output/coverage.out"
go tool pprof -top "$output/cpu.pprof" > "$output/pprof.txt"
go tool trace -pprof=sched "$output/trace.out" > "$output/trace-sched.pprof"
go test -fuzz=FuzzRoundTrip -fuzztime=5s
cp -R "$root/misc/harmony/testdata/detectors" "$work/detectors"
cd "$work/detectors"
for detector in race asan; do
    go build -"$detector" -o "$output/detector-$detector" .
    "$output/detector-$detector" safe
    code=0
    "$output/detector-$detector" "$detector" > "$output/$detector-error.txt" 2>&1 || code=$?
    [ "$code" -ne 0 ] || { printf '%s missed the intentional error\n' "$detector" >&2; exit 1; }
    case "$detector" in
        race) grep -q 'WARNING: DATA RACE' "$output/$detector-error.txt" ;;
        asan) grep -q 'AddressSanitizer: heap-buffer-overflow' "$output/$detector-error.txt" ;;
    esac
    printf 'PASS %s positive and negative cases\n' "$detector"
done
cd "$root/src"
go test -race -short -timeout=3m sync sync/atomic runtime/race
go test -short -timeout=3m internal/syscall/unix runtime/cgo runtime/pprof net/http/pprof crypto/internal/fips140test
if [ "${RUN_STD:-0}" = 1 ]; then
    go test -short -timeout=5m std
fi
printf 'PASS full OHOS functional regressions\n'
