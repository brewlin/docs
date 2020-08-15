---
title: Goos-底层协程实现(三)
toc: true
date: 2020-08-13 21:28:59
tags: [linux,c,php,ext,thread,coroutine,scheduler]
---

本节主要讲解什么是多线程协程的投递、调度、以及切换的汇编实现

- [Goos-多线程协程实现简要]()
- [Goos-协程底层实现(一)]()
- [Goos-线程协程隔离(二)]()
- Goos-线程切换实现(三)
- [Goos-抢占调度实现(四)]()
- [Goos-监控线程实现(五)]()

# 协程调度流程

## 全局队列队列投递
我们在php层创建一个协程的方法如下:
```
go(function()){
   //do something...;
});
```
调用`go`函数，将一个php函数参数传入协程调度器执行，这里go函数执行完毕之前是异步的，不会立即就执行该任务，而是将该任务投递到一个全局队列里，等待线程接收后处理

投递任务:
```cpp
//coroutine/Coroutine.cpp

long Coroutine::run()
{
    //投递到 proc 线程去执行该协程
    if(proc == nullptr){
        cout << "未初始化线程" <<endl;
        throw "未初始化线程";
    }
    proc->gogo(ctx);
    
    //本来是会新生成一个协程id返回的，但是目前没什么用
    return 1;
}
```
1. 主要是判断 线程调度器有没有初始化，`proc == nullptr`
2. 调用全局`proc->gogo(ctx)` 将封装好的一个ctx协程投递到线程中去

## 全局队列结构
```cpp
//runtime/Proc.h
class Proc
{
public:
    //method ....
public:
    //...省略其他字段
    condition_variable  cond;
    queue<Context *>    tasks;
private:
    vector<thread>      workers;
    mutex queue_mu;
    bool stop;

};
```
1. `cond` 条件变量，用户获取锁的时候使cpu睡眠时等待唤醒的条件
2. `tasks` 为一个全局队列，用于接收投递的协程G
3. `workers` 默认启动的线程M
4. `queue_mu` 线程锁

新的协程创建后会投递到`tasks`队列，然后随机唤醒一个线程M来处理该协程:
```cpp
//runtime/proc.cpp

void Proc::gogo(Context* ctx)
{
    unique_lock<mutex> lock(queue_mu);
    now = chrono::steady_clock::now();
    this->tasks.emplace(ctx);
    cond.notify_one();
}

```

## 协程调度执行
每个线程M的初始化执行后，会进入schedule事件循环，如果没有信号过来则默认会进入睡眠状态，等待唤醒后处理投递进来的协程G，并初始化协程环境后绑定当前M-G的关系后执行该php用户态函数
```cpp
//runtime/pro.cpp

void Proc::schedule()
{
    for(;;){
        Context*   ctx;
        Coroutine* co;
        //省略本地队列 。。相关逻辑
        {
            unique_lock<mutex> lock(this->queue_mu);
            this->cond.wait(lock,[this,rq]{
                return this->stop || !this->tasks.empty();
            });

            if(this->stop && this->tasks.empty())
                break;

            if(!this->tasks.empty()){
                ctx = move(this->tasks.front());
                this->tasks.pop();
                co = static_cast<Coroutine *>(ctx->func_data);
            }
        }
        if(co == nullptr){
            cout << "co exception:"<<co<<endl;
            continue;
        }
        //当前线程分配到一个未初始化的G
        if(co->gstatus == Gidle) co->newproc();
        //恢复被暂停的G
        else co->resume();
        //G运行结束 销毁栈
        if(ctx->is_end) co->close();
        //省略一些其他的。。。。
    }
}
```
1. 当前线程处于cpu睡眠态，等待唤醒的方式目前有两个情景
    - `gogo()` 协程投递的时候会触发唤醒随机线程 `cond->notify_one()`
    - `sysmon` 监控线程有一些管理任务会涉及当前线程去处理任务
2. `pop tasks` 出队列拿到一个协程G任务
3. 判断该G是`新协程`还是需要再次`恢复`的调度协程
4. 新协程为`co->newproc()` 开始走新协程的调用
5. 中断协程的恢复走`co->resume()`恢复协程的继续运行
6. 如果协程状态为`close`则回收该协程资源

# 协程的c栈内存模型

## 协程的创建执行
针对新协程的执行:
```
//coroutine/Coroutine.cpp

void Coroutine::newproc()
{
    callback->is_new = 0;
    callback->prepare_functions(this);
    PHPCoroutine::save_stack(&main_stack);
    GO_ZG(_g) =  this;
    //每次切入时出去时需要更新tick 和时间
    GO_ZG(schedwhen) = chrono::steady_clock::now();
    GO_ZG(schedtick) += 1;
    gstatus = Grunnable;
    ctx->swap_in();
}
```
1. 准备当前G的php环境，比如拷贝当前G对应引用的php`全局变量`，`类对象`，`外部引用`等等,这个会在`线程协程隔离中说明`
2. 保存当前`php栈`信息，如第一章中函数本质部分说的，在调用一个函数前，会将当前`sp,bp,ss:ip`等必要寄存器压栈保存，在函数放回的时候会找到该地址，然后进行跳转实现返回。不过这个是`php栈`
3. `GO_ZG(_g) = this` 将当前G绑定到M上 
4. `gstatus = Grunable` 标记当前G运行状态
5. `ctx->swap_in()` 正式执行该协程

