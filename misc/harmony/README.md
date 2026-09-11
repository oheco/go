# HarmonyOS PC Go 工具链

源码来自 GitHub `golang/go`，工作分支为 `release-branch.go1.27`，起始提交
`2ee6421c51553e7164590445f96b123a858c1f4a`，版本为 Go 1.27.1。
目标为 `GOOS=ohos GOARCH=arm64`，`runtime.GOOS`、`GOHOSTOS` 同样为 `ohos`。
所有工具采用标准名称和目录：`bin/go`、`bin/gofmt`、
`pkg/tool/ohos_arm64/{compile,link,asm,cgo,...}`。
最终验证记录和剩余失败见 [VALIDATION.md](VALIDATION.md)。

## 已在鸿蒙宿主验证的功能

| 功能 | 验证内容 |
| --- | --- |
| 纯 Go 工具链 | 原生自举、build/run/test/install、子进程和签名缓存 |
| cgo、PIE | C pthread 回调 Go、并发调用和 GC |
| c-archive、c-shared | 原生 C 程序链接静态库、dlopen 动态库；16 个 C 线程回调 Go |
| plugin | 完整上游测试：加载、重复加载、跨插件 type/itab 一致性 |
| shared、linkshared | 构建共享 runtime/sync，运行链接共享 Go 库的 cgo 程序 |
| race | 正常/故意竞争程序；sync、sync/atomic、runtime/race 短测试 |
| ASan | 正常 C 内存访问、故意越界；错误报告定位 Go 源码行 |
| fuzz、coverage、pprof、trace | 4 个 fuzz worker；覆盖率、CPU profile、trace 生成和分析 |

本 port **不支持 `-msan`**，与其他未支持 MSan 的 Go 平台一样，使用该选项时返回：

```text
-msan is not supported on ohos/arm64
```

MSan 不列入本次移植的待办或完成条件，发行包不包含 MSan 运行库。
此前的调查仅作为历史记录保留，见 [compiler-rt/MSAN.md](compiler-rt/MSAN.md)。
N-API / ArkTS 示例属于可选应用集成，已验证库的编译签名；应用内运行不属于本次
Go 工具链移植的完成条件。

## 需要的工具

Linux 引导阶段需要可用的 Go、Git、Bash 和 GNU tar；打包另外使用 Python 3。
纯 Go 引导无需 CMake。
鸿蒙原生构建需要 `sh`、已签名的 ohos/arm64 引导 Go，以及 PATH 中的
`binary-sign-tool`。cgo 另外需要 OHOS SDK 的 `clang`、`clang++`、`ld.lld`、
`llvm-ar`、musl sysroot 和编译器运行库。默认 C/C++ 编译器为 Clang，归档器为 llvm-ar。

`go tool pprof`、`trace` 等工具遵循 Go 1.27 的按需构建机制；它们没有预先出现在
`pkg/tool` 中并不表示工具链缺失。源码随发行目录提供，首次使用会构建并签名。

## 构建

流程：Linux Go → 认识 ohos 的 Linux Go → 交叉构建 ohos/arm64 引导工具链
→ 鸿蒙签名 → 鸿蒙原生自举。纯 Go 引导无需再构建一套单独命名的 musl Go。

Linux 中：

```sh
cd /mnt/linux_share/ohos/go
GOROOT_BOOTSTRAP=/mnt/linux_share/ohos/go-ohos/.toolchains/go \
    sh misc/harmony/bootstrap.sh
```

默认引导输出 `../bootstrap/ohos/go`。若目录已经存在，使用 `BOOTSTRAP_OUTPUT`
指定新的绝对路径。构建使用当前源码工作树，包含未提交的适配。

通过真实终端连接宿主：

```sh
python3 /mnt/linux_share/ohos/zshd/zshc
```

宿主执行：

```sh
pwd
cd /storage/Users/currentUser/dev/ohos/go
# 当前 zsh 宿主应用已验证可用的私有目录。
mkdir -p /data/storage/el2/base/haps/entry/files/go-tmp
export TMPDIR=/data/storage/el2/base/haps/entry/files/go-tmp
sh misc/harmony/sign.sh /storage/Users/currentUser/dev/ohos/bootstrap/ohos/go
GOROOT_BOOTSTRAP=/storage/Users/currentUser/dev/ohos/bootstrap/ohos/go sh src/make.sh
export PATH=/storage/Users/currentUser/dev/ohos/go/bin:$PATH
go version
go env GOOS GOHOSTOS GOARCH CGO_ENABLED CC CXX GOTOOLDIR
```

