# 第一次使用 GoPath

这份教程带你完成第一次练习，再进入 Kubernetes 实操。主要示例使用 Windows PowerShell；项目在本机的位置是 `E:\Desktop\cloud`，放在其他位置时替换这个路径即可。

GoPath 是一个终端练习工具。每道题给你一段可以编译、但行为有缺陷的 Go 代码；你阅读说明和测试，修改代码，用检查命令验证结果。CLI 输出主要是英文，题目说明和提示是中文。

## 1. 准备 Go 并构建工具

开始第一题只需要 Go 1.27 或更高版本、文本编辑器，以及首次下载课程依赖所需的网络。Docker、Minikube 等环境可以到对应实操题再准备。

打开 PowerShell，进入项目目录：

~~~powershell
Set-Location 'E:\Desktop\cloud'
go version
go build -o .\build\gopath.exe .\cmd\gopath
~~~

`go version` 应显示 Go 1.27 或更高版本；构建成功通常没有输出。如果提示找不到 go，先安装 Go 并确认其命令已加入 PATH，再重新打开 PowerShell。如果构建报错，先处理报错再继续。

已有 build/gopath.exe 时可以直接使用，但执行练习的测试仍需要 Go。

## 2. 创建自己的学习工作区

在刚才的 PowerShell 中执行：

~~~powershell
$GoPathExe = (Resolve-Path .\build\gopath.exe).Path
$GoPathWorkspace = Join-Path (Get-Location).Path 'gopath-workspace'

& $GoPathExe init $GoPathWorkspace
& $GoPathExe --workspace $GoPathWorkspace list
& $GoPathExe --workspace $GoPathWorkspace next
~~~

PowerShell 的 `&` 用于执行变量中保存的程序路径。后面的命令沿用这两个变量，因此切换目录后仍能找到工具和学习目录。

`init` 复制 20 道题的说明、起始代码、测试和部署材料。每个学习工作区只初始化一次；已有工作区提示 already initialized 时，直接运行 list/next 继续学习。

请区分这两个位置：

| 位置 | 用途 |
|---|---|
| `E:\Desktop\cloud` | 工具源码和课程模板，包含 README、cmd、internal、exercises |
| `E:\Desktop\cloud\gopath-workspace` | 你解题的目录，包含实际的 exercise.go、测试和学习进度 |

编辑学习目录中的文件。仓库 exercises 里的 .go.txt 是课程模板；修改它不会直接改变已经初始化的练习。

第一次 `next` 会显示类似：

~~~text
Next: 01-go-errors — 错误包装与判断
Read: ...\gopath-workspace\01-go-errors\README.md
Edit: ...\gopath-workspace\01-go-errors
Run: gopath --workspace "..." check 01-go-errors
~~~

Read 是说明文件，Edit 是题目目录。next 只显示信息，不会替你打开编辑器或切换目录。

## 3. 亲手完成第一题

### 阅读说明，先看一次失败

用自己的编辑器打开整个学习目录，或用记事本阅读第一题：

~~~powershell
notepad (Join-Path $GoPathWorkspace '01-go-errors\README.md')
& $GoPathExe --workspace $GoPathWorkspace check 01-go-errors
~~~

第一次检查应该失败，能看到：

~~~text
--- FAIL: TestFindUserMissingWrapsSentinel
    ... errors.Is(...) = false
Checks failed; exercise is not complete.
~~~

这是题目预设的缺陷。该测试表示：函数返回了错误文本，但调用者不能用 errors.Is 判断它的原因。阅读题目，确认 FindUser 缺少什么行为。

### 修改代码，需要时取一条提示

~~~powershell
notepad (Join-Path $GoPathWorkspace '01-go-errors\exercise.go')
& $GoPathExe --workspace $GoPathWorkspace hint 01-go-errors
~~~

只修改本题要求修改的代码，并保存文件。可以阅读 exercise_test.go 理解验收条件，保留必需测试；也可以增加自己的测试。

每运行一次 hint 展示下一条提示；提示次数会被记录，但不会让题目完成。先尝试，再取下一条提示。

### 再检查，确认完成

保存修改后执行：

~~~powershell
& $GoPathExe --workspace $GoPathWorkspace check 01-go-errors
~~~

若仍失败，就根据失败的测试继续调整。通过后会看到：

~~~text
Passed all required checks: 01-go-errors — 错误包装与判断
~~~

然后查看成绩和下一题：

~~~powershell
& $GoPathExe --workspace $GoPathWorkspace stats
& $GoPathExe --workspace $GoPathWorkspace list
& $GoPathExe --workspace $GoPathWorkspace next
~~~

在全新工作区中完成第一题后，stats 应显示 Completed 1/20 exercises，list 中第一题为 [x]，next 指向 02-go-injection。

直接在题目目录运行 `go test ./...` 也能获得测试反馈；`gopath check` 才会写入学习进度。项目源码根目录的 `go test ./...` 检查的是工具自身。

## 4. 以后每天怎么继续

重新打开 PowerShell 后，重新设置路径变量即可读取已有进度：

~~~powershell
$GoPathProject = 'E:\Desktop\cloud'
$GoPathExe = Join-Path $GoPathProject 'build\gopath.exe'
$GoPathWorkspace = Join-Path $GoPathProject 'gopath-workspace'

& $GoPathExe --workspace $GoPathWorkspace next
& $GoPathExe --workspace $GoPathWorkspace stats
~~~

通常按“读说明 → 看失败 → 修改并保存 → check → 复盘口述问题”的顺序完成一题。当前版本需要手动运行 check，尚未提供 watch 自动检查。

不带 ID 的 next、check、hint 选择整个工作区中第一道未完成的题。即使进入 09 的目录，也不会自动把 check 切到 09；练指定题时，始终带上完整 ID：

