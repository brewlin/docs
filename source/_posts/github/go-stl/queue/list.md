---
title: LIST
toc: true
date: 2018-10-24 21:28:59
tags: [go,algorithm,list,queue]
---

## 简介
基于链表的队列

push:O(1)

Pop:O(1)

入队和出队都可以达到O（1） 的时间复杂度。因为链表维护了头结点和尾节点从而使入队和出队都是O（1）

## api
### @push
```go
queue := NewQueue()
queue.Push(interface)
```

### @Pop
```go
queue.Pop()
```

### @IsEmpty
```go
queue.IsEmpty()
``` 