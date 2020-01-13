---
title: sleep
toc: true
date: 2020-01-13 21:28:59
tags: [linux,c,php,ext,hook,sleep]
---
在扩展内替换原生php内置sleep函数，使原有基于sleep的代码自动进行替换为协程`Cco::sleep()`调用
## @协程sleep
```php
cgo(function(){
    Cco::sleep(1);//协程切换
});
```
## @原生sleep
```php
cgo(function(){
    sleep(1);//进程阻塞
});
```
## @hook sleep
```php
cgo(function(){
    sleep(1);//协程切换
});
```
