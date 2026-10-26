# 08 — 用 Go 写极简 CRD Controller

完成 exercise.go 的 EnsureConfigMap：读取 Gadget.spec.message，在同一命名空间创建或更新同名 ConfigMap。保留其他 Data 键；没有变化时不发送 Update；API 失败交给队列重试。RunController 是完整可读的 Go 骨架，不需要控制器框架。

运行以下命令时，应位于 gopath init 安装后的本练习目录。

## 本地反馈

~~~powershell
go test -count=1 ./...
gopath check 08-k8s-controller
~~~

测试使用 fake 动态客户端和 typed 客户端，但真实启动 CRD informer 与限速工作队列，验证 CRD 更新、瞬态失败重试、成功后 Forget、取消和无变化时零写入。首次下载依赖后，本地检查不需要集群。固定 api / apimachinery / client-go v0.37.1，使用 Go 1.27。

事件回调只放入 namespace/name；worker 的 Reconcile 重新 Get 当前 Gadget。一个键可能合并多次事件，失败后重新排队，因此逻辑必须幂等。这里是单 worker 教学控制器；生产场景还需 RBAC、leader election、指标与冲突策略。

## Minikube 与 CRD 显式预装

CRD 定义是集群级操作，测试不会替你安装或删除。请在目标 Minikube 上显式执行：

~~~powershell
minikube start --kubernetes-version=v1.37.1
kubectl config use-context minikube
kubectl --context minikube apply -f crd.yaml
kubectl --context minikube wait --for=condition=Established --timeout=60s crd/gadgets.learning.example.io
~~~

只装 CRD 还没有调谐逻辑。启动 Go 控制器，在第二个终端创建自定义对象：

~~~powershell
kubectl --context minikube create namespace gopath-controller
go run ./cmd/controller -namespace gopath-controller
~~~

~~~powershell
kubectl --context minikube -n gopath-controller apply -f example.yaml
kubectl --context minikube -n gopath-controller get gadget demo
kubectl --context minikube -n gopath-controller get configmap demo -o yaml
kubectl --context minikube -n gopath-controller patch gadget demo --type=merge --patch-file update.yaml
kubectl --context minikube -n gopath-controller get configmap demo -o yaml
~~~

ConfigMap 中的 message 应从初始值变成更新值。再次应用相同的 Gadget 不应引起无效写入。本题只监听 Gadget，不自动修复单独修改 ConfigMap 的漂移；下一题会监听拥有的子资源并使用 ownerReference。

Ctrl+C 停止控制器，再清理手工命名空间并运行真实验收：

~~~powershell
kubectl --context minikube delete namespace gopath-controller
gopath check 08-k8s-controller --integration
~~~

集成测试必须使用 current-context=minikube；会检查 Gadget API 已可用，启动真实 Go 控制器并验证 Gadget 创建/更新和幂等性。测试只新建并清理自己的 gopath-08-* 命名空间；GOPATH_NAMESPACE 若设置，必须是新的 gopath-* 名字。CRD 会保留。普通 check 仅记为“待集成”，集成环境缺失或 CRD 未安装会明确失败。

## 能口述的解释

CRD 给 API server 注册一种资源，Controller 才实现其业务行为。调谐是“读取当前状态 → 计算差异 → 写入必要变化”，不是依次重播每一个事件。为什么队列存 key？为什么每个 Get 必须对应 Done？冲突或网络失败为什么要 AddRateLimited，成功为什么要 Forget？Get 返回 NotFound 为什么可能是正常删除？为什么不能每次循环都 Update？

参考：[client-go 工作队列示例](https://github.com/kubernetes/client-go/tree/v0.37.1/examples/workqueue)、[自定义资源](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)。
