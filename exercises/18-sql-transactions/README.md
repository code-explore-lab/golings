# SQL 转账的原子事务

实现 Transfer(ctx, db, from, to, amount)：amount 使用整数最小货币单位，必须为正；账户 ID 非空且不能相同。参数错误不能调用 BeginTx。

通过同一个 sql.Tx 顺序执行两条语句：
- UPDATE accounts SET balance = balance - ? WHERE id = ? AND balance >= ?，参数 amount、from、amount。
- UPDATE accounts SET balance = balance + ? WHERE id = ?，参数 amount、to。

扣款影响行数为 0 时返回 ErrInsufficientFunds；入账影响行数为 0 时返回 ErrAccountNotFound。只有两次更新均影响一行才提交。任意错误都返回给调用者，并让未提交事务回滚；底层错误保留 errors.Is 身份。

测试包含一个小型 database/sql 驱动，支持上述语句以验证原子性、提交和回滚，无需安装数据库。它不实现 SQL 解析、行锁、隔离级别或持久化；真实数据库的并发事务仍需独立集成测试。

运行 gopath check 18-sql-transactions。口述问题：为何不能先在 db 上扣款，再单独开启入账事务？Commit 失败后为什么不能直接向调用方宣称转账成功？

