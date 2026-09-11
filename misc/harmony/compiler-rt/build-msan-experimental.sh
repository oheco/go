#!/bin/sh
# Historical reproducer. MSan is outside this port's supported features.
# This does not enable Go -msan and is not part of the toolchain build.
# Run on Linux with CMake and a Clang capable of targeting aarch64-linux-ohos.
set -eu
llvm=${1:?Usage: sh build-msan-experimental.sh /path/to/openharmony-llvm /path/to/native-sdk /absolute/build-directory}
sdk=${2:?Missing native SDK directory}
build=${3:?Missing build directory}
case "$build" in /*) ;; *) printf 'Build directory must be absolute.\n' >&2; exit 1 ;; esac
here=$(CDPATH= cd "$(dirname "$0")" && pwd -P)
llvm=$(CDPATH= cd "$llvm" && pwd -P)
sdk=$(CDPATH= cd "$sdk" && pwd -P)
pinned=9b4f67cf99587aedb3c6ab31a2ef49358cd8d5e3
[ "$(git -C "$llvm" rev-parse HEAD)" = "$pinned" ] || {
    printf 'Expected OpenHarmony LLVM commit %s.\n' "$pinned" >&2
    exit 1
}
patch=$here/msan-experimental.patch
if git -C "$llvm" apply --check "$patch" 2>/dev/null; then
    git -C "$llvm" apply "$patch"
elif ! git -C "$llvm" apply --reverse --check "$patch" 2>/dev/null; then
    printf 'LLVM source does not match the experimental MSan patch.\n' >&2
    exit 1
fi
cxxinclude=$sdk/llvm/include/libcxx-ohos/include/c++/v1
[ -f "$cxxinclude/__config_site" ] || {
    printf 'Missing OHOS libc++ configuration in %s.\n' "$cxxinclude" >&2
    exit 1
}
# Linux selects the existing MSan CMake target; OHOS selects the native
# symbolizer sources. The actual compiler target and sysroot are OHOS.
# Native TLS is essential: emulated TLS recursively calls intercepted calloc.
cmake -S "$llvm/compiler-rt" -B "$build" \
    -DCMAKE_BUILD_TYPE=Release -DCMAKE_SYSTEM_NAME=Linux \
    -DCMAKE_SYSTEM_PROCESSOR=aarch64 -DOHOS=ON \
    -DCMAKE_C_COMPILER="${CC:-clang}" -DCMAKE_CXX_COMPILER="${CXX:-clang++}" \
    -DCMAKE_C_COMPILER_TARGET=aarch64-linux-ohos \
    -DCMAKE_CXX_COMPILER_TARGET=aarch64-linux-ohos \
    -DCMAKE_SYSROOT="$sdk/sysroot" -DCMAKE_TRY_COMPILE_TARGET_TYPE=STATIC_LIBRARY \
    -DCMAKE_CXX_FLAGS="-fno-emulated-tls -nostdinc++ -isystem \"$cxxinclude\"" \
    -DCOMPILER_RT_DEFAULT_TARGET_ONLY=ON \
    -DCOMPILER_RT_BUILD_BUILTINS=OFF -DCOMPILER_RT_BUILD_CRT=OFF \
    -DCOMPILER_RT_BUILD_SANITIZERS=ON -DCOMPILER_RT_SANITIZERS_TO_BUILD=msan \
    -DCOMPILER_RT_BUILD_XRAY=OFF -DCOMPILER_RT_BUILD_LIBFUZZER=OFF \
    -DCOMPILER_RT_BUILD_PROFILE=OFF -DCOMPILER_RT_BUILD_MEMPROF=OFF \
    -DCOMPILER_RT_BUILD_ORC=OFF -DCOMPILER_RT_BUILD_GWP_ASAN=OFF \
    -DCOMPILER_RT_INCLUDE_TESTS=OFF -DCOMPILER_RT_USE_BUILTINS_LIBRARY=OFF \
    -DLLVM_CMAKE_DIR="$llvm/llvm/cmake/modules" -DLLVM_MAIN_SRC_DIR="$llvm/llvm"
cmake --build "$build" --target clang_rt.msan-aarch64 clang_rt.msan_cxx-aarch64 -j "${JOBS:-4}"
printf 'Experimental archives: %s/lib/linux/\nNot installed; see MSAN.md for the failing native control test.\n' "$build"
