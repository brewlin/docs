---
title: Goos-底层协程实现(一)
toc: true
date: 2020-08-10 21:28:59
tags: [linux,c,php,ext,thread,coroutine,scheduler]
---
本节主要讲解什么是协程、协程的实现、php的协程封装的内容，从为什么我们需要协程到如何实现协程，主要讲解协程、php、c、汇编指令的相关关系，从整体窥探它的整个结构
- [Goos-多线程协程实现简要]()
- Goos-协程底层实现(一)
- [Goos-线程协程隔离(二)]()
- [Goos-线程切换实现(三)]()
- [Goos-抢占调度实现(四)]()
- [Goos-监控线程实现(五)]()

# 协程本质
协程最直观的就是我们将一个闭包函数当做`参数`丢给了某个任务去执行，那么实际执行的其实就是我们自定义的函数，如:
```php
<?php
Runtime::GOMAXPROCS(10);
function task(){
    echo "start doing sth";
}
go(task);

Runtime::wait();
```
可以看到我们将`task`函数交给`go`去执行，某些情况下和我们直接`task()`调用无任何区别，那么我们为什么还要通过`go`来调用呢，因为我们想要更好的控制该函数的生命周期，试下一下如下场景:

1. 网络等待导致当前进程阻塞与网络调用
```php
function task()
{
    //maybe 10s+ waiting
    $data = scoket_read(fd);
    //then do sting
}
```
2. 业务逻辑死循环导致进程挂起
```php
function task()
{
    for(;;){
        if(sth is true)
            then break loop;
    }
}
```
3. 单纯是需要利用多核cpu，且不想采用多进程的方式

如果我们直接就调用执行了`task`。上面的三种情况都会导致性能杀手或者进程卡死，有没有一种方法可以控制函数的执行并且能无需受到到所写的业务代码还要担心阻塞等心智负担呢。有没有办法能够充分利用多核实现并行执行呢，所以多线程协程才有了意义。

1. 针对第一种网络阻塞等三方接口调用导致的阻塞，将该代码丢入协程调度器去执行，那么发生阻塞的时候会自动跳过当前函数，继续执行其他任务，完美解决当前问题。当然我们还需要一个契机去恢复上一次函数的继续执行，这就是后续要实现的`poller`网络轮训器来作为调度过程的一部分，当网络事件到来则恢复刚才暂停的函数继续去执行
2. 如果某个函数长期占有cpu，导致其他函数得不到执行，这种情况就可以发起抢占，将当前函数从调度器中移除，继续执行其他的任务，很好的解决了进程卡死和效率低的问题
3. 当然对于多线程来说本身就是可以利用多核cpu的，这样就更好的控制了并发


继续回到协程本质的话题，协程本质就是可以通过调度器来管理一个用户自定义的函数，且该自定义函数被执行的期间的任务可以称为协程，和直接调用函数的区别在于协程的整个期间可以由内由外来进行控制，

# 函数本质
我们通过函数来将我们的业务逻辑划分为多个子集，为了更好的管理工程和设计，我们可以拿函数来作为例子讲解一下实际的执行过程

## php函数的实现
引用这里的文档:https://www.kancloud.cn/lifei6671/php-kernel/675135,来简单分析下函数的本质

php函数实际对应于c语言的`zend_function`结构体:
```c
typedef union  _zend_function        zend_function;

//zend_compile.h
union _zend_function {
    zend_uchar type;    /* MUST be the first element of this struct! */

    struct {
        zend_uchar type;  /* never used */
        zend_uchar arg_flags[3]; /* bitset of arg_info.pass_by_reference */
        uint32_t fn_flags;
        zend_string *function_name;
        zend_class_entry *scope; //成员方法所属类，面向对象实现中用到
        union _zend_function *prototype;
        uint32_t num_args; //参数数量
        uint32_t required_num_args; //必传参数数量
        zend_arg_info *arg_info; //参数信息
    } common;

    zend_op_array op_array; //函数实际编译为普通的zend_op_array
    zend_internal_function internal_function;
};
```
php实际有两种函数，一种是普通函数，另外一只是对象成员函数。 成员函数和普通函数的区别在于,底层zend_function指针内部的scope 会指向一个对象，普通函数则为`NULL`，成员函数则会指向当前的`zend_class_entry`对象指针来实现`this`功能

