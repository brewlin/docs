---
title: Logic 节点开放接口
toc: false
date: 2019-11-24 21:28:59
tags: [php,swoole,corotinue,grpc,http]
--

## @http api接口
开放api推送接口,暴露http接口，提供用户业务推送功能
```php
HttpRouter::post("/im/push/keys","/Api/PushKeyController/keys");
HttpRouter::post("/im/push/mids","/Api/PushMidController/mids");
HttpRouter::post("/im/push/room","/Api/PushRoomController/room");
HttpRouter::post("/im/push/all","/Api/PushAllController/all");
```
## @grpc 接口
提供cloud节点用户连接注册grpc接口，多节点可以采用负载均衡
```
HttpRouter::post('/im.logic.Logic/Connect', '/Grpc/Logic/connect');
HttpRouter::post('/im.logic.Logic/Disconnect', '/Grpc/Logic/disConnect');
HttpRouter::post('/im.logic.Logic/Heartbeat', '/Grpc/Logic/heartBeat');
```