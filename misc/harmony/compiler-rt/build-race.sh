#!/bin/sh
# Build the Go-specific TSan object; the SDK's C/C++ TSan is not interchangeable.
set -eu
root=$(CDPATH= cd "$(dirname "$0")/../../.." && pwd -P)
llvm=${1:?Usage: sh build-race.sh /path/to/llvm-project [output.syso]}
output=${2:-$root/src/runtime/race/race_ohos_arm64.syso}
pinned=51bfeff0e4b0757ff773da6882f4d538996c9b04
if [ "$(git -C "$llvm" rev-parse HEAD)" != "$pinned" ]; then
    printf 'Expected LLVM commit %s (see runtime/race/README).\n' "$pinned" >&2
    exit 1
fi
patch=$root/misc/harmony/compiler-rt/race-ohos.patch
if git -C "$llvm" apply --check "$patch" 2>/dev/null; then
    git -C "$llvm" apply "$patch"
elif ! git -C "$llvm" apply --reverse --check "$patch" 2>/dev/null; then
    printf 'LLVM source does not match the OHOS race patch.\n' >&2
    exit 1
fi
cd "$llvm/compiler-rt/lib/tsan/go"
export GOOS=linux GOARCH=arm64 CC=${CC:-clang}
export EXTRA_CFLAGS="${EXTRA_CFLAGS:-} -DGO_OHOS=1 -Wno-unknown-warning-option"
# Use the upstream source list and both debug/release compiler checks.
# Cross builds stop before linking the upstream C test; validate it on OHOS.
sed '/^\$CC \$OSCFLAGS \$ARCHCFLAGS test.c/,$d' buildgo.sh > build-ohos-object.sh
sh build-ohos-object.sh
cp race_linux_arm64.syso "$output"
printf 'Built %s\n' "$output"
