---
title: task
toc: true
date: 2020-1-21 21:28:59
tags: [php,swoole,coroutine,task]
---

## @composer
```php
{
  "require":{
    "brewlin/im-task"
  }
}
```

## @class
```php
class Task {
}
```

## @diliver 发送异步任务
通过`deliver`方法可以直接在task进程中执行对应object的方法到达异步执行任务
```php
use namespace example;

/** @var Task $task */
\bean(Task::class)->deliver(example::class,"method",[arg1,arg2,arg3.....]);
```