~~~powershell
& $GoPathExe --workspace $GoPathWorkspace check 09-k8s-operator
& $GoPathExe --workspace $GoPathWorkspace hint 09-k8s-operator
~~~

进度保存在学习目录的 .gopath/progress.json，关闭终端后仍保留。请保留学习目录；课程没有云端同步。要迁移时复制整个学习目录，包括 .gopath。

| 标记 | 含义 | 接下来做什么 |
|---|---|---|
| [ ] | 尚未通过 | 修改代码，运行 check |
| [~] | 单测通过，真实环境待验收 | 按本题说明准备环境，再运行 check --integration |
| [x] | 本题全部内置自动检查通过 | 复盘口述问题，然后继续下一题 |
| [!] | 上次通过后相关文件发生变化 | 保存修改，重新检查 |

等待 check 结束后再编辑文件。检查期间修改代码或配置，会要求重新运行，避免把未测试的版本记为完成。

## 5. 怎么进入 Kubernetes 重点练习

Go 基础不熟时，先完成 01–06。已有 Go 基础，可以直接指定 07–10，CLI 不强制按顺序解题：

| 顺序 | 重点 |
|---|---|
| 07–10 | client-go List/Watch/Informer、CRD Controller、Operator 调谐、Helm 生命周期 |
| 11–13 | namespace、cgroup、Go 服务与 Pod/Deployment/Service |
| 14–15 | Prometheus 埋点和真实抓取、Grafana 看板和查询 |
| 16–20 | 基础工程补充；19 仅学习调度概念 |

先读 07 的学习目录说明并完成 Go 代码：

~~~powershell
Set-Location (Join-Path $GoPathWorkspace '07-k8s-client-go')
notepad .\README.md
& $GoPathExe --workspace $GoPathWorkspace check 07-k8s-client-go
~~~

07–15 分为两层检查。本地单测在依赖下载完成后可使用 fake、fixture 或本地测试服务运行；真实环境检查必须先按本题 README 启动相应环境。

07 本地通过后显示 [~]。按它的 README 准备容器运行时、Minikube、kubectl，确认 current-context 是 minikube，并完成 List/Watch 观察实验，再运行：

~~~powershell
& $GoPathExe --workspace $GoPathWorkspace check --integration 07-k8s-client-go
~~~

通过后才成为 [x]。08/09 还需要显式预装各题 CRD；10 需要 Helm；11/12 需要 Linux/WSL2；13–15 按题目准备部署、Prometheus/Grafana。具体命令、成功信号和清理步骤都在各题 README。

题目说明中的 `gopath check ...`，在本教程的 PowerShell 中写成 `& $GoPathExe --workspace $GoPathWorkspace check ...` 即可；go run、kubectl 等命令按题目要求在相应练习目录执行。

next 会停在 [~] 的题。暂时没有真实环境时，可显式检查下一题，之后再补实操，例如：

~~~powershell
& $GoPathExe --workspace $GoPathWorkspace check 08-k8s-controller
~~~

缺少集群或服务时，实操检查会失败。单测通过、实操验收、口述解释是三种不同的学习证据。本项目实际执行过哪些验证，可查看 [验证记录](verification.md)。

## 6. 想把一题重新做一遍

先把需要保留的解答备份到学习目录外，再执行：

~~~powershell
& $GoPathExe --workspace $GoPathWorkspace reset 01-go-errors
~~~

它会先列出将覆盖的文件。只有你输入下面这行并按回车，才会恢复起始代码，同时清除本题完成记录和提示次数：

~~~text
RESET 01-go-errors
~~~

空输入或其他内容会取消。额外笔记和文件会保留；额外 .go 文件仍参与编译，所以 reset 后若出现重复声明，也要检查自己添加的文件。

## 常见问题

| 现象 | 处理 |
|---|---|
| 提示找不到 gopath | 本教程使用 `& $GoPathExe`，无需把工具安装进 PATH；重开终端后按第 4 节重设变量 |
| workspace is not initialized | 检查 GoPathWorkspace 是否指向已经 init 的目录；新目录先初始化 |
| next 一直显示同一题 | 它还没通过全部检查；查看 list，尤其留意 [~] 和 [!] |
| 手动 go test 通过，但完成数没变 | 对该题执行 gopath check，才能记录进度 |
| 第一次依赖下载失败或很慢 | 查看原始报错，确认 Go 版本及模块下载网络；修复下载环境后重试 |
| 提示 go.mod 在测试中变化 | 依赖解析更新了模块文件；确认修改后重新 check 一次 |
| required test 缺失或被修改 | 恢复同版本必需测试；或备份自己的代码后确认 reset 整题 |
| workspace is busy | 等另一个 check/hint/reset 结束；异常退出留下的锁按 [SOP](SOP.md) 核对后处理 |
| integration 无法连接 Docker、集群或服务 | 回到本题 README，先完成环境准备和就绪检查 |

## Linux 入口

在 Linux 的项目根目录中可用同样流程：

~~~bash
go version
go build -o ./build/gopath ./cmd/gopath
./build/gopath init ./gopath-workspace
./build/gopath --workspace ./gopath-workspace next
./build/gopath --workspace ./gopath-workspace check 01-go-errors
~~~

Windows 的 exe 不能直接作为 Linux 程序运行。WSL 中请在 Linux 环境构建工具、准备 Go 和相应实验环境。

更多说明：[项目首页](../README.md)、[学习与故障处理 SOP](SOP.md)、[课程与 JD 映射](jd-mapping.md)。参考解答保存在仓库每题的 solution 目录，可以解题后对照；初始化不会把它们复制到学习目录。
