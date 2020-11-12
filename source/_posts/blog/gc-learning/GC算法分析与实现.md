---
title: GC算法分析与实现
toc: true
date: 2020-11-09 21:28:59
tags: [algorithm,gc,c]
---

本系列文章主要阅读`垃圾回收的算法与实现`一书进行的实现解说，主要因为之前没有完整的gc算法实现的代码样例，书中各种实现都是基于伪码讲解，虽能理解作者的意思但难免还是有些抽象

遂写这些文章，记录下自己在学习算法理论并实际实现的过程。后续会追加分析其他语言的`gc实现`，深入理解生产级别是如何应用gc，以及如何极限的优化gc性能

# 目录

## 前言

1. 什么是root根? 
2. 什么是heaps堆?

## 算法实现
1. 标记清除算法 - 基础实现
2. 标记清除算法 - 多空闲链表法
3. 引用计数算法
4. GC  复制算法
5. 复制+标记清除 - 组合实现的多空间复制算法
6. 标记压缩算法 - 基础实现
7. 标记压缩算法 - two_finger实现
8. 保守式gc算法 - `当前都是基于保守式gc算法`
9. 分代垃圾回收 - 复制算法+标记清除组合实现
10. 增量式垃圾回收 - 三色标记


# OS 环境参数
```sh
> gcc -v
Thread model: posix
gcc version 7.5.0 (Ubuntu 7.5.0-3ubuntu1~18.04)

> uname -a
Linux ubuntu 4.4.0-157-generic #185-Ubuntu SMP Tue Jul 23 09:17:01 UTC 2019 x86_64 x86_64 x86_64 GNU/Linux
```
## 关于测试
每个算法实现目录都有`test.c`,都是对当前算法的简单逻辑验证

根目录有一个`auto_test.sh` 脚本可以一次性跑全部的测试用例
```sh
> cd gc-learning
> dos2unix auto_test.sh
> sh auto_test.sh
```

# 代码结构
```sh
gc-learning

----- gc.c
----- gc.h

----- mark-sweep 
----- mark-sweep-mulit-free-list
----- reference-count
----- copying
----- copying-or-mark
----- compact-lisp2
----- compact-two-finger
----- generational
----- tri-color-marking
```
所有的gc算法都依赖于公用`gc.c`中的的`heaps`堆内存池实现，可以先看[什么是堆?](./什么是堆?)了解内存管理

`gc.c`和`gc.h`是公用内存实现
