---
title: raft分布式一致性原理(二)
toc: false
date: 2018-10-26 21:28:59
tags: [go,raft,algorithm]
---
## @1 选举

此阶段为集群初始化，所有的节点都是`FOLLOWER`身份

1. 进行事件循环 

### @fllower 身份
1. 接受主节点心跳
2. 接受投票选举
3. 如果在超时时间里还没有处触发上面两个`事件` 则转换为 `CANDIDATE` 候选身份，作为新的候选人进行选举

### @candidate 身份
1. 开始选举，增加当前任期 term
2. 投票给自己
3. 广播rpc向所有节点发起选举投票
    - 当前任期`小于`其他节点则转换为fllower节点等待心跳到来或者超时
    - 当前依然为candidate 且 收到的投票大于 总结点`1/2` 则准换为 leader节点结束选举
4. 监听广播的投票结果
    - `收到心跳` 说明选举失败退化为follwer节点
    - `选举成功为leader` 结束选举
### @leader 身份
1. 广播所有节点心跳 和日志复制
2. 睡眠等待下一次心跳发送


### 原则

性质  |	描述
-------|------
选举安全原则（Election Safety）	|一个任期（term）内最多允许有一个领导人被选上
领导人只增加原则（Leader Append-Only）	|领导人永远不会覆盖或者删除自己的日志，它只会增加条目
日志匹配原则（Log Matching）	|如果两个日志在相同的索引位置上的日志条目的任期号相同，那么我们就认为这个日志从头到这个索引位置之间的条目完全相同
领导人完全原则（Leader Completeness)	|如果一个日志条目在一个给定任期内被提交，那么这个条目一定会出现在所有任期号更大的领导人中
状态机安全原则（State Machine Safety）	|如果一个服务器已经将给定索引位置的日志条目应用到状态机中，则所有其他服务器不会在该索引位置应用不同的条目

## @2日志复制
只有当该日志同步到所有node才可以进行提交`commited` 到状态机`state machine`

### @状态机
其实就是实际进行操作的区域，如果该服务是数据库，那么状态机就是实际执行命令储存的地方

当收到所有的节点回复可以提交到状态机后，然后leader节点进行提交，提交后在广播到其他节点，通知其他节点可以提交日志到状态机执行命令

到这里就算是成功提交了一条命令