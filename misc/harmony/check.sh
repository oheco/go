#!/bin/sh
# Exercise ordinary commands, executable caching and subprocesses on the host.
set -eu
: "${1:?Usage: sh check.sh /path/to/go}"
root=$(CDPATH= cd "$1" && pwd -P)
samples=$(CDPATH= cd "$(dirname "$0")/testdata" && pwd -P)
go_tool=$root/bin/go
work=$(mktemp -d "${TMPDIR:-/tmp}/go-check.XXXXXX")
trap 'rm -rf "$work"' 0
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
export GOROOT="$root" GOTOOLCHAIN=local GOENV=off GOWORK=off CGO_ENABLED=0
export GOCACHE="$work/cache" GOBIN="$work/bin" GOMAXPROCS=4
unset GOOS GOARCH GOFLAGS
"$go_tool" version
"$go_tool" env GOOS GOHOSTOS GOARCH GOHOSTARCH CGO_ENABLED GOROOT GOTOOLDIR GOTOOLCHAIN
[ "$("$go_tool" env GOOS GOHOSTOS GOARCH GOHOSTARCH)" = "$(printf 'ohos\nohos\narm64\narm64')" ]
[ "$("$go_tool" env GOTOOLDIR)" = "$root/pkg/tool/ohos_arm64" ]
case "$("$go_tool" tool dist list)" in
    *ohos/arm64*) ;;
    *) printf 'ohos/arm64 is absent from dist list.\n' >&2; exit 1 ;;
esac
cp -R "$samples/smoke" "$work/smoke"
cd "$work/smoke"
for index in 0 1; do
    files=$(GODEBUG=goindex=$index "$go_tool" list -f '{{join .GoFiles " "}}' ./constraints)
    [ "$files" = 'abi_linux_arm64.go platform_ohos.go unix.go' ] || {
        printf 'Unexpected platform source selection: %s\n' "$files" >&2
        exit 1
    }
done
printf 'PASS ohos/linux/unix constraints and architecture suffixes\n'
"$go_tool" build -o smoke .
./smoke
"$go_tool" run .
"$go_tool" run -x . 2> "$work/run-trace.log"
while IFS= read -r line; do
    case "$line" in
        */pkg/tool/ohos_arm64/compile*|*/pkg/tool/ohos_arm64/link*)
            printf 'Second go run unexpectedly rebuilt its executable: %s\n' "$line" >&2
            exit 1 ;;
    esac
done < "$work/run-trace.log"
printf 'PASS go run reused its signed executable cache\n'
"$go_tool" test -v -timeout=45s .
test_output=$("$go_tool" test -v -timeout=45s .)
printf '%s\n' "$test_output"
case "$test_output" in
    *'(cached)'*) ;;
    *) printf 'Second go test did not reuse its result cache.\n' >&2; exit 1 ;;
esac
"$go_tool" build -ldflags=-buildid= -o smoke .
./smoke -case=hello
"$go_tool" install .
"$GOBIN/harmony-smoke" -case=hello
printf 'PASS native Go build/run/test/install and executable cache\n'
