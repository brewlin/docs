---
title: tcp-客户端
toc: true
date: 2019-11-10 21:28:59
tags: [go,protocol,tcp,client]
---

## @客户端创建
```go
import (
	"fmt"

	"github.com/brewlin/net-protocol/pkg/logging"
	"github.com/brewlin/net-protocol/protocol/transport/tcp/client"
	_ "github.com/brewlin/net-protocol/stack/stackinit"
)
func init() {
	logging.Setup()
}
func main() {
    //注意不可以 目标IP为 127.0.0.1 导致无法发送 数据包
	con := client.NewClient("10.0.2.15", 8080)
	if err := con.Connect(); err != nil {
        fmt.Println(err)
        return
	}
    con.Write([]byte("send msg"))
    //阻塞等待读
	res, _ := con.Read()
	fmt.Println(string(res))
}

```
- 实现了基本的tcp客户端连接读写等函数
- 依赖tap虚拟网卡，所以需要启动网卡依赖
- 依赖`ARP`,`TCP`,`IPV4`等协议，所以默认注册了该协议
- 注意：`默认本地地址为192.168.1.0/24 网段，如果目标ip为127.0.0.1 导致无法arp查询物理层地址,请填写局域网物理机器ip,或者外网ip`

## @NewClient 创建客户端
构造函数传入目的ip,端口等参数，默认返回一个*Client 指针
```go
	con := client.NewClient("10.0.2.15", 8080)
```
注意:`默认本地地址为192.168.1.0/24 网段，如果目标ip为127.0.0.1 导致无法arp查询物理层地址`

## @Connect tcp连接握手
该函数主要处理两个任务
- 1.检查tap网卡是否启动，没有则默认初始化启动一个tap网卡拿到`fd`
- 2.进行tcp三次握手
```go
	if err := con.Connect(); err != nil {
        fmt.Println(err)
        return
	}
```
连接失败的情况举例:
1.`err = no remote link address`
- 这种情况一般表示该ip地址的arp查询失败，没有找到对应的mac地址
2.`err = connection was refused`
- 这个和linux socket 错误码一致 表示 对端未监听该端口,连接拒绝

## @Write 写入数据
```go
    con.Write([]byte("send msg"))
```
直接向对端连接写入数据，错误返回err

## @Read 读取数据
一次只读取一次数据，如果缓存没有读取完，则会返回 `ErrWouldBlock`错误，可以 在此监听该读方法
```
    //阻塞等待读
	res, _ := con.Read()
```


## @Readn 读取n字节数据
```
	// var p [8]byte
	// res, _ := con.Readn(p[:1])
	// fmt.Println(p)
```
可以根据传入参数填充对应的字节数数据，如果不够则会阻塞等待数据填充满为止

golang 的slice底层是一个指针，所以虽然传值，但是实际会复制指针，那么该slice实际值会在Readn（）函数里被改变填充完后返回

