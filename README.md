# DelayedKafka

A kafka and mysql backed priority queue for scheduling delayed events

✅ 1. 消息支持延迟

消息 id 存于 zset 中, 采用 createTime+delay 作为 score, 轮询到期的固定条数消息进行 delivery

✅ 2. 可靠性

mysql 事务保证消息投递成功

✅ 3. 可恢复

主服务启动时在 redis 里加锁后同步 mysql 和 redis 里的数据保持一致, 同步期间支持接入新消息, 就算 redis 未曾故障也有可能有内容被驱逐, 因此也需要检查同步

✅ 4. 可撤回

支持通过 id 删除排队中的消息

✅ 5. 可修改

删除后重发

✅ 6. 高可用

多实例 HA/主备模式

主备信息存储于 redis 中（TTL）,实例主动申请成为主节点 