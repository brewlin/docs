---
title: udp-客户端
toc: true
date: 2020-03-5 21:28:59
tags: [go,protocol,udp,client]
---

## @客户端创建
```go
package main

import (
	"fmt"

	_ "github.com/brewlin/net-protocol/pkg/logging"
	"github.com/brewlin/net-protocol/protocol/transport/udp/client"
)

func main() {
	con := client.NewClient("10.0.2.15", 9000)
	defer con.Close()

	if err := con.Connect(); err != nil {
		fmt.Println(err)
	}
	con.Write([]byte("send msg"))
	res, err := con.Read()
	if err != nil {
		fmt.Println(err)
		con.Close()
		return
	}
	// var p [8]byte
	// res, _ := con.Readn(p[:1])
	fmt.Println(string(res))
}
}

```
- 实现了基本的udp客户端连接读写等函数
- 依赖tap虚拟网卡，所以需要启动网卡依赖
- 依赖`ARP`,`udp`,`IPV4`等协议，所以默认注册了该协议
- 注意：`默认本地地址为192.168.1.0/24 网段，如果目标ip为127.0.0.1 导致无法arp查询物理层地址,请填写局域网物理机器ip,或者外网ip`

## @NewClient 创建客户端
构造函数传入目的ip,端口等参数，默认返回一个*Client 指针
```go
	con := client.NewClient("10.0.2.15", 8080)
```
注意:`默认本地地址为192.168.1.0/24 网段，如果目标ip为127.0.0.1 导致无法arp查询物理层地址`

## @Connect 不进行真正的连接，只处理一些初始化工作

## @Write 写入数据
```go
    con.Write([]byte("send msg"))
```
直接向对端连接写入数据，错误返回err，udp协议直接通过ip数据包像对端发送数据，因为无连接状态，需要等待对方的icmp报文
如果没有收到icmp报文表示发送成功，收到了icmp报文也需要在`read()函数`中才能标识出来

## @Read 读取数据 在这里可以判断对端服务是否正常，因为这里会返回用户层icmp报文情况
一次只读取一次数据，如果缓存没有读取完，则会返回 `ErrWouldBlock`错误，可以 在此监听该读方法
```
	res, err := con.Read()
	if err != nil {
		//这里的错误可能 就会是上面write 写入 对端数据后，对端返回的icmp control msg 表示一些异常情况，如对端端口不可达等
		//如果需要阻塞 进行arp查询等一些操作 会自动进行，这里一般不会出现
		fmt.Println(err)
		con.Close()
		return
	}
```


## @Readn 读取n字节数据
```
	// var p [8]byte
	// res, _ := con.Readn(p[:1])
	// fmt.Println(p)
```
可以根据传入参数填充对应的字节数数据，如果不够则会阻塞等待数据填充满为止

golang 的slice底层是一个指针，所以虽然传值，但是实际会复制指针，那么该slice实际值会在Readn（）函数里被改变填充完后返回