对于php函数还有两个区别,用户自定义函数和内部函数,虽然所有的函数都被包装成为了`zend_function`，但`zend_function`是一个联合体，所以不同类型的函数在结构上还是有区别的

1. php的内部函数、动态扩展提供的c函数等，这些都是直接存储了一个函数指针给php层面调用即可，即`zend_function->internal_function`指向的是c层面的函数指针，无需其他初始化操作
2. php用户自定义函数，这个时候就有点复杂了，这个层面是zend引擎通过词法、语法分析等将php代码翻译为opcode码、基本就是汇编代码，直接装到`op_array`中，在发生函数调用是,会将`op_array`载入全局execute_globals执行引擎，等待执行opcode码

## c函数的实现
c语言函数就显得非常纯粹了，完全是按照cpu的执行方式来进行思考的，需要完整的考虑内存如:堆、栈等信息，在c层面我们就能想到很多问题，那么我们来讲讲什么是堆？什么是栈?

对于cpu来说内部有多个寄存器，同一时间只能存储一个值，所以显然是不够的，我们的程序拥有无比复杂的变量定义和逻辑运算，例如x64位cpu有16个通用寄存器:
```
通用: %rax %rbx %rdx %esi %edi %rbp %rsp %8-%15 
栈段: %ss  %sp
码段: %cs %ip
数据段: %ds,%es
```
即使这样依然是不够的，我们需要一种比较持久的方法来存储我们的变量以及相关函数地址。那就是栈,那怎么标识一个栈的位置呢，比如栈的起始位置和结束位置:
```
cpu中有两个关键的寄存器用于标识栈的信息，ss:sp:bp 等段基础寄存器
ss: 指向栈段的顶点边界
sp: 指向的是栈底边界
bp: 一般在函数开始的时候，指向当前函数的栈底，sp=bp 然后对于栈上的变量都是基于bp+偏移量访问的
```

### 问题: 同时多个实例的内存是怎么区分的
每个程序在编译为机器码后，对应的ss，sp段寄存器的地址都是一样的，在编译期间就计算了，例如
![image](/images/blog/goos/1.png)

这个是一个win 16的debug.exe，可以看到每个cpu指令执行期间的每个寄存器的值，当你的程序被启动多次，也就是产生了多个进程时，对于cpu来说执行的指令没有任何区别，包括上面`ss`,`sp`对于的栈地址也是一样

那就产生一个疑问，这样的话多进程下岂不是变量共享了，其实到这里就需要引申一个`虚拟地址`的问题，其实我们的运行的程序所有变量的地址都是虚拟地址，在实际访问时，由操作系统转为实际的物理地址

