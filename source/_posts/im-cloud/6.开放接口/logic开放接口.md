---
title: Logic 节点开放接口
toc: false
date: 2019-11-24 21:28:59
tags: [php,swoole,corotinue,grpc,http]
---

## http api接口
开放api推送接口,暴露http接口，提供用户业务推送功能
```php
HttpRouter::post("/im/push/keys","/Api/PushKeyController/keys");
HttpRouter::post("/im/push/mids","/Api/PushMidController/mids");
HttpRouter::post("/im/push/room","/Api/PushRoomController/room");
HttpRouter::post("/im/push/all","/Api/PushAllController/all");
```
### @push/keys
根据`keys` 为id推送消息，key为client在注册cloud节点时分配的唯一值，每个端点（pc,android,ios）等注册等连接key值不同，所以会进行全部推送

[post] request:
```
{
    "keys":[x,x,x,x,],
    "msg":"bytes"
}
```
### @push/mids
根据`mids` 为id推送对应client消息，mid为业务方自行管理，同一用户每个端点等mid唯一

[post] request:
```
{
    "mids":[x,x,x,x,],
    "msg":"bytes"
}
```
### @push/room
进行房间广播，`type` + `room` 组合为房间唯一id
[post] request:
```
{
    "type":"product1",
    "room":"room1",
    "msg":"bytes"
}
```
### @push/all
广播消息，推送所有端点所有连接

[post] request:
```
{
    "msg":"bytes"
}
```

## grpc 接口
提供cloud节点用户连接注册grpc接口，多节点可以采用负载均衡
```
HttpRouter::post('/im.logic.Logic/Connect', '/Grpc/Logic/connect');
HttpRouter::post('/im.logic.Logic/Disconnect', '/Grpc/Logic/disConnect');
HttpRouter::post('/im.logic.Logic/Heartbeat', '/Grpc/Logic/heartBeat');
```