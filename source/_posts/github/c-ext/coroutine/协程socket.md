---
title: 协程socket
toc: true
date: 2020-01-12 21:28:59
tags: [linux,c,php,ext,coroutine,epoll,socket,tcp]
---

创建协程版socket，封装所有协程api，所有阻塞操作都会触发协程切换

## @Lib\Coroutine\Socket
```php
<?php
lib_event_init();

cgo(function ()
{
    $socket = new Lib\Coroutine\Socket(AF_INET, SOCK_STREAM, IPPROTO_IP);
    if($socket->fd < 0){
        var_dump("err");return;
    }
    $socket->bind("127.0.0.1", 9999);
    $socket->listen();
    while (true) {
        $conn = $socket->accept();
        cgo(function () use($conn)
        {
            $data = $conn->recv();
            $responseStr = "HTTP/1.1 200 OK\r\n
                        Content-Type: text/html\r\n
                        Connection: close\r\n
                        Content-Length: 11\r\n\r\n
                        hello world\r\n";
            $conn->send($responseStr);
            $conn->close();
        });
    }
});

lib_event_wait();
```
## @Constant
```
AF_INET, SOCK_STREAM, IPPROTO_IP
```
## @__construct 初始化

## @bind  绑定端口
## @listen 启动监听
## @accept 接收新连接
## @recv   读取缓冲区数据
## @send   向对端连接写入数据
## @close  关闭连接
