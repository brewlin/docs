---
title: event
toc: true
date: 2020-01-12 21:28:59
tags: [linux,c,php,ext,coroutine,epoll]
---

```php
lib_event_init();

dosthing...


lib_event_wait();
```
## @lib_event_init() 初始化全局变量和申请内存空间
```
LibG;
LibG.poll
LibG.poll.epollfd;
```

## @lib_event_wait() 轮训获取可读可写事件
`timer、socket、server、sleep`等模块依赖于event，所以需要显示调用event
```
while(LibG.running){
    epoll_wait(....)
}
```