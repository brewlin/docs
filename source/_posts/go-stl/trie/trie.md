---
title: 字典树
toc: true
date: 2018-10-24 21:28:59
tags: [go,algorithm,tree,trie]
---

## 简介
字典树是一种对字符串统计效率非常高的一种算法，当前字典树基于红黑树实现


## @NewTrie
```go
import (
    "github.com/brewlin/go-stl/trie"
)

trie := trie.NewTrie()
```

## @Add 加入字典树
```go
//func (t *Trie) Add(word string) 
trie.Add("test")
```
## @Contains 查询字符串是否存在
```go
//func (t Trie) Contains(word string) bool
flag := trie.Contains("test")//true
```
## @IsPrefix 查询是否是前缀
```go
// func (t Trie) IsPrefix(pre string) bool
trie.Isprefix("te")//test => true
```
