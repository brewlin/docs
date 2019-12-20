---
title: c扩展
toc: true
date: 2017-12-15 21:28:59
tags: [c,c++,php,extension]
---


## 基于c ，c ++ 封装php对象扩展

## @Lib
所有的扩展对象均以`Lib` 命名空间为开头
```php
namespace Lib;
```

## @`Lib/SharMem`
该扩展申请一块共享内存地址，提供php调用，用于多进程间共享数据

## @`Lib/Process`
该扩展初始化传入回调函数并创建子进程执行，子进程间可以通过channel通讯
