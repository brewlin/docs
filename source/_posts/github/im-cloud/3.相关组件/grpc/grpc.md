---
title: grpc
toc: true
date: 2020-1-27 21:28:59
tags: [php,swoole,grpc]
---

## @build
```
> cd pkg/grpc/bin
> sh  gen.sh
.......
```
## @method grpc 调用
组件封装有连接池机制，复用多个连接
```php
use Grpc\Client\GrpcLogicClient;
use Im\Cloud\Operation;
use Im\Logic\HeartbeatReq;

$heartBeatReq = new HeartbeatReq();
$host = env("APP_HOST","127.0.0.1").":".env("GRPC_PORT",9500);
$heartBeatReq->setServer($host);
$heartBeatReq->setKey($key);
$heartBeatReq->setMid($mid);
GrpcLogicClient::Heartbeat($grpcServer,$heartBeatReq);
```