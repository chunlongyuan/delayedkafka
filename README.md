# DelayedKafka

A kafka and mysql backed queue for scheduling delayed events

✅ 1. 消息支持延迟

消息 id 存于 zset 中, 采用 createTime+delay 作为 score, 轮询到期的固定条数消息进行 delivery

✅ 2. 可靠性

mysql 事务保证消息投递成功

✅ 3. 可恢复

redis 重启后 master 负责同步 mysql 里的数据到 redis

✅ 4. 可撤回

支持通过 id 删除排队中的消息

✅ 5. 可修改

删除后重发

✅ 6. 高可用

多实例 HA/主备模式,  node 抢占 master

### 本机测试步骤 

1. 启动本机测试环境 

```shell
make up
```
2. 创建测试表
```shell
go run main.go synctable
```
3. 启动核心服务
```shell
go run main.go start
```
4. 启动测试 consumer
```shell
go run example/consumer/main.go
```
5. 启动测试 producer
```shell
go run example/producer/main.go -delay 10s
```