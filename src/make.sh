#!/bin/sh
# Copyright 2026 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Native bootstrap on HarmonyOS, which does not ship Bash.
# Usage: GOROOT_BOOTSTRAP=/absolute/path/to/bootstrap/go sh make.sh
set -eu
cd "$(dirname "$0")"
case "$(uname -s):$(uname -m)" in
    HarmonyOS:aarch64|OHOS:aarch64|OpenHarmony:aarch64) ;;
    *) printf 'make.sh requires an ARM64 HarmonyOS host; use make.bash on Linux.\n' >&2; exit 1 ;;
esac
command -v binary-sign-tool >/dev/null || {
    printf 'binary-sign-tool not found. Check the LLVM tool directory in PATH.\n' >&2
    exit 1
}
: "${GOROOT_BOOTSTRAP:?Set GOROOT_BOOTSTRAP to the signed bootstrap Go tree}"
GOROOT=$(CDPATH= cd .. && pwd -P)
GOROOT_BOOTSTRAP=$(CDPATH= cd "$GOROOT_BOOTSTRAP" && pwd -P)
if [ "$GOROOT" = "$GOROOT_BOOTSTRAP" ]; then
    printf 'GOROOT_BOOTSTRAP must differ from the source GOROOT.\n' >&2
    exit 1
fi
export GOROOT GOROOT_BOOTSTRAP
export GOOS=ohos GOARCH=arm64 GOENV=off GOTOOLCHAIN=local GOWORK=off
export CGO_ENABLED=${CGO_ENABLED:-1} CC=${CC:-clang} CXX=${CXX:-clang++}
if [ "$CGO_ENABLED" = 1 ]; then
    command -v "$CC" >/dev/null || {
        printf 'C compiler not found: %s; check the LLVM PATH or set CGO_ENABLED=0.\n' "$CC" >&2
        exit 1
    }
fi
export GOMAXPROCS=${GOMAXPROCS:-4}
unset GOFLAGS GOEXPERIMENT
bootstrap_host=$(GOROOT="$GOROOT_BOOTSTRAP" "$GOROOT_BOOTSTRAP/bin/go" env GOHOSTOS GOHOSTARCH)
if [ "$bootstrap_host" != "$(printf 'ohos\narm64')" ]; then
    printf 'GOROOT_BOOTSTRAP must be a signed ohos/arm64 toolchain, got: %s\n' "$bootstrap_host" >&2
    exit 1
fi
rm -f runtime/runtime_defs.go cmd/dist/dist
printf 'Building Go cmd/dist using %s\n' "$GOROOT_BOOTSTRAP"
GOROOT="$GOROOT_BOOTSTRAP" CGO_ENABLED=0 GO111MODULE=off "$GOROOT_BOOTSTRAP/bin/go" build -o cmd/dist/dist ./cmd/dist
dist_env=$(./cmd/dist/dist env -p)
eval "$dist_env"
./cmd/dist/dist bootstrap -a "$@"
rm -f cmd/dist/dist
