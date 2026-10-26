# GoPath MVP 实施计划

## 目标与范围

用小而可验证的 Go 代码练习建立岗位技能证据。课程不是教程堆积，也不保证完成后即可胜任岗位。首版 20 题；优先级为 Kubernetes 扩展开发、容器底层和工作负载、可观测实践、基础分布式工程。调度仅一题概念模型，不读复杂调度器源码。

已批准的首版命令为 init、list、next、check、hint、reset、stats。watch、在线题库、GUI、数据库、自动 JD 解析、复杂 GPU 调度、生产 Operator 框架均不在范围内。

## 结构与主要决策

| 组件 | 选择及理由 |
|---|---|
| CLI | 标准库实现，固定命令解析，保持构建依赖小 |
| 题库 | JSON 元数据、Markdown 说明、.txt 模板，通过 embed 打包 |
| 学习目录 | 与工具源码分开；每题独立 module；starter/tests 映射到题目根部 |
| 测试执行 | 固定 Go 命令，不从题库元数据读取可执行 shell 命令 |
| 答案 | 单独 solution 目录，不嵌入、不安装、不用于学习者验收 |
| 状态 | .gopath/progress.json，版本化 JSON 与临时文件替换 |
| 并发写入 | check/hint/reset 用排他文件锁串行化，防止多终端丢失更新 |
| 状态有效性 | 通过时间加源文件指纹；提示次数不能成为通关证据 |
| 实操 | integration 元数据 + Go build tag；单测通过与实操通过分开 |
| 覆盖保护 | 初始化全量预检、排他创建；重置列文件并要求精确确认；拒绝链接路径 |

JSON 比 YAML 少一个核心依赖。模板后缀避免 Go 根模块误编译故意有缺陷的练习。独立 module 让 Kubernetes/Prometheus 依赖不污染 CLI。先用明确选择的 integration 检查保留真实实操门槛，不把 fake client 的结果冒充部署结果。

## 课程分期

1. 01–06：错误、接口注入、取消、有界并发、HTTP、结构化日志，是 Kubernetes 学习的必要前置。
2. 07–10：List-Watch/Informer、自定义 CRD 与调谐队列、Operator 维护 Deployment、Go 驱动 Helm 生命周期。
3. 11–13：Go 读取 Linux namespace/cgroup，健康检查服务、Dockerfile、Deployment/Service 和口述复盘。
4. 14–15：Prometheus 真实埋点/抓取、Grafana provisioning/查询/看板。
5. 16–20：有限退避、线程安全限流、SQL 事务、调度基本概念、优雅排空。

每题专注一个主要目标，有故意缺陷、确定性单测、渐进提示、独立解答；真实环境题另外提供带 tag 的 Go 集成检查。

## 里程碑与验收

| 里程碑 | 验收 |
|---|---|
| 发现与安装 | 元数据可发现，目录稳定排序；说明、代码、测试、嵌套资产正确安装 |
| 垂直闭环 | 单题修改/测试/记录/重启可用；失败、跳过、零测试不得通关 |
| 交互 | 提示不完成；统计准确；重置需确认；未知输入有明确错误 |
| Kubernetes 专项 | 07–10 是 Go 实操；含 informer/CRD/controller/operator/chart 与独立实操检查 |
| 容器与可观测 | 11–15 含 Go、部署/实验材料及可执行的真实环境验收 |
| 课程质量 | 作者检查验证所有 starter 编译但失败，解答通过，集成测试代码编译 |
| 发布复核 | gofmt、vet、核心测试、临时工作区 CLI 流程、Windows 构建及 Linux 交叉构建 |

## 风险和边界

真实环境检查需要学习者先准备 Docker/minikube/Helm、Linux cgroup v2 或 Prometheus/Grafana；环境故障必须明确失败。工具不自动启动整套基础设施。当前开发机 Docker daemon 未运行、kubectl 未设 context，因此本次可验证的范围是离线测试与真实检查的编译；运行结果见 [验证记录](verification.md)。

口述与实验解释需要人工复核；状态只表示自动检查。指纹不是安全证明，固定测试也不是系统沙箱。初始化不支持覆盖升级；更新题库建议创建新工作区并有选择地迁移自己的代码。进度替换不承诺跨文件事务或 Windows 断电原子性。重置中途遇到磁盘错误可能部分恢复，重要代码先备份。

未来可按真实 JD 增减元数据与练习，无需改命令路由；新增检查器应有明确需求和行为测试后再扩展。
