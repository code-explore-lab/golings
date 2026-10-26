# GoPath

GoPath 是 Rustlings 风格的 Go 工程学习 CLI：修改有缺陷的 Go 代码，阅读测试失败，逐条获取提示，通过检查后记录进度。首版有 20 道练习，优先训练 client-go、CRD、Controller、Operator 和 Helm，其次是容器底层与可观测工程。调度只安排一题概念练习。

CLI 核心仅使用标准库，不含 Web 应用或数据库。课程依赖放在各自的 Go module 内；参考答案不嵌入二进制、不安装到学习目录，也不会由普通检查命令执行。

第一次接触本项目，先读 [新手使用教程](docs/getting-started.md)：从环境准备到亲手完成第一题，再进入 Kubernetes 实操；包含预期输出、进度标记、重置和常见问题。

## 快速开始（Windows PowerShell）

完整课程使用 Go 1.27 或更高版本；本项目已在 Go 1.27.0 / Windows amd64 验证。CLI 核心的语言版本下限为 Go 1.23。Kubernetes、Prometheus 练习首次运行需要下载各自固定版本的依赖。

在项目根目录执行：

    go build -o .\build\gopath.exe .\cmd\gopath
    .\build\gopath.exe init .\gopath-workspace
    .\build\gopath.exe list
    .\build\gopath.exe next
    .\build\gopath.exe check
    .\build\gopath.exe hint
    .\build\gopath.exe stats

第一次 `check` 失败是预期行为：请编辑 `next` 输出目录中的 `exercise.go`，不要编辑课程源码目录里的模板。

在 Linux/macOS 上构建：

    go build -o ./build/gopath ./cmd/gopath
    ./build/gopath init
    ./build/gopath next

也可用 `go install ./cmd/gopath` 安装到 Go 的 bin 目录，加入 PATH 后直接运行 `gopath`。CLI 支持 Windows/Linux；namespace、cgroup 实操需要 Linux 或 WSL2，minikube 实操需要正在运行的容器运行时和集群。

## 工作区与命令

工作区按以下顺序选择：`--workspace <目录>`、`GOPATH_HOME`、当前目录或祖先目录中带 `.gopath` 的工作区，否则使用当前目录下的 `gopath-workspace`。不会修改标准 Go 环境变量 GOPATH。

    gopath init ./my-learning
    gopath --workspace ./my-learning next
    gopath --workspace ./my-learning check 07-k8s-client-go
    gopath --workspace ./my-learning check --integration 07-k8s-client-go

进入学习目录或其中某题目录后也能直接执行 `gopath next`。设置环境变量时，PowerShell 可使用 `$env:GOPATH_HOME = (Resolve-Path ./my-learning).Path`；Linux 可使用 `export GOPATH_HOME=/absolute/path/my-learning`。

| 命令 | 行为 |
|---|---|
| `init [目录]` | 复制题目说明、起始代码、必需测试和部署资产；有冲突则拒绝覆盖 |
| `list` | 列出全部练习与检查状态 |
| `next` | 显示第一道尚未完成的练习、说明路径和检查命令 |
| `check [id]` | 默认检查下一题；运行固定的 Go 单测 |
| `check --integration [id]` | 运行单测及显式开启的真实环境检查 |
| `hint [id]` | 按顺序展示一条提示，记录使用次数 |
| `reset <id>` | 列出将覆盖的文件，输入 `RESET <id>` 后恢复并清除该题记录 |
| `stats` | 显示自动检查完成数、实操待验收数、失效成绩数与提示次数 |

`watch` 留待后续版本实现。前置课程只是建议，可通过显式 ID 跳到 Kubernetes 练习。`next` 停在“实操待验收”题时，也可指定其他题的 ID 继续离线学习。

## 完成的含义

- `[ ]`：尚未通过；查看提示不会算完成。
- `[~]`：本地单测已通过，真实环境验收仍未通过。
- `[x]`：该题全部内置自动检查通过。
- `[!]`：代码、测试、模块或部署配置已改变，需要重新检查。

07–15 有独立的 integration 检查。普通 `check` 不会把 fake client、fixture 或配置 JSON 的成功当作真实部署成功。真实检查的前置环境、命令、所需变量和清理方式见每题 README；缺少环境会失败，不会跳过并记为成功。

检查使用固定参数 `go test -mod=mod -count=1 -json -timeout=2m ./...`，真实检查额外添加 `-tags=integration`；禁用外层 go.work 和 GOFLAGS 干扰。执行器要求实际测试通过，不接受零测试或跳过的测试。必需测试须与内置版本相同，可另外添加自己的测试。失败会清除旧的通过状态，保留提示次数。

检查期间请等待测试结束再编辑。工具会复查必需测试和输入指纹，途中修改源码、go.mod 或部署配置会拒绝记分并要求重跑；自动生成的 go.sum 可以保留。同一学习工作区的 check、hint、reset 使用写入锁，多个终端同时操作时会报忙，避免互相覆盖进度。

进度持久化在工作区 `.gopath/progress.json`，使用临时文件替换。检查后的源文件指纹用于识别后续修改；它是学习辅助记录，不是防作弊或考试系统。口述、实验观察与代码可读性需要人工复盘，自动检查不宣称能判断这些能力。

## 课程路线

| 阶段 | 练习 |
|---|---|
| Go 最小前置 | 01 错误包装、02 接口注入、03 Context、04 有界并发、05 HTTP 校验、06 slog |
| 最高优先级 | 07 client-go List/Watch/Informer、08 CRD Controller、09 Operator/Reconcile、10 Helm 生命周期 |
| 容器硬性基础 | 11 namespace、12 cgroup v2、13 Pod/Deployment/Service 与 Go 健康检查 |
| 可观测工程 | 14 Prometheus 埋点和抓取、15 Grafana 看板和查询 |
| 工程补充 | 16 退避重试、17 限流、18 SQL 事务、20 就绪与请求排空 |
| 少量概念 | 19 调度器过滤/评分与 Pending 原因 |

Kubernetes 题全部包含 Go 编码、离线测试和真实集群验证入口；Helm 题还包含 Chart。容器题需要完成 Go 代码和部署实验，并口述 namespace/cgroup、资源请求/限制、Pod/Deployment/Service 的区别。

## 开发与验证

    gofmt -w cmd internal
    go vet ./...
    go test ./...

模板以 `.go.txt` 保存，因此普通 `go test ./...` 检查 CLI 和课程格式，不会把待解题目编译进工具。维护者可显式运行以下课程作者检查：逐题验证 starter 可编译但测试失败、参考解答通过，并仅编译真实环境检查。

    go test -tags=curriculum -parallel=2 -timeout=30m -v ./exercises

该作者检查读取仓库中的参考答案，与学习者的 `gopath check` 分开；它不写学习者进度，也不执行真实环境测试。耗时取决于首次依赖下载。

详细说明：[实施计划](docs/PLAN.md)、[操作规程](docs/SOP.md)、[实现任务](docs/TASK.md)、[架构](docs/architecture.md)、[JD 映射](docs/jd-mapping.md)、[验证记录](docs/verification.md)、[新增练习](CONTRIBUTING.md)。

初始化不会覆盖已有文件。重置只覆盖列出的课程管理文件，保留额外文件，必须明确确认；重要解答请先自行备份或提交到版本控制。Go 测试会执行学习者代码，GoPath 不提供系统级沙箱。
