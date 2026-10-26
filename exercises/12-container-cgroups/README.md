# 用 Go 解读 cgroup v2 资源限制

完成 `ParseCPU`、`CPUCount` 与 `ParseMemory`，保留已提供的读取函数和 `cmd/inspect`。`cpu.max` 的格式是 `quota period` 或 `max period`，两个数字的单位都是微秒；`memory.max` 是字节数或 `max`。

## 离线验收

`go test ./...` 不需要 Linux、Docker、网络或特权。测试覆盖有限值、无本级上限、空白、非法数字、零周期、溢出、真实 fixture 文件读取与缺失文件错误。有限 CPU 配额必须为正数；内存 0 是有效上限；`max` 与 0 不能混淆。

## Linux/Docker 实操验收

在 Linux 或 WSL2 的练习目录执行，要求 Docker 使用 cgroup v2。先通过 `docker info --format '{{.CgroupVersion}}'` 确认输出为 `2`；v1 环境不能直接套用本题文件格式。

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o inspect ./cmd/inspect
docker run --rm --platform linux/amd64 --cpus=0.5 --memory=64m --mount "type=bind,src=$(pwd),dst=/work,readonly" -w /work alpine:3.22 ./inspect
docker run --rm --platform linux/amd64 --mount "type=bind,src=$(pwd),dst=/work,readonly" -w /work alpine:3.22 ./inspect
```

将集成测试预先编译为 Linux 二进制，再放进同一受限容器验证实际限制。这样 64 MiB 限制不需要容纳 Go 编译器：

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -c -tags=integration -o acceptance.test
docker run --rm --platform linux/amd64 --cpus=0.5 --memory=64m -e GOPATH_EXPECT_CPU=0.5 -e GOPATH_EXPECT_MEMORY=67108864 --mount "type=bind,src=$(pwd),dst=/work,readonly" -w /work alpine:3.22 ./acceptance.test -test.run TestIntegration -test.v
```

集成测试读取当前 Linux 环境的 `/sys/fs/cgroup`，要求 cgroup v2，并将 CPU 配额比值、内存字节数与显式期望值比较。缺少 Linux、文件或期望值会失败，不会 Skip。可用 `GOPATH_CGROUP_DIR` 指定已解析的真实 cgroup 目录。上面的预编译测试用于观察 64 MiB 容器，不会写 GoPath 进度。

要通过 CLI 记录实操成绩，在 Linux/WSL2 中将同版本的 Linux GoPath 二进制和已解题的学习目录挂入 Go 容器。为编译器保留 1 GiB 内存，依然验证真实且有界的 cgroup。先在项目根目录构建 `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/gopath-linux-amd64 ./cmd/gopath`，再把以下两个 `/abs/…` 替换为各自的真实绝对路径：

```sh
docker run --rm --platform linux/amd64 --cgroupns=private --cpus=0.5 --memory=1g \
  -e GOPATH_EXPECT_CPU=0.5 -e GOPATH_EXPECT_MEMORY=1073741824 \
  --mount "type=bind,src=/abs/gopath-workspace,dst=/work" \
  --mount "type=bind,src=/abs/build/gopath-linux-amd64,dst=/usr/local/bin/gopath,readonly" \
  -w /work golang:1.27-alpine \
  gopath --workspace /work check 12-container-cgroups --integration
```

学习目录必须可写，才能保存成绩；Go 工具链在容器内。首次编译可能较慢。此命令验证的是该容器的 0.5 CPU / 1 GiB 限制；64 MiB 实验另行保留观察记录。不要用无约束宿主机或 fixture 冒充受限容器。

非 amd64 主机可改为匹配宿主机的 GOARCH 并移除 `--platform`。第一条应观察到配额比值约 0.5、内存上限 67108864 字节。周期不要求固定为某个值；验证比值。第二条可能显示 `max`，但 Docker VM 或父 cgroup 仍可能有限制。本工具读取容器私有 cgroup 挂载根 `/sys/fs/cgroup`，不尝试解析任意宿主机进程的 cgroup 路径；可向 `go run ./cmd/inspect /path/to/cgroup` 传入已解析的目录。

把环境、读取值及解释保存在 `observations.md`。本练习不要求触发 OOM、不更改宿主机 cgroup、不运行压力测试。

## 口述问题

1. `50000 100000` 代表什么？为什么不是“固定绑定半个 CPU 核”？
2. CPU throttling 与 CPU affinity 有何区别？
3. `memory.max=max` 为什么不能证明机器内存无限？
4. Kubernetes 的 requests、limits 分别影响调度还是运行时？

参考：[Linux cgroup v2 文档](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)、[Docker 资源限制](https://docs.docker.com/engine/containers/resource_constraints/)。完整验收需要离线测试、受限容器的真实集成检查和观测说明。
