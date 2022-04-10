---
title: db
toc: true
date: 2020-1-29 21:28:59
tags: [php,swoole,database]
---

## @设置结果集为 array
```php
use Core\Event\EventDispatcherInterface;
use Core\Event\EventEnum;
use Core\Event\Mapping\Event;
use Hyperf\Database\Events\StatementPrepared;
use PDO;
/**
 * @Event(alias=EventEnum::DbFetchMode)
 */
class FetchModeEvent implements EventDispatcherInterface
{
    /**
     * @param $event
     */
    public function dispatch(...$param){
        $event = $param[0];
        if ($event instanceof StatementPrepared) {
            $event->statement->setFetchMode(PDO::FETCH_ASSOC);
        }
    }
}
```

## @db 操作
https://hyperf.wiki/#/zh-cn/db/querybuilder