这就是为什么你针对多个进程debug，查看同一个变量地址时都是相同的，但是实际所指的地址却不是同一个东西的原因
### 函数栈的形成
```
SP 栈顶: 0x10000
1. 这里是main函数        +--------------+ main函数起始地址
                         |              |
                    +    |              | 这里是本地的变量存储区域
                    |    +--------------+
                    |    |              |
                    |    |   arg(N-1)   | 这里起始就开始准备调用函数了
                    |    |              |
                    |    +--------------+
                    |    |              |
                    |    |     argN     |
                    |    |              |
                    |    +--------------+
                    |    |              |
2.start call        |    |Return address|  %rbp + 8
Stack grows down    |    |              |

===================================================================================

3.new function      |    +--------------+  新的函数栈起始地址
                    |    |              |
                    |    |     %rbp     |  在刚初始化的时候 sp=bp
                    |    |              |
                    |    +--------------+
                    |    |              |
                    |    |  local var1  |  %rbp - 8
                    |    |              |
                    |    +--------------+
                    |    |              |
                    |    | local var 2  | <-- %rsp
                    |    |              |
                    v    +--------------+
                         |              |
                         |              |
                         +--------------+
SS 栈段边界         0x00000
```
1. main函数开始执行时从sp栈开始初开始存储，这个sp当前是栈内存区域的最大边界，没新增一个变量或者一些存储操作则进行 压栈操作，如:
```c
#include <stdio.h>
int main()
{
    int a = 2;
    return 0;
}
会被翻译为如下汇编指令
main:
        push    rbp
        mov     rbp, rsp
        // 通过rbp -4 也就是用了栈的下面4字节来存储 int 2
        // 也就是压栈操作，其实这是一种直接操作栈的方式，这是编译器优化的结果
        // 正常情况下 应该使用  push 2;这种方式来操作栈，这样的话 sp始终会指向栈顶
        // 而通过偏移量来操作栈则不会引起 sp栈顶的变化
        mov     DWORD PTR [rbp-4], 2
        mov     eax, 0
        pop     rbp
        ret
```
2. 函数返回时的执行流程:
```c
int test(){
    return 2;
}
//汇编指令
test:
        //这里是压栈，当前的rbp其实是 调用放函数的rsp地址
        push    rbp
        //将当前栈顶 复制给rbp寄存器，从此开辟了一个新的函数栈区
        mov     rbp, rsp
        //这里就是我们程序实际逻辑开始的地方
        mov     eax, 2
        //程序结束，恢复调用方函数的栈底
        pop     rbp
        //这里就是返回调用方调用函数的地方，恢复函数继续运行
        ret
        //所谓函数返回，其实只是修改cpu的ip cs寄存器，修改cpu下一条需要执行的指令
        //那么下一条需要执行的指令其实就是 上面的Return address地址，我们也可以通过其他方法来实现ret
        // jmp %rbp+8;(%rbp+8  就是调用方函数的下一个cpu指令地址)从而实现了返回函数的功能
```
### 函数调用流程整理
来自：https://juejin.im/post/6844903930497859591  go plan9 汇编的函数调用图

因为总体流程大致相似
![image](/images/blog/goos/stack.png)
1. 每个函数执行期间 通过 `bp,sp`寄存器来表示内存区域
2. `bp`寄存器一般不会发生改变，一般通过bp+偏移量来获取相关栈上的变量
3. `sp`表示的是栈顶，调用`push`指令会自动修改sp指向的值
4. 通过整体流程的熟悉后，就能明白为什么栈数据是局部变量，会被回收（其实不是立即回收）
```
我们的栈是一个整段内存 0x00000 - 010000,整个栈内存都会不断复用，如上所示，当函数返回时，当前bp就会被恢复为之前调用方函数的栈，那么当前函数的区域就保持不变。
如果发生其他调用，则会复用当前函数的区域，则会覆盖当前变量

所以在c语言中返回一个局部变量地址，在其他地方依然能够访问的前提是因为没有新函数的栈内存将当前栈覆盖
```
### 栈&堆的区别
1. 栈是一块连续内存，由操作系统在程序执行期间为整个进程分配的生命周期
2. 堆内存是独立于当前栈的另外一个快内存，自然该内存不会受到像栈那样覆盖的影响，所以需要开发者自己管理，所以在c等静态语言中存在一个非常恐怖的问题(内存泄露),堆内存如果申请次数！=释放次数，那么你的内存就会逐渐飙升，等待系统给你kill吧


其实对于计算机来说，所有的都是二进制数据，没有代码和数据的区别，那怎么区分代码和数据呢，在cpu中有一个寄存器叫`ip`寄存器，存储的是下一条指令的地址，如果不发生中断的情况下顺序读取ip寄存器的值来进行执行，所谓的数据段只是应用层面划分的一块区域，使ip寄存器不会去访问该区域而实现的一个数据块，堆和栈就是典型的数据块，栈数据块会被多次复用，而堆数据块是栈快之外的额外需要向操作系统申请的一块内存


