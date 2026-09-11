#!/bin/sh
# Sign a Linux-built pure-Go bootstrap tree on the HarmonyOS host.
set -eu
case "$(uname -s):$(uname -m)" in
    HarmonyOS:aarch64|OHOS:aarch64|OpenHarmony:aarch64) ;;
    *) printf 'Run sign.sh on ARM64 HarmonyOS.\n' >&2; exit 1 ;;
esac
: "${1:?Usage: sh sign.sh /path/to/bootstrap/go}"
root=$(CDPATH= cd "$1" && pwd -P)
tool=$(command -v binary-sign-tool) || {
    printf 'binary-sign-tool not found. Check the LLVM tool directory in PATH.\n' >&2
    exit 1
}
work=$(mktemp -d "$root/.sign.XXXXXX")
trap 'rm -rf "$work"' 0
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
for input in "$root/bin/go" "$root/bin/gofmt" "$root"/pkg/tool/ohos_arm64/*; do
    [ -f "$input" ] || { printf 'Missing bootstrap executable: %s\n' "$input" >&2; exit 1; }
    if ! "$tool" sign -inFile "$input" -outFile "$work/signed" -selfSign 1 > "$work/sign.log" 2>&1; then
        cat "$work/sign.log" >&2
        exit 1
    fi
    chmod +x "$work/signed"
    mv -f "$work/signed" "$input"
    printf 'Signed %s\n' "$input"
done
"$root/bin/go" version
