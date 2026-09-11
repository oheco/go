# ArkTS / Go N-API example

This is optional application integration, outside the Go toolchain port's
completion criteria. Go compilation, runtime and cgo do not require ArkTS.

Build on the OHOS host (the LLVM native SDK and binary-sign-tool must be in PATH):

```sh
CGO_ENABLED=1 CC=clang go build -buildmode=c-shared -o libgoharmony.so .
```

Place `libgoharmony.so` in the test application's `entry/libs/arm64-v8a/`.
Copy `types/libgoharmony/` to `entry/src/main/cpp/types/libgoharmony/` and add
`"libgoharmony.so": "file:./src/main/cpp/types/libgoharmony"` to the entry module's
dependencies. `Example.ets` demonstrates the import, synchronous call and Promise.

`add` validates two signed 32-bit integers. `sumAsync` accepts 0–1,000,000,
runs Go allocation and GC on a native worker, then resolves a Promise on the
Ark thread. No `napi_value` or Go pointer crosses the worker boundary. Do not
unload the Go shared library during the lifetime of the process.

The library has been compiled and signed with the OHOS toolchain. Actual ArkTS
application execution still needs a configured Ability test application and
its build/signing tools. The standalone `caller/main.c` is an embedding test:
it requires the process to have registered the Ark runtime creation callbacks.
On the current zsh host, `napi_create_ark_runtime` returns `napi_invalid_arg`
because no application framework registered those callbacks. This is not a
successful ArkTS runtime test; do not use it as release evidence.
