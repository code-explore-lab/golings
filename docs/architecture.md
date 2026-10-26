# 架构与状态语义

## 组件

| 目录 | 职责 |
|---|---|
| cmd/gopath | 信号取消、标准输入/输出、退出码 |
| internal/cli | 命令校验、工作区选择、提示/确认、状态转换 |
| internal/catalog | 从 fs.FS 发现 JSON，验证 ID/目录/检查类型，稳定排序 |
| internal/workspace | 模板映射、安装预检、排他创建、重置、安全路径、测试校验、指纹 |
| internal/runner | 固定 Go 命令、环境隔离、超时、JSON 事件与实际通过测试计数 |
| internal/progress | 版本化 JSON 加载与临时文件替换 |
| exercises | 内嵌题库；独立 module 的起始模板、测试、部署资产 |
| exercises/*/solution | 维护者参考代码，仅在源码中存在 |

## 数据流

init：内嵌 JSON → catalog → 工作区模板预检 → 复制 README/starter/tests → 建立 .gopath。

check：获取工作区写入锁 → 加载进度 → 找到题目 → 校验路径、必需测试和输入指纹 → 在题目 module 执行固定 Go 测试 → 检查真实 pass/skip 事件 → 复查必需测试和输入一致性 → 更新完整源文件指纹与通过时间 → 保存 JSON → 释放锁。

list/next/stats：读取 catalog 和进度 → 对比工作区指纹 → 展示有效完成、本地通过、过期或未通过状态。

reset：加载状态 → 列出管理文件 → 精确确认 → 预检目标 → 替换模板文件 → 清除该题状态。未知文件保留；参考答案不参与任何学习者命令。

## 模板映射

`starter/exercise.go.txt` → `<工作区>/<id>/exercise.go`；`tests/exercise_test.go.txt` → 同目录的 `exercise_test.go`。嵌套资产保留路径，例如 `starter/chart/templates/deployment.yaml.txt` → `chart/templates/deployment.yaml`。README 一并安装，exercise.json 由工具读取，solution 不嵌入。

目标路径必须留在工作区内；文件或父目录是符号链接时拒绝操作。初始化先预检全部文件并用 O_EXCL 创建，重置先预检整题。它们防止常见误覆盖，不承诺抵抗恶意并发目录替换，也不提供执行沙箱。

## 状态与验收

每题记录 `hints_used`、`local_passed_at`、`completed_at` 和 `fingerprint`。记录存在不等于完成。时间戳非零且指纹匹配时成绩才有效。指纹覆盖 Go 源码、go.mod/go.sum、JSON、YAML、Dockerfile；构建出的二进制和普通笔记不影响成绩。

测试前后的输入指纹排除题目根目录的 go.sum，允许 Go 首次生成依赖校验和；其余相关输入必须一致。若 Go 更新了 go.mod，明确要求重跑；生成后的 go.sum 会纳入最终完整指纹。检查期间的编辑不能把旧编译产物的成功记到新源码上；这仍不承诺抵抗恶意并发修改。

普通题单测通过即设置 completed_at。`integration:true` 的题只设置 local_passed_at；明确运行 --integration 并通过正常测试及 TestIntegration 后才设置 completed_at。失败清除旧通过记录但保留提示次数。相关文件改变后成绩显示 stale；显式重新检查后才能恢复。unknown ID 的旧记录不计入当前课程统计。

执行器运行 `go test -mod=mod -count=1 -json -timeout=2m ./...`，实操额外带 `-tags=integration`。它固定 GOWORK=off、清除 GOFLAGS，并设置 GOPATH_INTEGRATION。调用不经过 shell，元数据不能提供任意命令。整个进程限十分钟，Go 测试本身限两分钟。零测试、任何跳过、退出失败都不能通过；integration 模式还要求至少一个 TestIntegration 通过。

这不是考试防作弊系统；学习者控制机器与代码。进度只记录内置自动检查，原理口述与人工审阅另行复盘。原始测试输出直接用于诊断，参考解答不会进入输出。

进度文件在同目录创建临时文件、写入、Sync、Close、Rename。未知版本或坏 JSON 显式报错。文件替换避免常见截断，但不承诺 Windows 断电原子性或整个 reset 的跨文件事务。已初始化工作区不支持自动覆盖升级，防止用新课程静默替换学习者代码。

check、hint、reset 从读取进度前到保存结束持有 `.gopath/write.lock`（O_EXCL 创建），包括 reset 确认期间，防止多个 CLI 进程丢失更新。冲突立即报忙，不等待、不自动清除锁；list/next/stats 可读取上次保存的状态。正常退出或取消检查会释放锁。强制结束进程后可能留下锁，必须确认没有写入进程后人工移走该文件；锁中记录 PID 与开始时间。
