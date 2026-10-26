# 10 — 用 Go 驱动 Helm 发布生命周期

完成 exercise.go 的 Helm.command，检查参数与取消，默认使用 helm 二进制，通过 Runner 传递独立 argv，并保留失败原因。Package / Install / Upgrade / Rollback 已给出清晰接口；练习重点是可测试的外部工具调用与真实发布流程。chart/ 是实际的 Deployment + Service Chart。

从 gopath init 安装后的本练习目录运行以下命令。

## 本地反馈

~~~powershell
go test -count=1 ./...
gopath check 10-helm-lifecycle
~~~

本地测试完全不需要 Helm 或集群：fake Runner 检查四种操作的全部参数、命名空间与 context、显式 revision、错误和取消；真实的测试子进程验证 os/exec 能保留 exit code 与 stderr。不要把参数拼成 shell 字符串。Helm wrapper 是标准库 Go 代码，使用 Go 1.27。

## Chart 与 Minikube 实操

需要 Helm 3 或 Helm 4、kubectl、Minikube；准备 Kubernetes v1.37.1，并提前拉取演示镜像。每次 Helm 集群操作均显式传 --kube-context minikube 和 namespace，等待就绪最多 30 秒。

~~~powershell
minikube start --kubernetes-version=v1.37.1
kubectl config use-context minikube
minikube image pull nginx:1.27
helm lint chart
helm template demo chart
go run ./cmd/helm -operation package -chart chart -destination dist
kubectl --context minikube create namespace gopath-helm
go run ./cmd/helm -operation install -release demo -chart dist/gopath-demo-0.1.0.tgz -namespace gopath-helm
kubectl --context minikube -n gopath-helm get deployment,service
go run ./cmd/helm -operation upgrade -release demo -chart dist/gopath-demo-0.1.0.tgz -namespace gopath-helm -replicas 2
helm history demo --kube-context minikube -n gopath-helm
go run ./cmd/helm -operation rollback -release demo -namespace gopath-helm -revision 1
kubectl --context minikube -n gopath-helm get deployment demo
helm history demo --kube-context minikube -n gopath-helm
~~~

应观察副本数 1 → 2 → 1。回滚到 revision 1 会产生新的 revision 3，历史不会被抹去。需要改模板或 values 时，可重新打包 Chart，再升级 release。服务使用 ClusterIP，命名空间内通过 Service 选中相同 app 标签的 Pod。

清理手工 release 与命名空间，再执行真实集成验收：

~~~powershell
helm uninstall demo --kube-context minikube -n gopath-helm
kubectl --context minikube delete namespace gopath-helm
gopath check 10-helm-lifecycle --integration
~~~

TestIntegrationHelmLifecycleOnMinikube 真正 lint、render、package，然后安装打包产物、升级并回滚；每一步读取实际 Deployment 就绪副本，最后检查 revision 3 已部署。整个测试有 100 秒上限；慢速拉镜像时先执行上面的 image pull，再重跑。缺失 Helm、Minikube、镜像或就绪失败会明确报错，不会 skip。

验收仅允许 kubeconfig current-context=minikube，并创建新的 gopath-10-* 命名空间。GOPATH_NAMESPACE 若设置，必须是尚不存在的新 gopath-* 名字。清理前核对命名空间 UID，再删除本测试创建的命名空间；已有命名空间绝不复用。普通 check 仅记录“本地通过，待集成”。

## 能口述的解释

Chart 和 values 如何得到 Kubernetes 资源？Chart 版本与 release revision 有何不同？升级时 Helm 怎样计算期望资源？为什么 rollback 产生新 revision？Helm 的发布历史与 Controller 的持续调谐分别解决什么问题？为什么 exec.CommandContext 使用 argv，而不经 shell 拼接？

参考：[Chart 模板](https://helm.sh/docs/chart_template_guide/)、[helm install](https://helm.sh/docs/helm/helm_install/)、[helm rollback](https://helm.sh/docs/helm/helm_rollback/)、[Minikube 镜像命令](https://minikube.sigs.k8s.io/docs/commands/image/)。
