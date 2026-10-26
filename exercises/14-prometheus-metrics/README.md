# Prometheus Go 埋点与独立 Registry

完成 `NewMetrics` 的注册与 `Instrument`。使用官方 `github.com/prometheus/client_golang v1.23.2`（其 go.mod 要求 Go 1.23.0），不手写指标文本。该依赖只属于本题，GoPath 核心不需要它。

`gopath_http_requests_total` 的标签为 `method`、固定的 `route`、最终 `status`；`gopath_http_request_duration_seconds` 是带 `method`、`route` 标签的 Histogram。`route` 使用 `/work` 这样的模板，不能放入 URL query、用户 ID 或完整原始路径。每次 `NewMetrics` 必须产生独立 registry；`/metrics` 本身不经过业务埋点。

## 离线验收

首次在网络可用时运行 `go mod download` 下载已固定的依赖，再运行 `go test ./...`。测试使用 recorder，不发网络请求、不依赖 Prometheus、不用 sleep。检查默认 200、显式状态码、重复 WriteHeader、计数及 Histogram 样本数；两次构造不能共享指标。

本题的 ResponseWriter 包装器面向普通 HTTP handler；不要求保留 Hijacker、Flusher 等可选接口，流式响应需另行设计。

## 真实抓取验收

以下在 Linux/WSL2 练习目录运行，需 Docker。若已在第 15 题或其他终端启动同名实验资源，复用它们，避免重复创建。镜像与配置版本固定，监听端口仅绑定回环地址。

```sh
docker network create gopath-observe
docker build -t gopath-metrics:v1 .
docker run -d --name gopath-metrics --network gopath-observe -p 127.0.0.1:2112:2112 gopath-metrics:v1
docker run -d --name gopath-prometheus --network gopath-observe -p 127.0.0.1:9090:9090 --mount "type=bind,src=$(pwd)/prometheus.yml,dst=/etc/prometheus/prometheus.yml,readonly" prom/prometheus:v3.5.0
curl http://127.0.0.1:2112/work
curl http://127.0.0.1:2112/metrics
```

打开 Prometheus `http://127.0.0.1:9090/targets`，等待 `gopath-service` 为 UP；配置每 5 秒抓取，首次抓取完成后再验收：

```sh
go test -tags=integration -run TestIntegration -v ./...
```

也可以在已初始化的学习工作区运行 `gopath check 14-prometheus-metrics --integration`。若改了端口，可通过 `GOPATH_PROMETHEUS_URL` 与 `GOPATH_METRICS_URL` 提供 base URL；默认分别为 `http://127.0.0.1:9090` 和 `http://127.0.0.1:2112`。集成测试调用真实服务 `/work`，检查 `/metrics`，再查询独立 Prometheus 的 HTTP API；没有服务时明确失败，不跳过。

在 Prometheus 查询 `sum(rate(gopath_http_requests_total{job="gopath-service"}[5m]))`，解释累计量与速率的差异。为第 15 题保留服务时不要停止。全部实验结束后先在第 15 题目录执行 `docker compose down`（若启动过 Grafana），再清理本题资源：

```sh
docker stop gopath-metrics gopath-prometheus
docker rm gopath-metrics gopath-prometheus
docker network rm gopath-observe
```

只清理这次实验创建的资源；不要删除共享容器或网络。若仅临时停止，可用 `docker start gopath-metrics gopath-prometheus` 恢复，不能重复创建同名容器。

## 口述问题

1. 为什么请求总量用 Counter，当前队列长度用 Gauge？
2. 为什么 `route="/users/123"` 容易引发高基数问题？
3. Histogram 的 `_bucket`、`_sum`、`_count` 分别表示什么？
4. `/metrics` 正常但 `up=0` 可能是哪一段网络或配置出错？

参考：[Prometheus Go 应用指南](https://prometheus.io/docs/guides/go-application/)、[client_golang v1.23.2 go.mod](https://github.com/prometheus/client_golang/blob/v1.23.2/go.mod)、[Prometheus HTTP API](https://prometheus.io/docs/prometheus/latest/querying/api/)。
