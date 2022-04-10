---
title: php_offset
toc: true
date: 2020-04-10 21:28:59
tags: [php,linux,c]
---
# @container_of 定义
在看`linux_os_link.c`内核链表的时候，看到的一个高级技巧，`通过结构体偏移量`定位实际对象的指针地址

定义如下：
```c
#define container_of(ptr, type, member) ({			\
        const typeof( ((type *)0)->member ) *__mptr = (ptr);	\
        (type *)( (char *)__mptr - __offsetof(type,member) );})
```
总的来说`ptr`一个`type`对象里面的`member`成员指针，现在如果你只有`member`成员的指针，但是`你想拿到type对象的地址`那么container_of就发挥了重要作用，如下图所示:
![image](/images/blog/linuxos/linux_link_os.png)
整个链表通过node指针串联起来，所以能够想到，当我们通过`*node`指针遍历所有的节点时，`我们怎么获取到整个对象的地址呢`答案当然是上面提到的`container_of`技巧:
```c
//假如们已经遍历到了第一个node节点
link_node *node;

//现在我们想获取 struct test对象指针则可以这样
struct test *obj = container_of(node,struct test,node);

//现在obj 就是模板对象的指针了，是不是很方便呢，
```
当然这种技巧主要还是为了节省内存，你也可以在node结构体中加入一个自定义的结构体指针指向`struct test`即可，就不用通过偏移量定位了

通过`container_of`显然可以节省一个指针内存的空间了，这在很多高性能场景必然发挥了重要作用


## container_of 解析（一）
我们先来看第一行
```
#define container_of(ptr, type, member) ({			\
        const typeof( ((type *)0)->member ) *__mptr = (ptr);	\
```
1. `((type *)0)->memeber` 通过将`0X00`地址转换为`type`自定义类型，再访问对应的member成员
2. `typeof` 编译期间获取`member`成员类型，其实就是获取`node`节点的结构体类型
3. `const node *_mptr = (ptr)` 首先ptr是链表的node节点指针，这行代码主要就是单独定义一个node指针执行`ptr`而已

## container_of 解析 (二)
这行代码是实际的偏移量计算代码
```
(type *)( (char *)__mptr - __offsetof(type,member) );})
```
1. `(char *)_mptr` 我们知道指针的运算受限于`指针类型`，如果指针类型为`int`那么`对int*指针 +1，则地址可能位移了4个字节`,所以强制转换为`char *` 保证更加精确
2. `_offsetof(type,member)` 计算出成员member相对于结构体对象的内存偏移量
3. `mptr - offsetof(type,membr)` 
```
//加入有一个结构体如下
typedef struct{
    char a,
    char b,
    link_node *node;
}test;

//mptr 就是node指针
//那么有如下计算
offsetof(test,node) = 2;
//因为node之前有两个字节，所以node相对于test结构体的偏移量为2

所以mptr-offsetof(test,node)  = test结构体的指针地址
```


# @offsetof 定义
offsetof可以用于计算某个成员相对于结构体对应的偏移量，这样当我们能拿到`任意成员地址时`，都能获取到结构体对象地址
```
#define __offsetof(TYPE, MEMBER) ((size_t) &((TYPE *)0)->MEMBER
```
1. `( (TYPE *)0 `) 将零转型为TYPE类型指针;
2. `((TYPE *)0)->MEMBER` 访问结构中的数据成员;
3. `&( ( (TYPE *)0 )->MEMBER )`取出数据成员的地址;
4. `(size_t)(&(((TYPE*)0)->MEMBER))`结果转换类型.巧妙之处在于将`0转换成(TYPE*)`，结构以内存空间首地址`0作为起始地址`，则成员地址自然为`偏移地址`；


# php扩展中的技巧场景
在通过c++开发对应php扩展`class`时，会有这样的场景，对应php扩展类实例化的时候通常对应一个`c++类`，那么就会存在`php-class`对应一个`c++-class`关系

那么他们怎么关联的呢？可能最容易想到的是
```
zend_declare_property_string(lib_co_server_ce_ptr, ZEND_STRL("obj"), "", ZEND_ACC_PRIVATE)

Test *test = new Test();
zend_update_property_string(lib_co_server_ce_ptr, getThis(), ZEND_STRL("obj"), Z_VAL_P(test));
```
总的来说就是在php属性中增加一个私有成员变量，将实例化的c++对象赋值给php成员变量

这种做法总的来说是灾难的，php内核不保证会做什么其他操作，非常不安全，还有就是每次访问对应的`c++对象`都需要进行读取操作，非常不友好

## 通过偏移量来绑定对应对象
这种方式也是官方推荐的方式，健全、安全、且友好

首先定义主体结构体
```
typedef struct
{
    Server *serv;
    zend_object std;
}serv
```
可以看出 std成员就是php对象实际指针，serv成员就是c++对象指针

定义对象生成事件
```
static serv* lib_server_fetch_object(zend_object *obj)
{
    return (serv *)((char *)obj - lib_server_handlers.offset);
}

static zend_object* lib_server_create_object(zend_class_entry *ce)
{
    serv *serv_t = (serv *)ecalloc(1, sizeof(serv) + zend_object_properties_size(ce));
    zend_object_std_init(&serv_t->std, ce);
    object_properties_init(&serv_t->std, ce);
    serv_t->std.handlers = &lib_server_handlers;
    return &serv_t->std;
}
```
1. 在php层面`new serv()`时，会调用`lib_server_create_object`函数，且函数内部我们`不是直接去创建一个zend_object返回`,而是创建一个`serv`
2. 当我们想要获取c++对象时会调用`fetch_object`函数传入php对象指针`obj`其实就是上面的那个`zend_object std`，所以根据上面的技巧我们显然可以通过偏移量来获得c++指针的地址
3. 结构体地址 也可以当做是第一个成员的地址`这是c语言`内存布局的特性，所以通过这个技巧就可以巧妙绑定c++对象以及php对象指针