#!/bin/sh
# Sign and atomically install locally cross-built development tools.
set -eu
root=$(CDPATH= cd "$(dirname "$0")/../.." && pwd -P)
stage=${1:?Usage: sh promote-tools.sh /path/to/stage}
signer=$(command -v binary-sign-tool) || {
    printf 'binary-sign-tool not found; check the LLVM tool directory in PATH.\n' >&2
    exit 1
}
for name in go link compile; do
    case "$name" in
        go) target=$root/bin/go ;;
        *) target=$root/pkg/tool/ohos_arm64/$name ;;
    esac
    "$signer" sign -inFile "$stage/$name" -outFile "$target.new" -selfSign 1
    chmod +x "$target.new"
    mv "$target.new" "$target"
done
"$root/bin/go" version
