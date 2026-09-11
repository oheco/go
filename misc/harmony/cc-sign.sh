#!/bin/sh
# Compiler wrapper for native tests that build C executables outside cmd/go.
# Object files and compiler queries pass through unchanged.
set -eu
output=a.out
link=yes
take_output=no
for arg do
    if [ "$take_output" = yes ]; then
        output=$arg
        take_output=no
        continue
    fi
    case "$arg" in
        -o) take_output=yes ;;
        -o?*) output=${arg#-o} ;;
        -c|-S|-E|-fsyntax-only|--version|-v|-print*|-dump*|'-###') link=no ;;
    esac
done
"${OHOS_TEST_CC:-clang}" "$@"
[ "$link" = yes ] && [ -f "$output" ] || exit 0
if ! llvm-readelf -h "$output" 2>/dev/null | grep -Eq 'Type:[[:space:]]+(EXEC|DYN)'; then
    exit 0
fi
signer=$(command -v binary-sign-tool) || {
    printf 'binary-sign-tool not found; check the LLVM tool directory in PATH.\n' >&2
    exit 1
}
signed=$output.signed.$$
log=$output.sign.$$.log
trap 'rm -f "$signed" "$log"' 0
if ! "$signer" sign -inFile "$output" -outFile "$signed" -selfSign 1 > "$log" 2>&1; then
    cat "$log" >&2
    exit 1
fi
chmod +x "$signed"
mv "$signed" "$output"
