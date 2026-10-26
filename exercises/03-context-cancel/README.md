# Context 取消传播

实现 `WaitForCancellation`，等待传入的 context 结束并返回标准取消原因。本题不需要计时器、轮询或外部服务。

在初始化后的本题目录运行 `go test ./...`，或运行 `gopath check 03-context-cancel`。测试覆盖已取消、已超时和运行中被取消的 context。

验收：context 有效时保持等待；取消后及时退出并保留 `context.Canceled`；超时后保留 `context.DeadlineExceeded`。测试中的超时只是防止错误实现永久阻塞。

口述复盘：`Done()` 和 `Err()` 各表达什么？取消信号为什么不能强制终止任意 goroutine？HTTP handler、Controller 和数据库调用如何向下传递取消？
