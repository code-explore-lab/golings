# Go 健康检查与 Kubernetes 工作负载

完成 `Health.SetReady` 与 `Health.Handler`，让存活与就绪表达不同状态。初始未就绪；`SetReady(true)` 后可接收流量；切为 false 后 `/readyz` 返回 503，`/livez` 仍返回 200。只允许 GET，其他方法返回 405，未知路径返回 404。用并发安全方式读写就绪状态。

## 离线验收

运行 `go test ./...`；支持 race 工具链时再运行 `go test -race ./...`。提供的 `cmd/service` 监听 8080，收到终止信号先取消就绪，再以有界 context 调用 HTTP shutdown。测试使用 `httptest`，不需要真实监听端口、Docker 或 Kubernetes。

## Minikube 实操验收

前置：Docker、Minikube、kubectl 和可运行的专用学习集群。本题只操作 `minikube` context 的 `gopath-workloads` namespace。先确认这里没有同名的既有业务资源；使用本题专用学习资源。

```sh
minikube start -p minikube
kubectl --context minikube cluster-info
docker build -t gopath-health:v1 .
minikube -p minikube image load gopath-health:v1
kubectl --context minikube create namespace gopath-workloads
kubectl --context minikube apply -n gopath-workloads -f manifests/
kubectl --context minikube rollout status -n gopath-workloads deployment/gopath-health --timeout=90s
kubectl --context minikube get -n gopath-workloads pods,deployments,services,endpointslices
kubectl --context minikube port-forward -n gopath-workloads service/gopath-health 18080:80
```

保持最后一条在独立终端运行；另一终端使用 `curl -i http://127.0.0.1:18080/readyz`，应为 200。PowerShell 可使用 `curl.exe`。观察清单中的 `requests`、`limits`、Deployment selector 与 Service selector；不要只以 `kubectl apply` 返回成功判定部署完成。

修改 Go 代码后，要重新 build、load 镜像，并执行 `kubectl --context minikube rollout restart -n gopath-workloads deployment/gopath-health`，再等待 rollout status、重建端口转发。本题使用固定的 `v1` 标签，重复 apply 相同清单不会替换已运行的 Pod。

保持端口转发后，在练习目录运行 `go test -tags=integration -run TestIntegration -v ./...`，或从 GoPath 工作区运行 `gopath check 13-k8s-workloads --integration`。测试通过明确的 minikube context 查询 Deployment、Service、EndpointSlice，检查两个就绪副本、探针与 requests/limits，并通过实际转发端口请求 `/readyz` 和 `/livez`。可通过 `GOPATH_HEALTH_URL` 指定自己的转发地址；缺少 kubectl、集群、资源或端口转发会失败，不会 Skip。

在学习 namespace 中选择并记录一个 Pod 名，然后执行 `kubectl --context minikube delete pod -n gopath-workloads <pod-name>`；观察 Deployment 恢复两个就绪副本，并重新运行 rollout 检查和端口转发。删除的是实验副本，用来验证控制循环；不操作其他 namespace。将命令输出、Pod UID 的变化和解释保存在 `observations.md`。实验结束后只删除本题资源：`kubectl --context minikube delete -n gopath-workloads -f manifests/`。

## 口述问题

1. 为什么不能在短暂依赖故障时直接让 liveness 失败？readiness 又应该如何处理？
2. Deployment 与 Pod 的关系是什么？删除 Pod 与删除 Deployment 有何不同？
3. Service 的稳定性来自哪里？为什么 selector 不匹配时 Service 可能没有可用端点？
4. requests、limits、探针、terminationGracePeriodSeconds 各影响哪个阶段？
5. `Shutdown` 的有界等待解决什么问题？移除 endpoint 与进程退出为何可能有时间差？

参考：[Kubernetes 探针](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/)、[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)、[Service](https://kubernetes.io/docs/concepts/services-networking/service/)。本地通过不代表真实集群验收完成。
