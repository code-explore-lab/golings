# 用 Go 观察 Linux namespace 隔离

完成 `exercise.go` 的 `ParseNamespaceLink` 和 `SameNamespace`。`ReadNamespaces` 与 `cmd/inspect` 已提供真实的 procfs 读取路径；不要把 fixture 当成实机实验结果。

目标：识别 `pid:[4026531836]` 中的 namespace 类型与 inode；拒绝损坏输入；比较两个 namespace 引用。一个进程在多种 namespace 中有不同的成员身份。相同 PID 数字也不意味着相同进程视图。

## 离线验收

运行 `go test ./...`。测试不访问 `/proc`，Windows/Linux 均可运行。合法 inode 必须为非零 uint64；允许输入两端空白；拒绝格式不完整、负数、溢出和尾随垃圾。`SameNamespace` 同时比较类型与 inode。

## Linux 实操验收

以下命令在 Linux 或 WSL2 的练习目录执行，要求 Go 与 Docker 可用；Windows 上的 fixture 测试不等于 Linux 验收。

```sh
go run ./cmd/inspect
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o inspect ./cmd/inspect
docker run --rm --platform linux/amd64 --mount "type=bind,src=$(pwd),dst=/work,readonly" -w /work alpine:3.22 ./inspect
docker run --rm --platform linux/amd64 --pid=host --mount "type=bind,src=$(pwd),dst=/work,readonly" -w /work alpine:3.22 ./inspect
```

真实 procfs 自动验收在 Linux/WSL2 的练习目录运行 `go test -tags=integration -run TestIntegration -v ./...`，或从同一个 Linux/WSL2 GoPath 工作区运行 `gopath check 11-container-namespaces --integration`。它读取 `/proc/self/ns` 与 `/proc/<当前 PID>/ns`，要求 pid、net、uts、mnt 的真实链接可解析且身份一致。Windows、缺失 procfs 或权限不足会明确失败，不会跳过。该测试证明当前进程的实际 namespace 观察；宿主机与容器、`--pid=host` 之间的比较仍须完成上述实验并记录。

非 amd64 主机可改为匹配宿主机的 GOARCH 并移除 `--platform`。比较 `pid`、`net`、`uts` 的输出；`--pid=host` 只改变 PID namespace 的共享方式，不能据此推断其他 namespace 都共享。Docker Desktop 的“host”是其 Linux VM，不能直接与 Windows 或 WSL2 发行版的 inode 比较；应在同一个 Linux Docker 宿主机上比较。不同 Linux 内核的 inode 数字没有比较意义。

把命令、运行环境和观测结果保存为本地 `observations.md`。无需 root、`--privileged` 或更改宿主机 namespace。缺少 `/proc` 时工具应报告错误，不能伪造观察值。

## 口述问题

1. PID namespace 为什么能让容器里的应用看到 PID 1？PID 1 的信号处理有什么特殊之处？
2. namespace 能限制内存吗？为什么还需要 cgroup？
3. 容器与虚拟机是否共享同一种内核边界？
4. 为什么比较 namespace 需要“类型 + inode”，而不是只比较 inode？

参考：[Linux proc_pid_ns(5)](https://man7.org/linux/man-pages/man5/proc_pid_ns.5.html)。完整验收包含离线解析测试、Linux procfs 集成检查、容器比较实验与口述。
