---
title: icmp协议
toc: false
date: 2019-10-24 21:28:59
tags: [go,protocol,udp]
---
# ICMP 协议

ICMP 的全称是 Internet Control Message Protocol 。与 IP 协议一样同属 TCP/IP 模型中的网络层，并且 ICMP 数据包是包裹在 IP 数据包中的。他的作用是报告一些网络传输过程中的错误与做一些同步工作。ICMP 数据包有许多类型。每一个数据包只有前 4 个字节是相同域的，剩余的字段有不同的数据包类型的不同而不同。ICMP 数据包的格式如下


```
https://tools.ietf.org/html/rfc792

 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     Type      |     Code      |          Checksum             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|                   不同的Type和Code有不同的内容                    |         
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

从技术角度来说，ICMP 就是一个“错误侦测与回报机制”， 其目的就是让我们能够检测网路的连线状况﹐也能确保连线的准确性﹐其功能主要有：

- 侦测远端主机是否存在。
- 建立及维护路由信息。
- 重导数据传送路径（ICMP 重定向）。
- 数据流量控制。
ICMP 在沟通之中，主要是透过不同的类别(Type)与代码(Code) 让机器来识别不同的连线状况。

## 完整类型列表
TYPE |	CODE	|Description
-----| ---------|----------|
0	|0|	Echo Reply——回显应答（Ping 应答）
3	|0|	Network Unreachable——网络不可达
3	|1|	Host Unreachable——主机不可达
3	|2|	Protocol Unreachable——协议不可达
3	|3|	Port Unreachable——端口不可达
3	|4|	Fragmentation needed but no frag. bit set——需要进行分片但设置不分片标志
3	|5|	Source routing failed——源站选路失败
3	|6|	Destination network unknown——目的网络未知
3	|7|	Destination host unknown——目的主机未知
3	|8|	Source host isolated (obsolete)——源主机被隔离（作废不用）
3	|9|	Destination network administratively prohibited——目的网络被强制禁止
3	|10|	Destination host administratively prohibited——目的主机被强制禁止
3	|11|	Network unreachable for TOS——由于服务类型 TOS，网络不可达
3	|12|	Host unreachable for TOS——由于服务类型 TOS，主机不可达
3	|13|	Communication administratively prohibited by filtering——由于过滤，通信被强制禁止
3	|14|	Host precedence violation——主机越权
3	|15|	Precedence cutoff in effect——优先中止生效
4	|0|	Source quench——源端被关闭（基本流控制）
5	|0|	Redirect for network——对网络重定向
5	|1|	Redirect for host——对主机重定向
5	|2|	Redirect for TOS and network——对服务类型和网络重定向
5	|3|	Redirect for TOS and host——对服务类型和主机重定向
8	|0|	Echo request——回显请求（Ping 请求）
9	|0|	Router advertisement——路由器通告
10	|0|	Route solicitation——路由器请求
11	|0|	TTL equals 0 during transit——传输期间生存时间为 0
11	|1|	TTL equals 0 during reassembly——在数据报组装期间生存时间为 0
12	|0|	IP header bad (catchall error)——坏的 IP 首部（包括各种差错）
12	|1|	Required options missing——缺少必需的选项
13	|0|	Timestamp request (obsolete)——时间戳请求（作废不用）
14	| |	Timestamp reply (obsolete)——时间戳应答（作废不用）
15	|0|	Information request (obsolete)——信息请求（作废不用）
16	|0|	Information reply (obsolete)——信息应答（作废不用）
17	|0|	Address mask request——地址掩码请求
18	|0|	Address mask

ICMP 是个非常有用的协议，尤其是当我们要对网路连接状况进行判断的时候。