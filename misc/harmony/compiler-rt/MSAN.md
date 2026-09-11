# MSan: unsupported; archived investigation

This port does not support MSan. The Go command rejects `-msan` for
`ohos/arm64` with `-msan is not supported on ohos/arm64`, using the same
mechanism as other unsupported Go platforms. MSan is outside the agreed port
scope and is not a pending task or release requirement. No MSan runtime is
included in the distribution.

The following records the earlier experiment for reference. Its scripts and
patches are not used by the Go bootstrap, functional checks or installed tools.
The SDK had no MSan runtime, and the experimental runtime did not pass the
native C control test.

The experiment uses `https://github.com/openharmony/third_party_llvm-project`
at commit `9b4f67cf99587aedb3c6ab31a2ef49358cd8d5e3`, which retains ARM64
39-bit shadow/origin mappings. Linux CMake and Clang can build the archives:

```sh
sh misc/harmony/compiler-rt/build-msan-experimental.sh \
    /path/to/openharmony-llvm /path/to/ohos/native /absolute/msan-build
```

The patch excludes the unsupported SysV shared-memory interceptor and handles
the absent glibc dynamic-TLS tracker on OHOS musl. `OHOS=ON` includes the
platform stack unwinder. `-fno-emulated-tls` prevents recursive allocation in
`__emutls_get_address`. The archives stay in the build directory; the script
does not change the SDK or enable any Go feature.

The native control program is `msan-native.c`. Compile it on OHOS using the
built archive, sign it with PATH's `binary-sign-tool`, then run it once with no
arguments (zeroed memory) and once with an argument (uninitialized memory):

```sh
clang -g -O1 -fsanitize=memory -fno-emulated-tls -fPIE -pie \
    -fno-sanitize-link-runtime msan-native.c \
    -Wl,--whole-archive /path/to/libclang_rt.msan-aarch64.a \
    -Wl,--no-whole-archive -lpthread -ldl -lm -o msan-native
binary-sign-tool sign -inFile msan-native -outFile msan-native.new -selfSign 1
chmod +x msan-native.new
mv msan-native.new msan-native
MSAN_OPTIONS=verbosity=1 ./msan-native
MSAN_OPTIONS=verbosity=1 ./msan-native unsafe
```

On the tested host, initialization succeeds, but **both** cases exit 134 before
`main`: `__interceptor_fopen` reports an uninitialized filename terminator in
`/system/lib64/libhilog_inner.so`, loaded by the musl startup code. The report
is outside the test program. This does not demonstrate a working detector,
nor establish a bug in that system library: writes in uninstrumented
dependencies can leave stale shadow memory.

The investigation stopped at this point. Compatibility of system dependencies
and interceptors was not established, and Go heap-layout changes or Go MSan
tests were not enabled. These are historical findings, not remaining work for
this port.
