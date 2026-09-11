# Go 1.27.1 OHOS ARM64 验证记录

验证日期：2026-09-12。源码工作分支 `release-branch.go1.27`，上游基点
`2ee6421c51553e7164590445f96b123a858c1f4a`。适配保留在当前工作树中。

## 环境与工具链

工具由鸿蒙原生 `src/make.sh` 自举生成，默认环境为：

```text
go version go1.27.1 ohos/arm64
GOOS=GOHOSTOS=ohos
GOARCH=GOHOSTARCH=arm64
CGO_ENABLED=1
CC=clang
CXX=clang++
GOTOOLDIR=$GOROOT/pkg/tool/ohos_arm64
```

宿主为 HUAWEI MateBook Pro S，系统版本 `7.0.0.105(SP10C00E101R7P3)`，
使用 LLVM 15.0.4、OHOS musl 和 PATH 中的 `binary-sign-tool`。
测试临时文件放在应用私有目录
`/data/storage/el2/base/haps/entry/files/go-tmp`，`GOMAXPROCS=4`。

## 本次修复与验证

| 范围 | 结果与证据 |
| --- | --- |
| 路径语义 | 修复创建、删除、重命名的末尾斜杠行为，保留符号链接和 `..` 的实际解析顺序；完整 `os`、`path/filepath` 通过 |
| 进程和用户 | cgo `os/user` 在 passwd 缺少应用 UID 时使用环境回退；pidfd 在打开后检查已回收进程；pidfd 专项重复 100 次通过 |
| exec 信号协调 | 保留异步抢占，补齐 exec 锁、待处理信号计数及 EINTR 重试；cgo `Test18146` 的 1000 次 exec 通过 |
| cgo | 完整 cgo、c-archive、c-shared、plugin 上游专项通过；库调用验证 16 个 C 线程回调、GC、环境继承和宿主参数 |
| 库参数 | 包初始化前恢复 `os.Args`；测试包含空参数、空格、中文及 5000 字节参数；procfs 无法访问时参数为空 |
| shared GC | 修复 LLD 的 GCData 重定位和读取区段检查；完整 shared 测试中其余用例通过后，原失败 `TestGCData` 修复并复测通过；最终自举版再次通过（18.173s）；旧 shared/linkshared 库和程序需要重建 |
| 直接链接与签名 | `go tool link` 默认 Clang 并自动签名；cmd/go 在写完 build ID 后签名；编译器 testdir、pack、链接器测试通过 |
| 诊断工具 | coverage、fuzz、CPU profile、trace、pprof 功能回归通过；正常与故意出错的 race、ASan 用例通过 |
| 原生自举 | `native-bootstrap-fork-fixed.log` 记录包含最终 vfork 修复的自举成功；`release-fork-fixed.log` 中 check.sh 和 check-full.sh 全部通过 |

`check-toolchain.sh` 使用 `cc-sign.sh` / `cxx-sign.sh` 为上游测试直接生成的 C
可执行文件与共享库签名；一般用户构建的默认 CC/CXX 仍为 clang/clang++。
Linux 侧路径、os/user、os/exec、pidfd 回归也通过。

## 最新全包检查

完成路径、cgo 和链接器修复后，关闭测试缓存，分别执行以下全包检查
（这轮检查先于最后的并发 vfork 修复）：

```sh
go test -v -short -count=1 -p=2 -timeout=10m cmd/...
go test -short -count=1 -p=2 -timeout=10m std
```

| 检查 | 通过的包 | 无测试的包 | 失败的包 | 退出码 |
| --- | ---: | ---: | ---: | ---: |
| cmd/... | 117 | 115 | 0 | 0 |
| std | 261 | 119 | 1 | 1 |

标准库剩余失败为 `syscall.TestPassFD`。当前应用沙箱的 SCM_RIGHTS 跨进程传递
丢失控制数据。原生 C 同样复现：接收 1 字节普通数据、0 字节控制数据，
控制缓冲区为 128 字节，返回 `MSG_CTRUNC`。该失败保留，未以跳过掩盖。

