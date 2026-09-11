# OHOS Go race runtime

Source: LLVM `51bfeff0e4b0757ff773da6882f4d538996c9b04`, the revision used by
Go's `runtime/race/README`. `race-ohos.patch` includes Go's ARM64
`Trace.local_head` fix and the OHOS 39-bit address mapping.

The object is named `runtime/race/race_ohos_arm64.syso`. OHOS source selection
continues to include Linux source, but excludes `_linux.syso` and
`_linux_arm64.syso`: precompiled Linux objects do not imply OHOS ABI compatibility.

For a native rebuild with Clang in PATH:

```sh
sh misc/harmony/compiler-rt/build-race.sh /path/to/llvm-project
```

For a Linux cross build, set `EXTRA_CFLAGS` to
`--target=aarch64-linux-ohos --sysroot=/path/to/ohos/native/sysroot` and `CC=clang`.
The script verifies the source revision and applies the patch idempotently.
The resulting object contains no executable signature; the final Go program
is signed normally by `go build`.

On OHOS, compile the patched `compiler-rt/lib/tsan/go/test.c` with
`clang -DGO_OHOS=1 -no-pie test.c race_ohos_arm64.syso -lpthread -o test`,
sign the executable using PATH's `binary-sign-tool`, and execute with
`GORACE='exitcode=0 atexit_sleep_ms=0'`. It intentionally reports races.
Also run `go test -race -short sync sync/atomic runtime/race` and the detector
positive/negative examples in `misc/harmony/testdata/detectors`.

39-bit layout: Go data/heap in the low 32 GiB (heap starts at 4 GiB), shadow at
64–128 GiB, metadata at 128–136 GiB. This mapping is for ordinary executables.
The patched LLVM C test passes and reports its intentional race. A c-shared
library fails to allocate shadow for musl's high-address Go data, so the Go
command rejects race with c-archive, c-shared, shared, plugin and linkshared.
Race with PIE is also rejected by Go's existing build-mode rule.
C allocations are not instrumented by Go's TSan runtime.

MSan is outside this port's supported features. The earlier investigation is
archived in [MSAN.md](MSAN.md) and is not a pending port task.
