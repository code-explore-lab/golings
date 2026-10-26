# 验证记录

日期：2026-10-25。环境：Windows amd64、Go 1.27.0、Helm 4.3.0、Docker CLI 29.8.1。CLI 核心仅依赖标准库，go.mod 声明 Go 1.23；本次未单独使用 Go 1.23 运行测试。Linux 产物为交叉编译，未在 Linux 主机上运行。

## 自动检查

| 检查 | 结果与范围 |
|---|---|
| gofmt | 全部 .go 和 .go.txt 均无待格式化文件 |
| `go test -count=1 -timeout=3m ./...` | 通过：题库格式、catalog、CLI、progress、runner、workspace；命令入口可编译 |
| `go vet ./...` | 通过 |
| `go test -race ./cmd/... ./internal/...` | 通过：核心并发及 CLI/执行器回归 |
| 全量课程作者检查 | 20/20 通过，205.684 秒；起始代码编译及预期失败、参考解答通过、07–15 实操测试编译全部通过 |
| 并发练习参考解答 race | 04、07、08、09、17、20 通过；Kubernetes 依赖编译使用 GOMAXPROCS=2、-p=2 |
| Windows 构建 | 通过：build/gopath.exe |
| Linux amd64 构建 | 通过：GOOS=linux、GOARCH=amd64、CGO_ENABLED=0，产物 build/gopath-linux-amd64 |

全量作者检查命令为：

    go test -tags=curriculum -parallel=2 -timeout=30m -v ./exercises

该检查在临时工作区逐题验证：起始代码能够编译，但行为测试确实失败；覆盖为独立参考解答后，同一组测试通过；07–15 的 integration 测试仅编译，不执行。它是维护者检查，不属于学习者的 check，不写学习者进度。首次 Kubernetes 竞态依赖并行编译占用过高，已停止相应进程，并以受限并发顺序重跑通过。

核心行为检查覆盖：初始化拒绝覆盖；答案不进入嵌入题库或学习工作区；提示不完成；零测试、跳过、缺少实操测试不通过；外部 GOFLAGS/保存的 Go 环境配置不能过滤必需测试；测试修改或缺失被拒绝；失败清除旧通过记录；文件变化显示成绩失效；状态损坏明确报错；重置需确认并保留额外文件。

另有真实 Go 子进程回归：测试启动后改变源码、必需测试或 go.mod 均拒绝记分，生成 go.sum 可以通过且纳入最终指纹；同一工作区的 check/hint/reset 互斥，忙时不修改状态，失败后释放锁，遗留锁不会自动删除。

## CLI 手动流程

使用最终 Windows 二进制及全新临时学习目录，实际执行：

1. 初始化 20 题，next 指向 01，hint 增加提示次数但完成数仍为 0。
2. 原始 01 起始代码因错误包装测试失败，命令退出 1，未遗留写入锁。
3. 只在临时学习目录将错误包装从 %v 改为 %w，check 通过。
4. 在题目子目录以新进程运行 stats/next，读取到 1/20 和下一题 02，证明祖先目录发现与持久化可用。
5. 空输入取消 reset，解答保持；输入精确的 RESET 01-go-errors 后恢复起始代码。
6. reset 保留额外 notes.md，完成及提示记录清零，写入锁正常释放；学习目录没有 solution 目录。

## 离线部署资产检查

Helm 实际工具已通过 `helm lint`、`helm template`、`helm package`，产出可渲染的 Deployment/Service 及 gopath-demo-0.1.0.tgz。没有执行集群安装、升级或回滚。

Grafana 课程的 `docker compose config --quiet` 通过，验证 Compose 配置可解析；没有启动容器或访问 Grafana/Prometheus。

## 尚未执行的真实环境验收

Docker CLI 报告无法连接 dockerDesktopLinuxEngine，kubectl 报告 current-context 未设置，因此未执行 07–15 的真实环境检查：Minikube List/Watch、CRD/Controller/Operator、Helm 安装升级回滚、Linux namespace/cgroup、工作负载就绪和 Service、Prometheus 抓取、Grafana 看板与查询。

尝试 kubectl 的 client dry-run 也因缺少 API discovery 环境而失败；不把它记为 YAML 验证通过。没有自动启动集群，也没有创建、修改或删除真实集群资源。

这些题的本地单测成功只会成为 [~]，必须按各题 README 准备环境后显式运行 `check --integration` 才能完成。08/09 需先显式安装 CRD；11/12 需 Linux/WSL2。真实验收结果、部署后的运行兼容性及人工口述仍须另行复核。
