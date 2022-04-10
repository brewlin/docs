---
title: process
toc: true
date: 2020-1-25 21:28:59
tags: [php,swoole,process]
---


## @construct 
```php
use Process\Contract\AbstractProcess;

class DemoProcess extends AbstractProcess
{
    public function __construct()
    {
        $this->name = "process_name";
    }

    public function check(): bool
    {
        return true;
    }

    /**
     * 自定义子进程 执行入口
     * @param Process $process
     */
    public function run(Process $process)
    {

    }
}
```
## @register 注册自定义进程
```php
ProcessManager::register("demo-process",new DemoProcess());
```

## @start 启动
主动伴随swoole进程模型启动，交由swoole mangager进程管理