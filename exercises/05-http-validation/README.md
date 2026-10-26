# HTTP 请求校验

实现 `CreateUserHandler`：解码 JSON 中的 `name`，去除首尾空白并校验，再返回新建用户响应。测试通过 `httptest` 调用 handler，无需启动服务器。

在初始化后的本题目录运行 `go test ./...`，或运行 `gopath check 05-http-validation`。

验收：非法 JSON、类型错误、缺失或空白名称返回 HTTP 400；有效输入返回 HTTP 201，响应的 `Content-Type` 为 `application/json`，JSON 中的名称已去除首尾空白。注意响应头与状态码的写入顺序。

口述复盘：解码成功为什么不等于输入有效？客户端错误和服务端错误如何区分？为什么应该在 HTTP 边界完成输入校验？
