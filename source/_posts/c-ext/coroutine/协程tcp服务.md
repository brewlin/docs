---
title: 协程tcp服务
toc: true
date: 2020-01-05 21:28:59
tags: [linux,c,php,ext,coroutine,epoll,socket,tcp]
---

创建协程版server，封装所有协程api，所有阻塞操作都会触发协程切换

## @Lib\Coroutine\Server
```
//初始化全局对象 epoll等内存空间初始化
lib_event_init();

//协程运行
cgo(function(){
    $serv = new Lib\Coroutine\Server("127.0.0.1", 9999);
    while(1){
        $connfd = $serv->accept();
        cgo(function()use($serv,$connfd){
            $msg = $serv->recv($connfd);
            var_dump($msg);
            $serv->send($connfd,$msg);
            $serv->close($connfd);

        });
    }
});

//epoll event 轮循 检查事件
lib_event_wait();
```
