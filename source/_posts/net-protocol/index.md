---
title: index
toc: false
date: 2018-10-24 21:28:59
tags: [go,protocol]
---
# net-protocol
https://github.com/brewlin/net-protocol
基于go 实现链路层、网络层、传输层、应用层 网络协议栈 ，使用虚拟网卡实现
## @demo
```
相关demo以及协议测试在cmd目录下
```
`cd ./cmd/*`
## @application 应用层
- [x] [http](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [websocket](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [dns](http://wiki.brewlin.com/wiki/net-protocol/index/)


## @transport 传输层
- [x] [tcp](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [udp](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [port](http://wiki.brewlin.com/wiki/net-protocol/index/) 端口机制 

## @network 网络层
- [x] [icmp](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [ipv4](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [ipv6](http://wiki.brewlin.com/wiki/net-protocol/index/)

## @link 链路层
- [x] [arp](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [ethernet](http://wiki.brewlin.com/wiki/net-protocol/index/) 

## @物理层
- [x] tun [tap](http://wiki.brewlin.com/wiki/net-protocol/index/) 虚拟网卡的实现

## @客户端
发起客户端请求
- [x] [http client](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [websocket client](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [tcp client](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [udp client](http://wiki.brewlin.com/wiki/net-protocol/index/)
- [x] [dns client](http://wiki.brewlin.com/wiki/net-protocol/index/)
## 协议相关构体

### 1.应用层相关协议
应用层暂时只实现了`http`、`websocket`、`dns`等协议。都基于tcp、对tcp等进行二次封装

http protocol:
```
	http 协议报文
	GET /chat HTTP/1.1
	Host: server.example.com
	Upgrade: websocket
	Connection: Upgrade
	Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
	Origin: http://example.com
	Sec-WebSocket-Protcol: chat, superchat
	Sec-WebSocket-Version: 13
```
websocket protocol:
```
			websocket 数据帧报文

     0               1               2               3               4
     0 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8
     +-+-+-+-+-------+-+-------------+-------------------------------+
     |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
     |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
     |N|V|V|V|       |S|             |   (if payload len==126/127)   |
     | |1|2|3|       |K|             |                               |
     +-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
     |     Extended payload length continued, if payload len == 127  |
     + - - - - - - - - - - - - - - - +-------------------------------+
     |                               |Masking-key, if MASK set to 1  |
     +-------------------------------+-------------------------------+
     | Masking-key (continued)       |          Payload Data         |
     +-------------------------------- - - - - - - - - - - - - - - - +
     :                     Payload Data continued ...                :
     + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
     |                     Payload Data continued ...                |
     +---------------------------------------------------------------+

```
### 2.传输层相关协议
传输层实现了`upd`、`tcp`、灯协议，并实现了主要接口

tcp protocol:

```
		     tcp 首部协议报文
0               1               2               3               4
0 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          Source Port          |       Destination Port        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                        Sequence Number                        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                    Acknowledgment Number                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  Data |           |U|A|P|R|S|F|                               |
| Offset| Reserved  |R|C|S|S|Y|I|            Window             |
|       |           |G|K|H|T|N|N|                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|           Checksum            |         Urgent Pointer        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                    Options                    |    Padding    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                             data                              |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

udp-protocol:
```
udp 协议报文
```


端口机制

### 3.网络层相关协议

ip protocol:
```
              ip头部协议报文
0               1               2               3               4
0 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8 1 2 3 4 5 6 7 8
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|Version|  LHL  | Type of Service |        Total Length         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  Identification(fragment Id)    |Flags|  Fragment Offset      |
|           16 bits               |R|D|M|       13 bits         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| Time-To-Live  |   Protocol      |      Header Checksum        |
| ttl(8 bits)   |    8 bits       |          16 bits            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|               Source IP Address (32 bits)                     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|              Destination Ip Address (32 bits)                 |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                    Options (*** bits)          |  Padding     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```