# JD 要求与练习证据

优先级按本项目的求职目标排序。Go 前置课用于支撑最高优先级实践；调度只占一题，不深入复杂算法或调度器源码。前置字段用于学习建议，CLI 不强制锁题。

| 优先级 / JD 要求 | 练习 | 可复核证据 |
|---|---|---|
| 最高：client-go、Informer、List-Watch | 07-k8s-client-go | Go List 与事件处理；fake 行为测试；minikube Pod 变更检查 |
| 最高：CRD 与 Controller | 08-k8s-controller | CRD 驱动 ConfigMap；调谐队列、错误重试、幂等写入；集群测试 |
| 最高：Operator、Reconcile | 09-k8s-operator | 自定义资源声明 Deployment 期望状态；幂等调谐、漂移修复及所有权 |
| 最高：Helm 基础 | 10-helm-lifecycle | Go 注入 Runner 的行为测试；Chart、打包、安装、升级、回滚 |
| 硬门槛：namespace | 11-container-namespaces | Go 解析/读取 /proc/self/ns；Linux 主机/容器观察与原理口述 |
| 硬门槛：cgroup | 12-container-cgroups | Go 解析 cpu.max/memory.max；受限容器实测；限制/隔离差异解释 |
| 硬门槛：Pod/Deployment/Service | 13-k8s-workloads | Go 健康检查；Dockerfile、探针、Deployment/Service，真实服务检查 |
| 可观测：结构化日志 | 06-structured-logging | slog JSON 事件、HTTP method/path/status 字段测试 |
| 可观测：Prometheus | 14-prometheus-metrics | 官方 client_golang、独立 registry、指标端点及真实 Prometheus 查询 |
| 可观测：Grafana | 15-grafana-dashboard | Go 生成/校验 dashboard，provisioning 与查询，真实 Grafana API 验收 |
| 加分：算力调度基础认知 | 19-scheduler-concepts | requests/可用量、过滤/评分、Pending 原因的简单 Go 模型；不实现复杂算法 |
| Go 错误与接口设计 | 01-go-errors、02-go-injection | 错误链、注入替身、错误传播 |
| Context 与有界并发 | 03-context-cancel、04-worker-pool | 取消与 deadline、结果顺序、工作并发与数量限制；可额外运行 race |
| HTTP/API 设计 | 05-http-validation | 输入校验与 HTTP 状态码表格测试 |
| 基础故障处理 | 16-retry-backoff、17-rate-limiter | 可取消有限重试、可控时间与线程安全令牌桶测试 |
| 数据库工程 | 18-sql-transactions | database/sql 事务提交、回滚与错误处理测试；不冒充真实数据库调优 |
| 服务生命周期 | 20-engineering-capstone | readiness、停止接收新请求、等待在途请求排空的确定性测试 |

完成证据分层：普通单测验证局部行为，integration 检查验证真实环境结果，人工复盘验证能否解释机制与排查过程。07–15 本地测试成功只显示“实操待验收”。课程完成不保证岗位胜任，也不代表掌握分布式系统全部复杂性。

后续按真实 JD 定制时，可增减练习目录，调整 metadata 的 category、objectives、prerequisites 和 acceptance，并更新此映射；无需修改无关 CLI 路由。
