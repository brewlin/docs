---
title: runtime
toc: true
date: 2020-01-13 21:28:59
tags: [linux,c,php,ext,hook,runtime]
---
## `Lib\runtime::enableCorutine` 启动hook机制
```php
<?php
Lib\Runtime::enableCoroutine();

cgo(function(){
    sleep(1); // == Cco::sleep(1);
});
```