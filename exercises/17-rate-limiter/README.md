# 单进程令牌桶限流

实现 NewTokenBucket(capacity, refillEvery, now) 和 Allow(now)。容量与补充周期必须为正数。初始桶满，每经过一个完整周期补充一个令牌，Allow 原子地消耗一个令牌并返回 true；无令牌时返回 false。

now 由调用者提供，测试无需真实睡眠。保留未满周期的余量，空闲时令牌数不能超过 capacity。时钟倒退不补充，也不能把 last 倒退；并发调用必须安全。

运行 gopath check 17-rate-limiter；支持 race 检测的环境可在本练习目录额外运行 go test -race ./...。

口述问题：为什么速率为每秒 1 个令牌却可以瞬时通过多个请求？为什么该实现不能对多个副本提供全局限流？