## C栈内存结构
首先是c栈的内存申请过程
```cpp
//runtime/Context.cpp

Context::Context(run_func func,void *data):_fn(func),func_data(data)
{
    bp =  new char[DEFAULT_STACK];
    make_context(&cur_ctx,&context_run, static_cast<void *>(this),bp,DEFAULT_STACK);
}
```
1. 创建一个`8k`的堆内存，用于实现c的函数栈
2. 调用`make_context` 初始化该栈帧内存结构

在实际执行用户传递的php函数前还有一个包装流程:
```
//runtime/Context.cpp

/**
 * 主要运行的函数
 * @param arg
 */
void Context::context_run(void *arg)
{
    Context *_this = static_cast<Context *>(arg);
    _this->_fn(_this->func_data);
    _this->is_end = true;
    GO_ZG(_g) = nullptr;
    _this->swap_out();
}
```
1. 这里的`_this->func_data`就是协程G对象，实际执行单元
2. `_this->_fn` 是一个函数指针，指向`PHPCoroutine.cpp::run()`,该函数初始化php栈帧信息，准备执行实际的php函数
3. `_this->_fn` 执行完毕则代表该G生命周期完毕，否则说明当前G已经被暂停，切换出去了，等待恢复继续执行
4. `_this->is_end =true` 标志当前G 已结束,
5. `GO_ZG(_g) = nullptr` 解绑当前`G - M`的绑定关系
6. `_this->swap_out()`  这里很重要，当前函数依然是在协程范围内，所以必须显式通过`swap_out()`模拟函数`return`返回到之前的函数调用，否则没有任何意义，因为cpu不知道下一条待执行的指令是什么，无法回到正常的执行流程

## c栈的内存模型
这里比较重要，需要将堆内存转换为函数栈，且将一些必要配置初始化，例如将协程G的函数地址压栈`(压堆)`，以及增加安全机制

通过调用`make_context` 将堆内存转换为普通c函数栈模型，为实现函数调用做准备
```cpp
//runtime/asm/make_context.cpp

#define NUM_SAVED 6
void make_context (asm_context *ctx, run_func fn, void *arg, void *sptr, size_t ssize)
{
  if (!fn)
    return;
  ctx->sp = (void **)(ssize + (char *)sptr);
  *--ctx->sp = (void *)abort; 
  *--ctx->sp = (void *)arg;
  *--ctx->sp = (void *)fn;
  ctx->sp -= NUM_SAVED;
  memset (ctx->sp, 0, sizeof (*ctx->sp) * NUM_SAVED);
}
```
1. 检查`fn`是否存在
2. 因为函数栈内存是从高地址往低地址增长，所以 `sp寄存器指向的栈顶`要指向堆的结束地址位置`((void **)(ssize + (char *)sptr))`
3. `*--ctx->sp = (void *)abort;` 其实就是压栈，将一个`abort`函数地址压栈，且sp地址自动下移，abort函数是一个保障机制，如果某个协程G没有实现跳转回主流程，则调用`abort`报异常
4. `*--ctx->sp = (void *)arg;` 将函数参数压栈，在跳转的时候可能需要传递参数，到时候通过`popq %rdi`将arg送入`rdi`寄存器实现函数传参
5. `*--ctx->sp = (void *)fn`   将函数地址压栈，cpu在执行时通过获取该地址后跳转，实现函数调用
6. `ctx->sp - = 6`; 腾出6个变量的位置，用于存储上下文信息，比如在函数切换前要保存之前的寄存器变量信息
![image](/images/blog/goos/3-stack.png)

# 协程切换的汇编解析
协程切换的汇编实现为`/runtime/asm/jump_context.s`:
```assembly
.text
.globl jump_context
jump_context:
    pushq %rbp
    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15
    movq %rsp, (%rdi)
    movq (%rsi), %rsp
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx
    popq %rbp
    popq %rcx
    popq %rdi
    jmpq *%rcx
    popq %rcx
    jmpq *%rcx

```
## 汇编指令解析
- `.text` 标明下面是一块代码段，在cpu指令执行过程中能够确认他们是指令段而非数据段
- `.globl jump_context` 这里相当于c语言声明一个函数名的作用，对于cpu来说函数其实就是一个指令地址，这里也是用于在连接过程中将当前函数的地址进行标记

