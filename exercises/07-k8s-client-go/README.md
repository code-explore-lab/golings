# 07 — client-go Pod List / Watch / Informer

这题必须写 Go。完成 exercise.go 中 ListPods 和删除墓碑解包，使用真实 client-go 类型；cmd/observe 是可运行的入口。命令均在 gopath init 安装后的本练习目录执行，模板文件的 .txt 后缀由 CLI 自动移除。

## 本地反馈

~~~powershell
go test -count=1 ./...
gopath check 07-k8s-client-go
~~~

本地测试使用 fake API，但实际启动 client-go informer，验证初次 List、后续 Watch 更新/删除、命名空间范围、错误传播、取消与 tombstone。回调得到对象副本，不能改坏 informer 缓存。首次运行会下载固定版本的 Go 模块；缓存齐全后测试不需要集群。不要修改验收测试来绕过失败。

## Minikube 实操

需要 Go 1.27、kubectl、可运行的 Minikube。模块固定 client-go / api / apimachinery v0.37.1，目标 Kubernetes v1.37.1。

~~~powershell
minikube start --kubernetes-version=v1.37.1
kubectl config use-context minikube
kubectl --context minikube get nodes
kubectl --context minikube create namespace gopath-pods
go run ./cmd/observe -namespace gopath-pods
~~~

在第二个终端创建、修改、删除同一命名空间的 Pod，观察第一个终端的 add / update / delete 输出：

~~~powershell
kubectl --context minikube -n gopath-pods run watched --image=registry.k8s.io/pause:3.10
kubectl --context minikube -n gopath-pods label pod watched learner=updated --overwrite
kubectl --context minikube -n gopath-pods delete pod watched
~~~

用 Ctrl+C 停止观察程序，然后清理本次手工创建的命名空间：

~~~powershell
kubectl --context minikube delete namespace gopath-pods
gopath check 07-k8s-client-go --integration
~~~

集成测试要求 kubeconfig 的 current-context 恰为 minikube，真实调用 API 创建 Pod、List 并观察新增/删除。它新建 gopath-07-* 命名空间并仅删除自己创建的该命名空间；不会使用已有命名空间。可选 GOPATH_NAMESPACE 必须是尚不存在且以 gopath- 开头的名字。失败不会被 skip 当作通过。

普通 check 通过仅记为“本地通过，待集成”；只有 --integration 通过才完成本题。这里不等待 Pod Ready，验收的是 API 观察逻辑，不依赖暂停镜像成功拉取。

## 能口述的解释

List 是一次快照；Watch 从资源版本继续获取变化；Informer 把两者组合成缓存和事件分发。初次 List 中的已有对象也触发 Add。Watch 可能断开并重新 List，事件不等于只发生一次。删除事件为什么可能只剩 tombstone？回调为什么应短小、为什么不能修改缓存对象？取消 context 如何结束 watch 和 goroutine？

参考：[client-go 示例](https://github.com/kubernetes/client-go/tree/v0.37.1/examples)、[高效检测资源变化](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)。