`make.sh` 默认构建支持 cgo 的标准库，可设置 `CGO_ENABLED=0` 做纯 Go 自举。
引导 Go 必须是不同目录中的已签名 ohos/arm64 版本。默认 `GOMAXPROCS=4`。
源码和构建产物在共享目录互通：Linux `/mnt/linux_share` 对应宿主
`/storage/Users/currentUser/dev`；宿主执行时使用宿主路径。

## 回归与打包

```sh
# 在鸿蒙执行。使用私有 TMPDIR，源码和结果可放在共享目录。
sh misc/harmony/check.sh /storage/Users/currentUser/dev/ohos/go
TEST_OUTPUT=/storage/Users/currentUser/dev/ohos/bootstrap/full/regression \
    sh misc/harmony/check-full.sh /storage/Users/currentUser/dev/ohos/go
# 上游专项测试：all / paths / cgo / tools / std。每套日志和退出码独立保存。
TEST_OUTPUT=/storage/Users/currentUser/dev/ohos/bootstrap/full/upstream \
    sh misc/harmony/check-toolchain.sh /storage/Users/currentUser/dev/ohos/go all
```

Linux 中对完成自举、签名的目录打包：

```sh
sh misc/harmony/package.sh /mnt/linux_share/ohos/dist
```

输出 `go1.27.1.ohos-arm64.tar.gz` 和 SHA-256 文件，包内根目录为 `go/`。
打包会把共享目录呈现的 `0777` 权限恢复为标准的源码 `0644`、工具和目录 `0755`；
Git 记录的可执行测试文件保留执行权限。输出文件已存在时拒绝覆盖。
race 对象的来源和重建方法见
[compiler-rt/README.md](compiler-rt/README.md)。ArkTS 集成见
[testdata/napi/README.md](testdata/napi/README.md)。

## 平台实现

`ohos` 同时匹配 `linux`、`unix` 构建标签及 Linux 源文件后缀；`_ohos.go`、
`_ohos_arm64.go` 仅匹配鸿蒙。替换同包 Linux 实现时，需要在 Linux 文件加
`//go:build !ohos`。预编译的 `_linux.syso` 和 `_linux_arm64.syso` 不复用，
避免误用 ABI 或虚拟地址布局不兼容的二进制。`IsOhos=1`、`IsLinux=0`。

Go 链接器生成 OHOS musl 解释器路径；cgo 使用外部链接，以支持 TLSDESC 重定位。
Go shared 库使用 LLD 的 `--apply-dynamic-relocs` 写入后续链接所需的 GCData
地址；读取 GC 位图时只接受可加载区段，避免把 ELF 注释误当作指针位图。
此前工具链生成的 Go shared/linkshared 库及调用程序应重新构建，以纳入该修复。
Go 的 ARM64 TLS 使用 TLSDESC，允许 musl 动态加载 Go 库。共享库初始化提供
musl 未传入的空 argv/auxv，并从 libc 取得宿主环境变量、从 `/proc/self/auxv`
取得辅助向量。c-archive/c-shared 在包初始化前从 `/proc/self/cmdline` 恢复
宿主 `os.Args`，保留空参数、空格和长参数；procfs 不可访问时参数仍为空。

最终 build ID 写入后，`cmd/go` 对可执行文件、共享库和插件自动签名，成功后原子
替换，保留标准名称。签名失败返回错误。缓存包含签名策略标识，避免误用未签名
产物；`cmd/dist` 对直接生成的 go_bootstrap 同样签名。直接使用 `go tool link`
也会在鸿蒙宿主签名，可用 `-ohossign=false` 留给后续处理。外部工具修改或 strip
二进制后需要重新签名。签名器只从 PATH 查找。

