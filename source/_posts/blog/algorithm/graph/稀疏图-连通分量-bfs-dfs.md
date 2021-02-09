---
title: 稀疏图-连通分量-bfs-dfs
toc: true
date: 2021-02-09 22:50:11
tags: [algorithm,graph]
---
[toc]

针对稀疏图讲解，且结构按[稀疏图-邻接表SparseGraph]()中的结构进行测试

## 全局参数和初始化说明
全局参数
```
SparseGraph graph(5,false);
bool* visited;
//求连通分量
int count = 0
```
1. 构建5个顶点的图，且是无向图
2. 在遍历图的时候需要有个结构来存储每个节点是否已经访问过了`visited`
3. count 用来记录连通分量


初始化
```
visited = new bool[5];
for(int i = 0; i < 5; i ++)
    visited[i] = false;
graph.addEdge(0,1,1);
graph.addEdge(0,3,1);
graph.addEdge(1,1,1);
graph.addEdge(1,3,1);
graph.addEdge(2,4,1);
graph.addEdge(2,1,1);
```
总的来说就是构建一个记录节点是否访问的数组，然后增加一些默认的边

## dfs 深度优先遍历
流程图:
![](/images/blog/graph/YOQPRHWROJ.png)
1. 从0节点开始递归遍历，0只有2一条边，开始遍历2

![](/images/blog/graph/CWDTXGINVD.png)
2. 2对应有`4,3`两条边，先遍历4，4有`1,3`两条边，继续遍历1

![](/images/blog/graph/ZABCEOJUYU.png)
3. 1对应有3 和4两条边，`4已经遍历过了`只遍历3

总的访问顺序:`0-2-4-1-3`,但其实完整的遍历顺序的话会有很多重复访问，只是已经被访问过了就不会再继续访问了

==注意==:如果连通分量只有1的话，最外层循环是不需要的只需要`depth_first_search(0)`即可完成深度优先遍历，因为==从中任何一个顶点都能遍历完整个图==


最外层从每个顶点开始沿着边开始进行递归遍历
```c
for(int i = 0 ; i < graph.points ;i ++)
    if(!visited[i]){
        depth_first_search(i);
    }
```

进行递归遍历，如果已经遍历过了则不需要再次进行遍历
```
void depth_first_search(int v)
{
    if(visited[v]) return;
    //记录当前节点为已经访问过了
    visited[v] = true;
    cout << v << " ";
    //遍历某个点对应的所有的边
    for(auto i : graph.g[v]){
        depth_first_search(i->b);
    }
}
```


## bfs 广度优先遍历

流程图:
![](/images/blog/graph/CMYDNAHBJB.png)
1. 从0开始遍历，将0对应的边`2`加入未访问队列,继续读取2，然后将2对应的边`4,3`加入为访问者队列

![](/images/blog/graph/KBGRAJPYXS.png)
2. 访问`4`，将对应的边`1`加入队列，继续访问3,已经没有未访问的边了，无需在加入队列

![](/images/blog/graph/VGDCUFFMCA.png)
3. 将最后一个1访问后结束遍历


其实和其他广度优先遍历的机制是一样的

实现代码:
```
//广度优先遍历
void breadth_first_search()
{
    //设置一个队列
    queue<int> q;
    q.push(0);
    while(!q.empty()){
        int v = q.front();
        q.pop();
        for(auto i : graph.g[v]){
            if(!visited[i->b]){
                q.push(i->b);
                cout << i->b  <<" ";
                visited[i->b] = true;
            }
        }
    }

}
```

## 连通分量
![](/images/blog/graph/JJFGFVJPKR.png)

如上图是有两个独立的块关系的，这种通过任意某个点无法关联所有电的情况就存在多个连通分量，上图的连通分量为`2`

其实在求连通分量只需要在深度优先遍历上增加一个步骤即可

```
for(int i = 0 ; i < graph.points ;i ++)
    if(!visited[i]){
        count ++;
        depth_first_search(i);
    }
```
1. ==上面特别强调过==:如果连通分量为1（代表任一节点进行深度遍历都可以访问完所有节点），但是在有多个连通分量的图中是不可能访问完所有节点的
2. 所以需要在外面一次对每个顶点都进行一次遍历，如果后面顶点没有被访问过则一定能够证明有多个连通分量存在


例如: 第一次从0开始访问，只能访问到2。那么外层继续访问到1时，发现居然没有被访问过，那么肯定1所在的网状关系是独立与0所在的区域的，连通分量`count ++`.

后面依次对`2,3,4`顶点继续深度优先遍历时发现已经被访问过了，则结束访问，至此连通分量为`2`
