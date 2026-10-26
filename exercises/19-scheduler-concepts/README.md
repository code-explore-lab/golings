# 调度器基础概念模型

本题只做基础认知：Pod 资源请求与 nodeSelector 决定哪些节点可行；评分从可行集合中选择一个节点。Node.Available 表示已经扣除其他负载后的可用资源，单位分别为毫核和 MiB。

实现 EligibleNodes：要求 CPU 和内存都足够、所有 Selector 标签匹配。负资源请求或负可用资源不合法，不应进入结果。保持输入顺序，不修改输入。

实现 ChooseNode：从可行节点选择分配后剩余 CPU 最大者；并列按节点名称字典序最小者选择；无可行节点返回空字符串和 false。

这是一个刻意简化的策略示例，不复制 kube-scheduler 的完整行为。完成后能口述 Filter、Score、Bind 的区别，以及资源不足和标签不匹配为何会导致 Pending 即可。无需实现 GPU 调度、抢占、拓扑、复杂评分或阅读调度器源码。

运行 gopath check 19-scheduler-concepts。

