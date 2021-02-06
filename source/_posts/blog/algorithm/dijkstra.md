---
title: Dijkstra最短路径
toc: true
date: 2021-02-06 17:50:43
tags: [algorithm,graph,golang]
---

dijkstra利用松弛操作找到最短的路线距离，假设当前图结构为稀疏图，结构如下:
![](/images/blog/graph/BGTRJIKUBX.png)

按照直观来说，`0-4`的最短路径有如下几种选择
![](/images/blog/graph/VVKKYYDLXB.png)

最优路显然是`0-2-3-4`,权值只有`7`当属最小

接下来看看如何实现该寻路过程

## 结构说明
```c
SparseGraph graph(5,false);
bool* visited;
int * from;
int * distTo;
```
总共4个额外数组来维护寻路过程的记录

1. `visited`,当对某一节点`left`以及所有对应的边`right`进行访问时对`visited[left] = true`进行标记，表明以及访问过当前节点了
2. `from`,这个作为辅助数据，在计算到最短路径后，如果需要打印完整的路径，则需要from来记录每次寻路的过程
3. `distTo`，`distTo[i]`记录了原点`s`到`i`的最短距离

## bfs进行寻路
```c
void breadth_first_search(int s)
{
    //采用最小索引堆来做,默认初始化n个顶点空间
    IndexMinHeap<int> qp(graph.points);
    //默认插入一个 源起始点
    qp.insert(s,0);
    while(!qp.isEmpty()){
        //每次获取s原点最短的那个距离
        int left = qp.extraMinIndex();
        //标记该节点已经被访问过了
        visited[left] = true;
        //接下来访问该节点的所有邻边
        for(auto edge : graph.g[left]){
            //查看对应的邻边有没有被访问过,edge->a 就是当前id, edge->b才是领边 edge->v 代表权值
            int right  = edge->b;
            int length = edge->v; 
            if(!visited[right]){
                //判断from 路径有没有记录 || 如果[s -> left + left->right] < [s -> right] 说明找到了更短的距离
                if(from[right] == -1 || distTo[left] + length < distTo[right]){
                    //更新当前被访问的right节点的来源节点left
                    from[right] = left;
                    //更新距离: s->right = s->left + left->right
                    distTo[right] = distTo[left] + length;
                    //判断队列里有没有访问过当前的 right节点
                    if(qp.contain(right)) qp.change(right,distTo[right]);
                    else qp.insert(right,distTo[right]);
                    
                }
            }
        }

    }

}
```
1. 从s原点开始,加入到最小索引堆中`qp.insert(0,0)`,因为`0-0`的距离默认为0，所以第一个节点默认权值为0 
2. 接下来就是对`0`点的各个边进行扫描，如果发现有更短的距离，则直接更新新的距离
```
如上图所示:
0 -> 4 : 距离为9  ，那么 distTo[4] = 9;

0->2->4 : 因为 0->2 = 2, 2->4 = 6 , 存在 0-2-4(8) < 0-4(9)

所以找到了更优的路径：
o->4 = 8,

依次类推，实现最短路径的查找
```
3. 不断重复步骤2，直到遍历完可达的顶点后，结束寻路计算
4. 最后的结果应该如下模式
![](/images/blog/graph/XWBXUNRBCW.png)

## 路径展示
检查是否通路
```c
//判断0-4是否有路
if(distTo[4]){
    cout << " 1-4 有路:" << distTo[4] << endl;
}else{
    cout << " 2-4 有路" <<endl;
}

```

路径展示
```c
int i = 4;
cout << "4 " ;
while(i != -1 ){
    cout << "->";
    cout << from[i];
    i = from[i];
}

   
```

## 完整的寻路流程图

![](/images/blog/graph/QYGEZQYLNK.png)

## 完整的代码
```c
#include "SparseGraph.h"
#include "IndexMinHeap.h"
#include <stack>
#include <queue>

SparseGraph graph(5,false);

//代表是否访问过
bool* visited;
//记录路径，from[i] 表示查找的路径i的上一个节点
int * from;
//记录权值 这里的权值采用int来表示
int * distTo;
//bfs 广度优先遍历
void breadth_first_search(int s)
{
    //采用最小索引堆来做,默认初始化n个顶点空间
    IndexMinHeap<int> qp(graph.points);
    //默认插入一个 源起始点
    qp.insert(s,0);

    while(!qp.isEmpty()){
        //每次获取s原点最短的那个距离
        int left = qp.extraMinIndex();
        //标记该节点已经被访问过了
        visited[left] = true;
        //接下来访问该节点的所有邻边
        for(auto edge : graph.g[left]){
            //查看对应的邻边有没有被访问过,edge->a 就是当前id, edge->b才是领边 edge->v 代表权值
            int right  = edge->b;
            int length = edge->v; 

            if(!visited[right]){
                //判断from 路径有没有记录 || 如果[s -> left + left->right] < [s -> right] 
                if(from[right] == -1 || distTo[left] + length < distTo[right]){
                    //更新当前被访问的right节点的来源节点left
                    from[right] = left;

                    //更新距离: s->right = s->left + left->right
                    distTo[right] = distTo[left] + length;

                    //判断队列里有没有访问过当前的 right节点
                    if(qp.contain(right)){
                        qp.change(right,distTo[right]);
                    }else{
                        qp.insert(right,distTo[right]);
                    }
                }
            }
        }

    }

}
int main(){
    visited = new bool[5];
    from    = new int[5];
    distTo  = new int[5];
    for(int i = 0; i < 5; i ++){
        from[i] = -1;
        visited[i] = false;
        //disTo[i] 记录了原点s到i的最短距离
        distTo[i] = 0;
    }
    graph.addEdge(0,1,1);
    graph.addEdge(0,4,2);
    graph.addEdge(1,4,6);
    graph.addEdge(1,3,3);
    graph.addEdge(3,4,4);

    //标记从1开始到其他任意节点的最短路径
    int p = 1;
    breadth_first_search(p);

    //判断1-4是否有路
    if(distTo[4]){
        cout << " 1-4 有路:" << distTo[4] << endl;
    }else{
        cout << " 2-4 有路" <<endl;
    }


    // //查看 2-3有没有路径
    // if(!visited[3]){
    //     cout << "不存在" <<endl;
    // }
    // //查看路径
    int i = 4;
    cout << "4 " ;
    while(i != -1 ){
        cout << "->";
        cout << from[i];
        i = from[i];
    }

   

}
```
