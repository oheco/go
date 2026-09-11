#!/bin/sh
# Cross-build a pure-Go ohos/arm64 bootstrap distribution on ARM64 Linux.
set -eu
case "$(uname -s):$(uname -m)" in
    Linux:aarch64|Linux:arm64) ;;
    *) printf 'bootstrap.sh requires ARM64 Linux.\n' >&2; exit 1 ;;
esac
root=$(CDPATH= cd "$(dirname "$0")/../.." && pwd -P)
: "${GOROOT_BOOTSTRAP:?Set GOROOT_BOOTSTRAP to an existing Linux Go tree}"
GOROOT_BOOTSTRAP=$(CDPATH= cd "$GOROOT_BOOTSTRAP" && pwd -P)
export GOROOT_BOOTSTRAP
output=${BOOTSTRAP_OUTPUT:-$root/../bootstrap/ohos/go}
case "$output" in /*) ;; *) output=$PWD/$output ;; esac
if [ -e "$output" ]; then
    printf '%s already exists; choose a new BOOTSTRAP_OUTPUT.\n' "$output" >&2
    exit 1
fi
work=$(mktemp -d "${TMPDIR:-/tmp}/go-bootstrap.XXXXXX")
stage=
trap 'rm -rf "$work"; if [ -n "$stage" ]; then rm -rf "$stage"; fi' 0
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
mkdir "$work/go"
tar -C "$root" --exclude=./.git --exclude=./pkg --exclude=./bin -cf "$work/source.tar" .
tar --touch --no-same-owner --no-same-permissions -xf "$work/source.tar" -C "$work/go"
(
    cd "$work/go/src"
    unset GOROOT GOFLAGS GOEXPERIMENT
    GOOS=ohos GOARCH=arm64 CGO_ENABLED=0 GOENV=off GOTOOLCHAIN=local GOWORK=off GOMAXPROCS=${GOMAXPROCS:-4} bash make.bash
)
# Preserve standard native tool names, as in src/bootstrap.bash.
mv "$work/go/bin/ohos_arm64/go" "$work/go/bin/go"
mv "$work/go/bin/ohos_arm64/gofmt" "$work/go/bin/gofmt"
rmdir "$work/go/bin/ohos_arm64"
mkdir -p "$(dirname "$output")"
stage=$(mktemp -d "$(dirname "$output")/.bootstrap.XXXXXX")
tar -C "$work" --exclude=go/pkg/obj --exclude=go/pkg/bootstrap --exclude=go/pkg/tool/linux_arm64 --exclude=go/pkg/linux_arm64 -cf "$work/distribution.tar" go
tar --touch --no-same-owner --no-same-permissions -xf "$work/distribution.tar" -C "$stage"
mv "$stage/go" "$output"
printf 'ohos/arm64 bootstrap distribution: %s\nSign it on HarmonyOS with misc/harmony/sign.sh before use.\n' "$output"
