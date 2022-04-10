---
title: queue
toc: true
date: 2020-1-28 21:28:59
tags: [php,swoole,queue]
---
## @producer 生产数据
```php
use App\Lib\Producer;
/** @var Producer $producers */
$producers = \bean(Producer::class);
//发送到队列里
producer()->produce($producers->producer($pushmsg));
```
## @consumer 消费队列数据
```php
use ImQueue\Amqp\Message\ConsumerMessage;
use ImQueue\Amqp\Result;
/**
 * Class Consumer
 * @package App\Lib
 */
class Consumer extends ConsumerMessage
{
    public function __construct()
    {
        $this->setExchange(env("EXCHANGE"));
        $this->setQueue(env("QUEUE"));
        $this->setRoutingKey(env("ROUTE_KEY"));
    }

    /**
     * 主流程消费数据入口
     * @param PushMsg $data
     * @return string
     */
    public function consume($data): string
    {
        return Result::ACK;
    }

    /**
     * 重新排队
     * @return bool
     */
    public function isRequeue(): bool
    {
        return true;
    }

}
```