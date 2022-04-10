---
title: websocket-客户端
toc: true
date: 2019-12-6 21:28:59
tags: [go,protocol,tcp,client,http,websocket,request,response]
---

## @客户端创建
```go
package main

import (
	"fmt"
	"github.com/brewlin/net-protocol/pkg/logging"
	"github.com/brewlin/net-protocol/protocol/application/websocket"
)

func init()  {
	logging.Setup()

}
func main(){
	wscli ,_ := websocket.NewClient("http://10.0.2.15:8080/ws")
	defer wscli.Close()
	//升级 http协议为websocket
	if err := wscli.Upgrade();err != nil {
		panic(err)
	}
	//循环接受数据
	for {
		if err := wscli.Push("test");err != nil {
			break
		}
		data,_ := wscli.Recv()
		fmt.Println(data)
	}
}

```
- 实现了基本的websocket客户端,升级http协议，发送数据，接受数据等方法
- 依赖tap虚拟网卡，所以需要启动网卡依赖
- 依赖`ARP`,`TCP`,`IPV4`等协议，所以默认注册了该协议
- 注意：
	- 1.外网请求需要使用`tool/up` 方式启动网卡配置数据包转发
	- 2.未实现dns查询域名，必须使用ip测试

## @NewClient 创建客户端
构造函数传入url,默认返回一个*Client 指针
```go
cli,err := http.NewClient("http://10.0.2.15:8080/ws")
```
## @Close 关闭连接
结束后，需要手动关闭连接，底层进行tcp四次挥手结束两端状态
```go
	defer wscli.Close()
```
## @Upgrade 升级协议
该方法主要执行两个步骤
- 1. 发起http情况，告诉服务端为websocket协议
- 2. 对服务端返回的http响应，进行校验，校验通过后保持tcp连接，升级为websocket协议
```go
	//升级 http协议为websocket
	if err := wscli.Upgrade();err != nil {
```

## @push 推送数据
```go
 wscli.Push("test")
```
添加数据

## @Recv 获取数据
读取该websocket流 接受的数据，本质为tcp流数据，经过websocket协议解包后处理
```go
	data,_ := wscli.Recv()
```

