---
title: http-客户端
toc: true
date: 2019-12-5 21:28:59
tags: [go,protocol,tcp,client,http,request,response]
---

## @客户端创建
```go
package main
import (
	"fmt"
	"github.com/brewlin/net-protocol/pkg/logging"
	"github.com/brewlin/net-protocol/protocol/application/http"
)
func init()  {
	logging.Setup()
}
func main(){
	cli,err := http.NewClient("http://10.0.2.15:8080/test")
	if err != nil {
		panic(err)
		return
	}
	cli.SetMethod("GET")
	cli.SetData("test")
	res,err := cli.GetResult()
	fmt.Println(res)

}

```
- 实现了基本的http客户端,发起请求和接收响应等get post方法
- 依赖tap虚拟网卡，所以需要启动网卡依赖
- 依赖`ARP`,`TCP`,`IPV4`等协议，所以默认注册了该协议
- 注意：
	- 1.外网请求需要使用`tool/up` 方式启动网卡配置数据包转发
	- 2.未实现dns查询域名，必须使用ip测试

## @NewClient 创建客户端
构造函数传入url,默认返回一个*Client 指针
```go
cli,err := http.NewClient("http://10.0.2.15:8080/test")
```

## @设置请求方法
设置请求的http方法 `GET,POST`等
```go
	cli.SetMethod("GET")
```

## @添加请求数据
```go
	cli.SetData("test")
```
添加数据

## @获取响应结果
该方法真正执行tcp连接，发送数据，和读取响应数据
```go
	res,err := cli.GetResult()
```