## 函数翻译后的cpu指令
再来看看一个c语言函数被编译后的汇编指令，因为汇编语言已经是最底层的语法表达，基本就是二进制指令一一对应，所以可以用汇编来表示最底层的cpu指令

下面是一个函数调用`test`和定义一个全局变量的例子
```c
#include <stdio.h>

char *str = "string data";

int test(){
    return 2;
}
int main()
{
    int a = 2;
    test();
    return 0;
}
```
编译后的汇编
```assembly
.LC0:
        .string "string data"
str:
        .quad   .LC0
test:
        push    rbp
        mov     rbp, rsp
        mov     eax, 2
        pop     rbp
        ret
main:
        push    rbp
        mov     rbp, rsp
        sub     rsp, 16
        mov     DWORD PTR [rbp-4], 2
        mov     eax, 0
        call    test
        mov     eax, 0
        leave
        ret
```
汇编最左边是一个标号，也可以当做地址，在其他地方直接通过该标号就可以引用到该地址

1. main: 这个标号是在程序启动时由外部来进行调用跳转的，所以程序开始的地方就是 `main`标号，也就是 `$rip = main:`设置 ip寄存器为main，开始执行main函数
2. `push rbp`: 基本所有的函数在执行前都要执行这行指令，表示将之前的栈底`rbp`保存起来，我们知道函数调用返回后需要恢复当前的栈环境，那么在调用函数之前，要保存当前的栈信息，所以需要`push rbp`
3. `mov rbp,rsp`: 这个就比较清楚了，表示开辟一个新栈，把当前的`栈顶`设置为新函数的`栈底`,那么新函数的执行环境就在新的栈空间使用
4. `sub rsp,16` : 这个模拟压栈，我们知道`rsp`代表的是栈顶，那么我们也可以手动将栈顶下移一定的空间，而申请的空间我们可以存储变量等信息,这行和手动执行2次`push ***`是相同的，因为push首先`rsp -= 8`然后在将数据写入栈区
5. `mov dword ptr [rbp-4],2`: 步骤4的时候新开辟了16字节的空间，这里就是通过对rbp进行偏移量来获取第一个4字节空间，然后将2存储进去，实现的一种手动压栈
6. `mov eax,0`: 这个没什么特别的，ax寄存器一般用作计算、传参等作用的寄存器，这里先初始化恢复为0
7. `call test`: 如1所说的，test是一个标号，也是一个地址，所以这里实际的执行可以分为如下两个步骤:
```
push cs //将代码段基段 cs保存起来
push ip //将ip段保存起来，这里相当于这个ip就是返回地址，当被调用函数返回的时，会获取当前换个ip在jmp %rip
jmp  test // 跳转到test标号的地址，实现函数调用
```
8. test: 进入test函数内部，首先执行`push rbp` 保存上一个函数的栈底指针
9. `mov rbp,rsp`: 和main函数一样开辟新栈
10. `mov eax,2`: 这里就是我们的c代码`return 2`的实际汇编指令，因为返回一般用ax寄存器存储，所以这里现将2存入eax寄存器
11. `pop rbp`: 恢复main函数的栈底指针，准备返回到main函数的下一行代码继续执行
12. `ret`: 可以表示为如下汇编`pop ip`实际就是获取main函数的之前保存的ip值，然后恢复到ip寄存器中，实现函数返回
13. 最后讲讲全局变量:
```
//char *str = "string data";
c代码会被翻译为如下的汇编指令，可以看到全局变量也是放到整个代码段上面的，如何区分该代码是数据还是代码呢，区别就在我们的程序如何去对待他
比如我们不管在何时引用.LCO时都是把他当做一个数据来处理，而不是加载到ip当做指令来执行
.LC0:
        .string "string data"
str:
        .quad   .LC0
```

# 协程的创建
这里来讲讲我们php扩展怎么创建一个协程，php代码和扩展的c代码怎么交互的问题

