---
title: 协程tcp服务
toc: true
date: 2020-01-05 21:28:59
tags: [linux,c,php,ext,coroutine,epoll,socket,tcp]
---

创建协程版server，封装所有协程api，所有阻塞操作都会触发协程切换

## @Lib\Coroutine\Server
```php
//初始化全局对象 epoll等内存空间初始化
lib_event_init();

//协程运行
cgo(function ()
{
        $server = new Lib\Coroutine\Server("127.0.0.1", 9991);
        $server->set_handler(function (Lib\Coroutine\Socket $conn) use($server) {
                    $data = $conn->recv();
                    $responseStr = "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nConnection: close\r\nContent-Length: 11\r\n\r\nhello world\r\n";
                    $conn->send($responseStr);
                    $conn->close();
                    // Sco::sleep(0.01);
        });
        $server->start();
});


//epoll event 轮循 检查事件
lib_event_wait();
```

## @__construct 初始化

## @set_handler($callback)  回调触发事件
## start() 启动监听
