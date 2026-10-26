# 用 Go 生成 Grafana 看板并验证数据链路

完成 `exercise.go` 的 `BuildDashboard(datasourceUID)`：校验数据源 UID，用 Go 结构体生成可由文件 provisioning 加载的 dashboard JSON。`cmd/dashboard` 已提供落盘工具，不需要手工在 UI 里反复编辑看板。看板 UID 固定为 `gopath-http`，标题 `GoPath HTTP Service`，schemaVersion 为 39，version 为 1，刷新 5 秒，时间范围最近 15 分钟。

## 离线代码验收

运行 `go test ./...`。数据源 UID 接受 1–40 个 ASCII 字母、数字、下划线或连字符；非法输入返回错误且不返回 JSON。两个 timeseries 面板横向各占 12 格、宽 12 高 8，ID 分别为 1、2，引用调用者传入的 UID，查询 RefID 为 `A`，图例为 `{{route}}`。相同输入的 JSON 应稳定一致。

请求速率面板标题 `Request rate`，单位 `reqps`：

```promql
sum by (route) (rate(gopath_http_requests_total{job="gopath-service"}[5m]))
```

延迟面板标题 `P95 latency`，单位 `s`：

```promql
histogram_quantile(0.95, sum by (le, route) (rate(gopath_http_request_duration_seconds_bucket{job="gopath-service"}[5m])))
```

Counter 是累计值，`rate` 表达每秒增长量并处理计数器重置。直方图的 `le` 是 bucket 上界，聚合时丢掉它就无法正确计算分位数。P95 不是平均值，也不能把各实例 P95 直接求平均。没有请求时延迟图出现空值属于正常现象，不能用伪造数据填满图。

## Docker / Grafana 实操验收

先完成第 14 题，保留它的 `gopath-metrics`、`gopath-prometheus` 容器以及 `gopath-observe` Docker 网络。本题的 Compose 只增加 Grafana；数据源从 Grafana 容器通过 `http://gopath-prometheus:9090` 访问 Prometheus。容器内 `localhost` 指向容器自身。

在本题练习目录执行：

```sh
go run ./cmd/dashboard
docker compose up -d
docker compose logs grafana
```

生成的 `dashboards/gopath-http.json` 必须来自自己的 Go 实现。Compose 将该目录与 `provisioning` 只读挂载，后者通过固定 UID 连接数据源和看板。已提供的 API 与配置对应固定镜像 `grafana/grafana:13.2.4`。浏览器打开 [GoPath 看板](http://127.0.0.1:3000/d/gopath-http)，登录本地学习账号 `admin` / `gopath-learning`。端口仅绑定回环地址；该账号只用于本实验。首次登录如提示修改密码，可修改后通过环境变量为验收提供新密码。

向第 14 题服务发一些真实请求，等待至少两个抓取周期：

```sh
curl http://127.0.0.1:2112/work
curl http://127.0.0.1:2112/work
curl http://127.0.0.1:2112/work
go test -tags=integration -run TestIntegration -v ./...
gopath check 15-grafana-dashboard --integration
```

PowerShell 使用 `curl.exe`。速率查询至少需要两个样本；先确认 Prometheus target 为 UP，再排查空图。缺少服务、密码不匹配、看板未加载或没有业务指标时，集成测试会明确失败，不会 Skip。可通过 `GOPATH_GRAFANA_URL`、`GOPATH_GRAFANA_USER`、`GOPATH_GRAFANA_PASSWORD` 指定本实验的 Grafana 连接和凭证。默认值与 Compose 相同。

真实验收会读取 Grafana 健康 API、数据源与已加载看板，再经 Grafana 的数据源代理调用 Prometheus 查询 API，要求 `up{job="gopath-service"}` 为 1、`/work` 请求计数大于零，并执行两个面板的实际 rate/P95 表达式，要求返回有限的非负样本。样本不足或无近期流量时，等待抓取并发送请求后重试。它验证实际服务链路，离线 JSON 测试不能替代。此版本使用仍可用的 `/api` HTTP API；升级 Grafana 主版本时应根据官方迁移说明检查接口兼容性。

在看板观察请求速率和 P95 的变化，把命令、Grafana 版本、截图、查询结果及解释记录到 `observations.md`。结束时在本题目录执行 `docker compose down`，随后按第 14 题 README 清理其两个容器和网络。不要删除其他实验资源。

## 口述问题

1. 埋点、抓取、存储查询、展示分别由哪个组件负责？
2. 为什么请求总数通常用 Counter，请求耗时适合 Histogram？
3. 为什么 `rate` 后再按 route 聚合？为什么 Histogram 聚合要保留 `le`？
4. 数据源名称与 UID 有什么不同？为什么容器里的数据源 URL 不用 localhost？
5. 图上没数据时，如何依次定位请求、`/metrics`、Prometheus target、PromQL 与 Grafana 数据源？

参考：[Grafana 文件 provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/)、[Grafana 数据源 HTTP API](https://grafana.com/docs/grafana/latest/developer-resources/api-reference/http-api/api-legacy/data_source/)、[Prometheus 查询函数](https://prometheus.io/docs/prometheus/latest/querying/functions/)。自动验收通过后仍需完成观测记录与口述。
