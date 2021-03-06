---
title: discovery
toc: true
date: 2020-1-31 21:28:59
tags: [php,swoole,discovery]
---
## @env 配置文件
discovery.php
```php
<?php

return [
    'consul' => [
        'address' => env("DISCOVERY_ADDRESS","127.0.0.1"),
        'port'    => env("DISCOVERY_PORT","8500"),
        'register' => [
            'ID'                => '',
            //只注册了grpc 服务，其他都是私有的
            //tcp 和websocket   通过nginx负载均衡即可
            'Name'              => 'grpc-im-cloud-node',
            'Tags'              => [],
            'enableTagOverride'=> false,
            'Address'           => env("APP_HOST","127.0.0.1"),
            'Port'              => (int)env("GRPC_PORT",9500),
            'Check'             => [
                'id'       => '',
                'name'     => '',
                'http'      => "http://127.0.0.1:".env('DISCOVERY_CHECK_PORT',9500)."/health",
                'interval' => "10s",
                'timeout'  => "10s",
            ],
        ],
        'discovery' => [
            'name' => 'grpc-im-logic-node',
            'dc' => 'dc1',
            'near' => '',
            'tag' =>'',
            'passing' => true
        ]
    ],
];
```
## @register 注册服务
```php
$registerStatus = provider()->select()->registerService();
if(!$registerStatus){
    CLog::error("consul register false sleep 1 sec to reregiseter");
    Coroutine::sleep(1);
}
```
## @deregister 注销节点
```php
//注销节点
$discovery = config("discovery");
provider()->select()->deregisterService($discovery['consul']["register"]['Name']);
```