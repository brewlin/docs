---
title: memory
toc: true
date: 2020-1-26 21:28:59
tags: [php,swoole,memory]
---
## @construct   
使用bean注解自动注入到container中，在swoole启动前就需要申请好内存并初始化，所以需要使用bean注解
```php
use Core\Container\Mapping\Bean;
use Memory\Table;
use Memory\Table\Type;
use Memory\Table\MemoryTable;
/**
 * @Bean()
 */
class CloudClient
{
    public static $table = null;
    /**
     * CloudClient constructor.
     */
    public function __construct()
    {
        $memorySize = (int)env("MEMORY_TABLE",1000);
        $column = [
            "Address" => [Type::String,20],
            "Port"    => [Type::String,10],
        ];
        self::$table = Table::create($memorySize,$column);
    }
}
```

## @Table：：create 创建共享内存
- size  内存大小
- column 数据结构