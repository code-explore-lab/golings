# 有界 Worker Pool

实现 `MapBounded`：最多使用 `workers` 个 worker 并发转换输入，按输入顺序返回结果。`workers < 1` 按一个 worker 处理；空输入返回空结果。

在初始化后的本题目录运行 `go test ./...`，或运行 `gopath check 04-worker-pool`。测试同时检查并发重叠、并发上限、结果顺序和边界输入。逐个串行处理也会失败。

通过后可运行 `go test -race ./...` 检查数据竞争。竞态检测需要平台支持及对应的 C 编译工具链。

口述复盘：为什么每个输入启动一个 goroutine 不等于有界并发？任务分发、结果下标和 worker 退出如何协作？增加 worker 数量在 CPU 密集与 I/O 密集场景分别有什么限制？