## php创建协程
php执行一个协程函数
```php
function task()
{
    echo "go task";
}
go(task);
```
c层面获取该函数`wrapper/coroutine.cpp`：
```c
PHP_FUNCTION(go_create)
{
    zend_fcall_info fci = empty_fcall_info;
    zend_fcall_info_cache fcc = empty_fcall_info_cache;
    //1 -1 可变参数
    ZEND_PARSE_PARAMETERS_START(1,-1)
    Z_PARAM_FUNC(fci,fcc)
    Z_PARAM_VARIADIC("*",fci.params,fci.param_count)
    ZEND_PARSE_PARAMETERS_END_EX(RETURN_FALSE);
    long cid = PHPCoroutine::go(fcc.function_handler,fci.params,fci.param_count);
    RETURN_LONG(cid);
}
```
通过`PHP_FUNCTION`申明一个提供给php调用的api，`go`实际执行的是c的`go_create`。`fci,fcc`可以表示一个php传过来的函数参数.
通过`PHPCoroutine::go`来初始化一个协程，并投递到调度器去执行

```c
/coroutine/PHPCoroutine.cpp
/**
 * 创建一个协程G运行
 * @param call
 * @return
 */
long PHPCoroutine::go(zend_function *func,zval *argv,uint32_t argc)
{
    ZendFunction *call = new ZendFunction(func,argv,argc);
    Coroutine *ctx = new Coroutine(run, call);
    return ctx->run();
}
```
1. 拷贝当前用户函数，因为多线程协程情况下已经采取了线程隔离`TSRM`,所以该闭包任务呗调度到其他线程执行时环境不同，且当前函数返回后可能被回收等因素，需要对用户的函数进行硬拷贝，拷贝会专门在线程隔离中说明。
2. 创建一个`G`绑定当前php用户函数，等待投递调度

```c
//coroutine/Coroutine.cpp
 * 投递到调度到其他线程CPU中去执行
 * @return
 */
long Coroutine::run()
{
    //投递到 proc 线程去执行该协程
    if(proc == nullptr){
        cout << "未初始化线程" <<endl;
        throw "未初始化线程";
    }
    proc->gogo(ctx);
    return 1;
}
```
到这里基本就完成了一个php协程创建到执行的过程了，`proc->gogo`后面就是属于调度和任务投递的事情了，这个是多线程调度处理的，会有专门的章节讲解

# 全局队列与本地队列
目前实现的多线程协程基于两个队列来调度任务，一个是全局队列，所有线程获取时需要枷锁，另外一个是本地队列，目前只处理被调度过的协程，不接受新协程投递
```c
//runtime/proc.cpp
            unique_lock<mutex> lock(this->queue_mu);
            this->cond.wait(lock,[this,rq]{
                return this->stop || !this->tasks.empty() || !rq->q->isEmpty();
            });

            if(this->stop && this->tasks.empty() && rq->q->isEmpty())
                break;

            if(!this->tasks.empty()){
                ctx = move(this->tasks.front());
                this->tasks.pop();
                co = static_cast<Coroutine *>(ctx->func_data);
            }else{
                co = rq->q->pop();
            }
        }
        if(co == nullptr){
            cout << "co exception:"<<co<<endl;
            continue;
        }
```
1. `tasks` 是一个全局队列，新创建的协程优先投递到tasks等待所以线程获取，这里的问题就是会导致竞争严重，多线程会同时获取锁来争抢该协程
2. `rq->q` 是一个本地队列，通过`GO_ZG(rq)`来获取该队列，所以调度的前提就是本地队列和全局队列都有数据则触发调度循环，获取待处理的协程进行切入

# 协程的释放
协程的释放，目前协程的释放会回收c栈和php栈，会极大的影响性能，后面会实现c和php栈复用，更好的提高性能
```c
//coroutine/coroutine.cpp
void Coroutine::close()
{
    zend_vm_stack stack = EG(vm_stack);
    free(stack);
    restore_stack(&main_stack);
    delete ctx;
    delete this;
}
```
1. 将当前的通过堆申请的栈销毁，也就是销毁php栈
2. 恢复在协程切入前的主php栈，模拟函数返回
3. 删除`ctx`也就是c栈，回收c栈
4. `delete this`删除G相关内存，回收内存
