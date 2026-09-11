#!/bin/sh
# Package an already built/signed OHOS GOROOT on Linux.
set -eu
root=$(CDPATH= cd "$(dirname "$0")/../.." && pwd -P)
exec python3 "$root/misc/harmony/package.py" "$root" "${1:?Usage: sh package.sh /absolute/output/directory}"