`check-toolchain.sh` 使用 `cc-sign.sh` / `cxx-sign.sh` 为上游测试直接编译的 C
可执行文件和共享库签名；对象文件、预处理和编译器查询不签名。正常 cgo 默认
仍使用 `clang` / `clang++`，直接 Go 链接器也默认使用 Clang 和 llvm-ar。

其余适配包括 pprof 的 Linux RSS 单位和映射信息、fchmodat2 不可用时的回退、
链接器 fallocate 预分配不可用时的处理、设备文件 splice 的无损回退，以及 Unix
stream EOF 返回无效 sockaddr 时的处理。FIPS 故意篡改二进制的测试在运行前重签名，
确保测试的是 FIPS 校验失败。

## 当前宿主的限制

工具链 `cmd/...` 短测试已全部通过（117 个包），最新标准库测试仅剩
`syscall.TestPassFD` 失败。随后修复了并发 vfork 损坏 musl 信号链的问题，最终
工具链的 runtime 连续三轮短测试通过。详见 [VALIDATION.md](VALIDATION.md)。
共享目录不提供完整的 chmod、Unix socket、硬链接等
语义，临时文件应放在应用私有目录。即使使用私有目录，仍观察到：

- 硬链接被拒绝；`/`、`/tmp` 和部分 `/etc` 项目不允许读取。
- 宿主会限制已运行二进制和已加载库的内容与时间戳修改。工具链通过写入新文件
  再替换旧文件更新产物；`os.Chtimes` 对此类受保护文件仍可能返回权限错误。
- 内核对部分带末尾斜杠的文件路径采用了不同于 Linux 的行为。原生 C 的
  `openat` 同样能在此类路径创建文件。`os` 已兼容这些差异，保留符号链接目标
  和 `..` 的实际解析顺序；完整 `os` / `path/filepath` 测试已通过。
- 应用沙箱对 namespace/mount 系统调用发送 SIGSYS；相关已确认的特权测试
  明确跳过。强制 clone3 测试也被沙箱终止，普通进程创建和 pidfd 测试通过。pidfd 查找额外检查句柄是否已退出，
  处理宿主在 Wait 后仍能短暂打开已回收进程句柄的时序差异。
  没有通过伪造成功来实现这些操作。
- Unix socket 的跨进程文件描述符传递在当前沙箱中丢失辅助数据，接收端返回
  `MSG_CTRUNC`。使用 128 字节控制缓冲区的原生 C 程序复现了相同结果，
  `syscall.TestPassFD` 保留失败以记录此限制。
- Unix datagram 发送时自动绑定地址；测试比较实际发送端地址。其余已修复的
  Unix socket、writev、splice 用例均保留行为断言。
- OHOS SDK 不提供 `pthread_cancel`；对应 cgo 测试明确跳过。SDK 的 ASan 可用，
  但未提供 LSan；泄漏检测通过原生能力探测后跳过。
- 应用 UID 可能没有 passwd 条目；`os/user.Current()` 的 cgo 实现此时使用
  `$USER` / `$HOME` 回退，与纯 Go 一致，数字 UID/GID 仍来自内核。
- 当前 race 对象针对普通可执行文件验证，39 位布局将 Go 堆限制在低 32 GiB。
  race 动态库实测无法映射高地址全局数据所需的 shadow；`-race` 与 c-archive、
  c-shared、shared、plugin、linkshared 的组合在构建时明确拒绝。
  `-race -buildmode=pie` 也按 Go 的构建模式规则拒绝。

异步抢占信号与 exec 的协调已补齐，1000 次并发 exec 压力用例通过，
无需关闭异步抢占。并发 vfork 使用每个运行线程独立的状态，防止子进程重置信号时
误改父进程的 musl 信号链；新增回归验证 1600 次并发启动和后续故障恢复。
Go 保留宿主预装的 DFX/C 信号处理器，C 代码中的同步崩溃
按 Go 的转发规则交给原处理器。回溯测试在受控子进程中恢复默认 SIGSEGV
处理器，C 信号转发测试另行保留。标准库的最终逐包结果见 [VALIDATION.md](VALIDATION.md)。
MSan 和 ArkTS 应用集成不列入移植待办。不要把当前结果当作完整
上游标准库全通过或所有构建模式组合均已支持的证明。