上述 cmd/... 使用上游短测试模式；会跳过部分耗时的 cgo/共享库测试，不能替代
前述完整专项结果。标准库本轮的 runtime 通过（126.422s），这轮检查后又定位
并修复了下面的并发 vfork 问题。

## 并发 vfork 信号链问题

多轮 runtime 测试曾在 `TestPanicInlined`、`TestSetPanicOnFault` 等可恢复故障
中直接 SIGSEGV 退出。独立 overlay 诊断捕捉到：故障前 musl 保存的 SIGSEGV
用户处理器从 Go 函数地址变成了 SIG_DFL，信号没有进入 Go 的处理函数。

根因是多个 vfork 子进程共享父进程地址空间，却使用同一个全局
`inForkedChild` 布尔值。一个子进程清除此值时，另一子进程可能尚在重置信号；
后者误入 libc `sigaction`，改写了父进程共享的 musl 信号链。
OHOS 现改用每个 M（运行线程）独立的 fork 状态，子进程只重置自己的内核信号
处理器，不改写 libc 的共享状态。正常 C 信号处理与 sanitizer 拦截保持不变。

16 个并发 worker、共 1600 次启动的复现程序在旧实现中检测到处理器被清空；
修复后处理器保持不变，随后的 nil 故障可正确恢复。
新增回归 `TestCgoOhosConcurrentFork` 也通过，Linux 侧信号/故障回归通过。
最终原生自举后，`go test -v -short -count=3 -timeout=10m runtime` 连续三轮全部
通过（200.525s），每轮均包含 1600 次并发 fork、故障恢复和 C 信号转发测试。
最初与自举同时运行的 runtime 检查曾因编译器被替换而中断，不能计作有效全套结果。
诊断 overlay 不在发行源码和工具中。

## 明确限制

- `-msan` 按项目约定不支持，返回 `-msan is not supported on ohos/arm64`；
  不列入移植待办，发行包不包含 MSan 运行库。
- 普通 race 可执行程序可用；race 与 PIE、动态库和 linkshared 的组合不支持。
- SDK 不提供 `pthread_cancel`；LSan 能力探测显示 `detect_leaks` 不支持。
  对应测试按实际能力跳过，ASan 正常执行。
- 硬链接、读取部分系统目录及 namespace/mount/强制 clone3 被宿主权限或沙箱
  限制。相关已确认的特权测试明确跳过，普通进程创建和 pidfd 测试保留。
- 已运行或已加载的二进制 inode 可能拒绝修改和 Chtimes；构建通过写入新文件后
  替换完成更新。shared 重建测试对同内容复制品更新时间戳，保留原有重建断言。
- 仅验证 ohos/arm64 和当前宿主。N-API / ArkTS 应用集成不作为 Go 工具链移植
  的完成条件；官方 Go 自动工具链下载服务没有 OHOS 发行包，默认使用本地工具链。

## 原始证据

以下路径相对于 Go 仓库：`../bootstrap/full/`。

- 自举和功能回归：`native-bootstrap-sep12.log`、`release-regression-sep12.log`。
- 完整 cgo/路径：`cgo-path-final1.log`、`shared-fixed.log`、
  `shared-gcdata-fixed.log`、`sanitizers-fixed.log`、`pidfd-fixed.log`。
- 全包检查：`tools-release/tools.log`、`tools-release/status.tsv`、
  `std-release-sep12.log`。
- runtime 稳定性：`runtime-final-verbose.log`、`runtime-release-repeat.log`、
  `runtime-release-repeat2.log`、`setpaniconfault-stress.log`、`nil-thread-repro.log`。
- vfork 根因与修复：`runtime-signal-overlay-run4.log`、`concurrent-fork-baseline.log`、
  `concurrent-fork-fixed2.log`、`runtime-fork-linux.log`、`runtime-final-fork-fixed.log`。
- 最终工具链：`native-bootstrap-fork-fixed.log`、`release-fork-fixed.log`、
  `shared-gcdata-release.log`、`paths-final-fork-fixed.log`。
- 平台限制：`fd-pass.log`；Linux 回归：`path-linux-regression.log`、
  `user-exec-linux.log`、`pidfd-linux.log`。
