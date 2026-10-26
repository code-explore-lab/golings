# 结构化日志字段

实现 `NewRequestLogger` 和 `LogRequest`，使用标准库 `slog` 向注入的 writer 输出 JSON 请求日志，无需日志服务。

在初始化后的本题目录运行 `go test ./...`，或运行 `gopath check 06-structured-logging`。测试会解析多条不同请求的日志，避免把整段文本或固定值当成结构化字段。

验收：每次请求记录一条有效 JSON；`method`、`path`、`status` 为独立字段，状态码为数值；路径取 `r.URL.Path`；使用注入的 logger，不依赖全局默认 logger。

口述复盘：结构化日志怎样支持按状态码聚合？完整 URL 和路径有什么区别，查询参数是否适合记录？请求 ID 可以怎样在各层调用之间传递？
