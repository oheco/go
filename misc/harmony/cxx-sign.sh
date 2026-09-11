#!/bin/sh
set -eu
export OHOS_TEST_CC=clang++
exec sh "$(dirname "$0")/cc-sign.sh" "$@"
