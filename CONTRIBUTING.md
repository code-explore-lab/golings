# 贡献与新增练习

## 目录格式

在 `exercises/<id>/` 下创建：

    exercise.json
    README.md
    starter/go.mod.txt
    starter/exercise.go.txt
    tests/exercise_test.go.txt
    solution/exercise.go.txt

部署资产可放 `starter/chart/...`、`starter/deploy/...`、`starter/cmd/...`。模板以 .txt 结尾，安装时去掉该后缀；starter 和 tests 的内容合并到题目根目录，所以不能有同名目标。README 会复制；solution 和 exercise.json 不复制到学习区。不要把答案放在起始代码、测试提示或部署资产中。

## 元数据

例如：

    {
      "id": "21-example",
      "title": "示例标题",
      "category": "Go 工程",
      "difficulty": "基础",
      "description": "需要解决的具体问题",
      "objectives": ["一个主要概念与可观察结果"],
      "hints": ["第一条方向提示", "第二条更具体提示"],
      "acceptance": ["行为检查", "失败场景检查"],
      "concept": "工程原理的简短解释",
      "check": "go-test",
      "integration": false,
      "prerequisites": ["01-go-errors"]
    }

ID 必须与直属目录同名，只含字母、数字、下划线或连字符。当前只支持 go-test 检查器；不接受任意命令或 shell 脚本。为 ID 使用两位数字前缀维持稳定顺序。字段完整性、Go 语法、依赖 ID 与部署验收入口由课程格式测试检查。

真实环境题设置 `integration:true`，增加 `tests/integration_test.go.txt`，首行 `//go:build integration`，至少一个测试名为 `TestIntegration...`。该测试应真实验证声明的环境结果，缺少环境时明确失败，不能通过 t.Skip 假装通过。普通测试依旧必须离线、确定、能用 fixture/fake/httptest 验证。

## 设计与验证

先定义可观察行为和失败场景，再写测试、完整解答、起始缺陷。起始代码须可编译但行为测试失败；解答必须通过相同必需测试。优先注入 clock、waiter、runner、client 等窄依赖；测试不要依赖随意 sleep、公网和不相关服务。阻塞测试使用通道同步和限时失败保护。并发题用 race 检查，不能仅检查最后结果。

每题说明应包含目标、需修改代码、检查命令、验收标准、原理口述问题和相关官方资料。涉及部署时写明环境、资源作用域、成功信号、错误诊断与清理；Kubernetes 练习限制到 minikube 和测试拥有的 namespace，集群级 CRD 需显式预安装。Helm 题须包含真实 Chart 生命周期而非只断言参数。

提交前执行：

    gofmt -w cmd internal
    go vet ./...
    go test ./...
    go test -tags=curriculum -parallel=2 -timeout=30m -v ./exercises

对 .go.txt 模板也运行 gofmt，例如 PowerShell：

    Get-ChildItem exercises -Recurse -Filter *.go.txt | ForEach-Object { gofmt -w $_.FullName }

课程作者检查显式读取源码中的解答，仅用于验证题库质量；不更新学习者成绩。它检查起始代码编译、初始测试失败、解答通过，以及实操检查可编译；不会真的创建集群资源。需要真实环境的测试另外运行，并如实报告结果。

新增练习通常只涉及自己的目录和 docs/jd-mapping.md，不需改 CLI。嵌入通配符会自动发现目录。保持根模块无外部依赖；题目外部库版本在独立 go.mod 中固定并注明目标环境。不要让课程测试依赖学习者手工复制参考答案。

评审时确认初始化不覆盖、reset 必须明确确认、提示不完成、失败不保留旧通关、零测试/跳过不通过、参考答案不进入二进制。CLI 核心改动增加针对错误状态的行为测试，避免只重复实现逻辑。