## 保存上下文
- 保存当前函数的上下文，对于程序上下文来说，其实细分到cpu，就是保存该函数时刻的寄存器对应的值和函数栈bp,sp的位置，基本靠这些就可以标明当前某个函数的执行状态了
```
下面的6个寄存器应该符合调用者规约，也就是在调用其他函数前应该由调用者保存起来，防止在子函数中被篡改
%rbx,%rbp,%r12,%r13,%r14,%r15

    pushq %rbp 将当前函数栈底rbp寄存器保存起来 
    pushq %rbx 将基地址寄存器保存起来，bx操作评率较高，bx默认指 ds数据段，一般用作内存访问
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15
```
## 函数参数

- `接受参数`: jump_context() 接收两个参数，`prev指针，其实就是当前c栈`,`next* 目标c栈`，因为函传参的底层汇编实现是通过寄存器来实现的，所以`prev,next`参数默认是按照保存到`rdi,rdx`寄存器中

![image](/images/blog/goos/3-stack2.png)

顺便提一下:通常如果参数比较少的话（一般6个作为界限），则通过寄存器进行传参数。顺序为:
```
%rdi,%rsi,%rdx,%rcx,%r8,%r9 
依次对应
func(arg1,arg2,arg3,arg4,arg5,arg6);
```


如果超过了6个，就需要栈来辅助接受函数参数了，如上图所示。在调用者函数栈顶前一个，则是存储的函数参数
## 函数栈祯切换
这个就是核心功能了，我们知道在正常的函数调用执行流中，我们都是使用了程序装载前分配的那个系统栈，不出意外从程序开始到结束都是不断的复用该程序栈。
但是由于协程的出现，基于堆内存模拟的函数栈。那么在调用函数的时候就必须`切换栈`
![image](/images/blog/goos/3-stack3.png)
如上图，只要是默认的c函数或者业务函数都是基于系统栈祯的，例如调用`A`函数的时候默认在系统栈祯下面使用新的空间来存储A函数的栈祯，都是使用的系统栈

而如果我们此时要进行协程调用，则需要将cpu的`sp`等寄存器切换到我们的协程B函数的栈祯首地址，那么cpu的执行流就会切换到协程B栈上执行，所有的变量内存都会依赖心的协程B栈，`注意`:毕竟协程B的栈是堆模拟出来的，所以是预分配有限制大小的内存，在使用的时候不要越栈，并且协程B栈执行完后一定要恢复到兄台你栈祯的继续执行

```
movq %rsp, (%rdi)
movq (%rsi), %rsp
```
这两条汇编指令实现了系统栈 - 协程栈的切换，`rsp`当前系统栈栈顶，`rdi`第一个函数参数,保存调用者的栈信息(可能是系统栈，也可能是协程切换了多次，也可能是协程栈本身)。`rsi`第二个函数参数,保存的被调用者函数的地址信息(可能是协程栈祯，也可能是协程结束后，准备切换为系统栈的栈祯)


## 恢复环境上下文
到这里已经切换到了协程栈，远离的系统栈，下面的汇编指令是实现恢复上下文寄存器，`在第一次协程创建的时候是空的`，但是当切换多次后就会发现，者6个寄存器永远保持上一个协程状态的环境

```
popq %r15
popq %r14
popq %r13
popq %r12
popq %rbx
popq %rbp
```
pop的过程如下，总的来说就是将栈上的变量恢复到寄存器中，实现函数状态的恢复，`第一次协程是没有意义的`因为默认是6个寄存器的占位符
![image](/images/blog/goos/3-stack4.png)

## 函数的调用
目前的函数栈祯如下
```
------------
|   abort  |
------------
|   arg    |
------------
|   func   |
-----------
```
可以看到我们的函数栈只剩下三个值了，接下来的汇编指令将pop栈，实现函数调用
```
popq %rcx
popq %rdi
jmpq *%rcx

```
这里有两次`pop`说明出栈了两个数据`func,arg`刚好对应我们的函数地址和函数参数地址，

`popq %rcx`： 这里将函数地址保存到`rcx寄存器`,为什么选rcx寄存器呢，没啥区别，选啥都可以，反正就是为了拿到函数地址而已

`pop %rdi` :  这里就是将arg指针保存到`rdi`寄存器，前面说过函数传参按照顺序来说第一个参数的寄存器就是`rdi`所以讲arg指针保存到`rdi`寄存器实现函数传参

`jmpq *%rcx`:  这里就是真正执行的函数调用，`rcx`保存的是我们的函数地址，jumq 就是让cpu跳转到该函数指令地址执行，实现函数调用


## 协程栈收尾
到这里的时候，说明程序已经崩溃了，目前协程栈内存模型:
```
------------
|   abort  |
------------
```
如果执行到这里，说明我们的协程==没有切回主系统栈==,那么这里直接调用`abort`给一个通知

```
popq %rcx
jmpq *%rcx
```
执行`abort()`函数


# 总结
我们通过自己申请一块堆内存来模拟函数栈实现函数调用是为了更好的控制该函数的生命周期，以此实现函数的暂停、恢复等操作，有点类似于线程，但是性能更好、代价更小，甚至和普通函数调用无差别，这就是协程、一种用户态线程