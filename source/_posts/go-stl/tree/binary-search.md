---
title: 二分搜索树
toc: true
date: 2018-12-08 21:28:59
tags: [go,algorithm,tree,binary-tree]
---

## 简介
二分搜索树,基础版的实现，用于展示，因为在多种情况下效率比较差，没有自平衡，会退化成链表等,建议使用红黑树，或者hash表等，

## @NewBSTree 
```go
//NewBSTree return BSTree
func NewBSTree() *BSTree
```

## @Size 获取大小
```go
//Size get size
func (b BSTree) Size() int 
```
## @Add 添加节点
```go
//Add add
func (b *BSTree) Add(e int) 
```
## @Contains 检查是否存在
```go
func (b BSTree) Contains(e int) bool 
```
## @PreOrder 前中后序遍历
```
func (b BSTree) PreOrder() 
```
## @LevelOrder 层序遍历
```
func (b BSTree) LevelOrder() 
```
## @Min 最小节点
```
//mini get mini value
func (c BSTree) Min() int 
```
## @Max 最大节点
```
func (c BSTree) Max() int
```
## @RemoveMin 去除最小的
```
func (c *BSTree) RemoveMin() int
```
## @RemoveMax 去除最大节点
```
func (c *BSTree) RemoveMax() int
```
## @Remove 去除节点
```
func (c *BSTree) Remove(e int)
```
