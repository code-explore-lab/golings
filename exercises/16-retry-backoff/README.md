# 可取消的指数退避重试

实现 Retry：最多调用 operation 共 attempts 次；第一次立即执行，失败后按 baseDelay、2×baseDelay 等待；成功立即返回。最后一次失败不再等待，返回该错误。超过 time.Duration 上限时饱和到最大值，不能溢出成负数。

attempts 必须大于零，baseDelay 不能为负，operation 和 wait 必须非 nil，否则返回参数错误。每次操作前检查 context；wait 收到原 context，等待失败立即返回。操作错误应保持 errors.Is 可识别。

先实现起始文件中的 TODO，再执行 gopath check 16-retry-backoff。测试通过注入 wait 记录时长，不会访问网络或真实休眠。

口述问题：为什么重试必须有次数上限？哪些非幂等操作不能无条件重试？实际系统中如何加入抖动和错误分类？本题不要求实现这些扩展。

