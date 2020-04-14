---
title: c扩展
toc: true
date: 2018-12-15 21:28:59
tags: [c,c++,php,ext]
---


## 基于c ，c ++ 封装php对象扩展

## @Lib
所有的扩展对象均以`Lib` 命名空间为开头
```php
namespace Lib;
```
## @env
```
supoort:
> php7.3.5
> php7.3.10
```
项目中对于协程相关底层实现参考 `https://github.com/php-extension-research/study.git` 实现，并在此之上做了一些重构，详情请关注原协程实现
## @`cgo()`
创建一个协程运行
## @`Lib/SharMem`
该扩展申请一块共享内存地址，提供php调用，用于多进程间共享数据
## @`Lib/Process`
该扩展初始化传入回调函数并创建子进程执行，子进程间可以通过channel通讯
## @`Lib/Timer`
提供定时任务和对于timer操作，基于epoll阻塞实现定时器，采用链表保存时间任务，有待提高性能
## @`Lib/Coroutine/Server`
提供携程化socket服务，监听tcp协议
## @`Lib/Coroutine/Socket`
## @`event`
显示声明异步事件模式
## @`Hook`
对原生php底层函数进行hook替换
## @`Lib/Thread/Pool` 
线程池，可以在任何地方创建并投递后台任务进行处理
## @`Lib/Thread/Pool/Future` 
投递任务给线程池处理后，可以阻塞获取对应的返回值
