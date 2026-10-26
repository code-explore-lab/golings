# 09 — Operator = CRD + Controller + 领域调谐

AppPlan 描述期望的副本数和镜像，Go Reconcile 把它实现为 Deployment。完成 EnsureDeployment：创建受 AppPlan 控制的子对象；只更新自身管理的字段；已收敛就不写；拒绝接管其他人拥有的 Deployment。

所有命令从 gopath init 安装后的本练习目录执行。

## 本地反馈

~~~powershell
go test -count=1 ./...
gopath check 09-k8s-operator
~~~

本地用 fake API 验证真实 Reconcile、Deployment spec、ownerReference、UID 所有权、无写入的幂等性、API 失败以及 informer 对子资源漂移的自动修复。使用 Go 1.27 和固定的 Kubernetes Go 模块 v0.37.1。首次下载后，单元测试无需集群。

RunOperator 有两个 informer：AppPlan 变化排入自身 key；受控 Deployment 变化通过 ownerReference 排入父 AppPlan key。worker 只调用一个 Reconcile。模板没有依赖 controller-runtime，便于看清它在更高层封装了什么。

## Minikube 实操

先显式安装集群级 CRD；测试自身只读取 CRD，不会安装或清理 CRD：

~~~powershell
minikube start --kubernetes-version=v1.37.1
kubectl config use-context minikube
kubectl --context minikube apply -f crd.yaml
kubectl --context minikube wait --for=condition=Established --timeout=60s crd/appplans.learning.example.io
kubectl --context minikube create namespace gopath-operator
go run ./cmd/operator -namespace gopath-operator
~~~

在第二个终端观察 Go 调谐产生的实际 Deployment：

~~~powershell
kubectl --context minikube -n gopath-operator apply -f example.yaml
kubectl --context minikube -n gopath-operator get appplan,deployment,pod
kubectl --context minikube -n gopath-operator get deployment demo -o yaml
kubectl --context minikube -n gopath-operator scale deployment demo --replicas=4
kubectl --context minikube -n gopath-operator get deployment demo --watch
~~~

副本数应被改回 AppPlan.spec.replicas。停止 kubectl watch 后，修改父资源，再验证新期望状态；删除父资源后，垃圾回收应删除其 Deployment：

~~~powershell
kubectl --context minikube -n gopath-operator patch appplan demo --type=merge --patch-file update.yaml
kubectl --context minikube -n gopath-operator get deployment demo
kubectl --context minikube -n gopath-operator delete appplan demo
kubectl --context minikube -n gopath-operator get deployment
~~~

Ctrl+C 停止 Go Operator，并清理本次手工命名空间：

~~~powershell
kubectl --context minikube delete namespace gopath-operator
gopath check 09-k8s-operator --integration
~~~

真实验收要求 current-context=minikube，手工预装 AppPlan CRD；测试创建新的 gopath-09-* 命名空间并只清理自己的对象。GOPATH_NAMESPACE 必须是未存在的新 gopath-* 名字。验收创建带真实 UID 的 AppPlan，验证 ownerReference、重复调谐后期望 spec 与 generation 稳定、子资源漂移修复及父对象删除后的垃圾回收。本地 fake 测试另外严格断言零 API 写入；真实集群的原生 Deployment controller 可能写 status 并改变 resourceVersion，因此不能用 resourceVersion 判断幂等。初始副本数为 0，后续检查 API spec，不要求镜像拉取就绪。普通 check 仅标为待集成。

## 能口述的解释

Operator 把领域知识放进自定义资源和持续调谐逻辑：不仅描述资源，还自动维护它。ownerReference 中的 UID 防止同名重建对象被当成旧主人；Controller 标记表达控制关系，垃圾回收据此处理子资源。finalizer 用于删除前的外部清理，是另一种机制。这里没有外部资源，因此无需 finalizer，也未实现 status.conditions、准入 webhook、leader election 或生产发布流程。

参考：[Operator 模式](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)、[属主与附属对象](https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/)。
