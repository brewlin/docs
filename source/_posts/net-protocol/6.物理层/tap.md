---
title: tool
toc: true
date: 2020-01-10 21:28:59
tags: [go,protocol,eth,tap,tool]
---

### 1.创建一个tap模式的虚拟网卡tap0

```
sudo ip tuntap add mode tap tap0
```

### 2.开启该网卡

```
sudo ip link set tap0 up
```

### 3.设置该网卡的ip及掩码 |添加路由

```
sudo ip route add 192.168.1.0/24 dev tap0
//增加ip地址
sudo ip addr add 192.168.1.1/24 dev tap0

```

### 4.添加网关

```
sudo ip route add default via 192.168.1.2 dev tap0
```

##  @删除网卡
### 1.删除虚拟网卡

```
sudo ip tuntap del mode tap tap0
```


