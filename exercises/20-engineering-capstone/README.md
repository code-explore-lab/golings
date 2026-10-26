# 工程回归练习：就绪检查与排空

场景：服务已收到关闭信号，但仍接受新请求，导致请求在容器结束时被中断。补齐 Gate，以小改动修复准入和排空逻辑，并阅读测试中的缺陷复现。

接口约定：
- NewGate 返回初始就绪的门控对象。
- Ready 用于 readiness：正常 200，开始排空后 503；本题不把它作为 liveness。
- Wrap 包装业务 handler，排空开始前的请求继续执行；之后的新请求返回 503，不能调用业务 handler。
- BeginDrain 幂等，开始拒绝新请求。
- Wait(ctx) 在已进入的请求全部结束后返回 nil；未完成时响应 ctx 取消。先排空完成时，即使传入已取消的 context，也返回 nil。

所有共享状态必须安全同步；不能通过 sleep 猜测 handler 是否进入。测试通过 channel 协调请求和关闭顺序。支持 race 检测的环境可运行 go test -race ./...。

本题只实现应用内排空门控，不重复整套 Kubernetes 项目。真实上线时还需要处理信号、调用 http.Server.Shutdown，配置 readiness 和 terminationGracePeriodSeconds，并关注负载均衡传播延迟。口述这些连接关系即可。

