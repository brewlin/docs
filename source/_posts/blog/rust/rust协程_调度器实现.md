---
title: 无栈协程_调度器实现
toc: true
date: 2022-04-07 21:28:59
tags: [linux,c,golang,rust,coroutine,scheduler]
---
本次文章专注于分析rust的协程实现，并提前介绍它与其他语言的协程有什么区别

接下来我们就从有栈和无栈切入了解协程在内存方面的布局，典型的有栈协程就是golang，这里的`有栈`其实是一个潜台词，代表的是需要单独申请堆内存作为栈的意思

那么有栈协程对应的就是=>需要独立申请一份内存作为代码指令运行的栈

# 关于栈的解析
典型的就是c语言的栈，来看看c语言的运行栈的情况吧
```
void test1(long  var3,long var4){
//函数参数也是变量，看看各个编译器的实现
//一般也是存在栈上，所以从下面图片可以看到即时我没有显示定义两个变量，但实际还是占用了栈
}
void test(){
    long var1 = 1; //用long 是因为64位下占8字节方便我下面画图。。。
    long var2 = 2;
    test1();
    var2 = 3;
}
```
当前没有讲解从main函数开始，是因为main函数有点特殊，因为main的参数只有`argc,argv`。且直接通过`(%rsp),8(%rsp),分别就能拿到 argc,argv。不是直接像其他那样通过`%rdi,%rsi`拿，为了保持简洁明了，直接用了非main函数其他的函数作为例子讲解

## 栈的内存结构
上文的c语言在编译运行后的系统栈内存使用如下：
```
|    +--------------+   %rbp <---------------------                
|    |   var1       |                             |
|    +--------------+                             |
|    |   var2       |                             |
|    +--------------+                             |
|    |   arg1       |                             |
|    +--------------+                             |
|    |   arg2       |                             |
|    +--------------+                             |
|    |Return address|    test                     |
|    +--------------------------- %rsp            |
|    |   %rbp       |    --------------------------
|    +--------------+    
|    |   var3       |    
|    +--------------+
|    |   var4       |    
|    +--------------+
|    |Return address|    test1
V    +--------------------------- %rsp
```
1. 首先栈是从高地址到低地址发展
2. 所有的变量会提前在编译阶段就计算好栈中的地址
```
test函数内:
//var1的变量
var1 = -8(%rbp);
var2 = -16(%rbp)

test1函数内:
var3 = -8(%rbp);
var4 = -16(%rbp);
```

## 栈的上下文恢复
从上文我们可以看到，整个程序运行就是各种函数调用，都会不断的追加到栈中，不停的阔栈，整个进程都是复用的同一个栈，所有变量和地址都是依赖于`%rbp`进行地址定位(依赖于编译器实现，一般都是根据%rbp进行定位)

那么当`test1`函数返回后，如何恢复`test`函数的环境呢？
```
1. 变量的定位需要依赖`%rbp`,所以将(test1.%rbp)恢复到(test.%rbp)即可
2. %rsp指向的栈顶，当test1函数返回后，其实 %rsp = %rbp即可，因为test1函数的基站其实就是test函数的栈顶
3. 当然还要恢复%rip，下一条指令执行的地址，也就是test函数var2=3的地址，也是在test1函数返回时需要恢复的
```
`rsp,rip,rbp`基本上靠这个就组成了整个函数的上下文，这也就理解了c类语言的内存栈布局

那么就可以非常明了的解释了为什么不能轻松操作栈引用的问题了
```c
int* test1(int var3) {
    return &var3;
}
void test(){
    int* var1 = test1(10);
    int* var2 = test1(20);
}

|    +--------------+   %rbp <---------------------                
|    |   var1       |                             |
|    +--------------+                             |
|    |   var2       |                             |
|    +--------------+                             |
|    |   arg1       |                             |
|    +--------------+                             |
|    |   arg2       |                             |
|    +--------------+                             |
|    |Return address|    test                     |
|    +--------------------------- %rsp            |
|    |   %rbp       |    --------------------------
|    +--------------+    
|    |   var3       |    
|    +--------------+
|    |Return address|    test1
V    +--------------------------- %rsp
```
1. 每次调用进入到`test1`函数时，都返回了`lea -8(%rbp)`,因为%rbp的值是一样的所以他们引用的地址其实都是一样的
2. 所以即使`*var1`第一次调用时的值是 10;
3. 但是第二次调用时`*var1`所指向的地址内的值被20替换，导致最后`*var1,*var2`都是20

# 有栈协程 -> golang
上面的栈例子中典型的用了c语言的例子，程序总体同步顺序运行，全部公用一个系统栈，并且随着不断的函数调用、函数返回，这个栈会不断复用

而有栈协程中的`有栈`到底是什么意思，这里其实是约定成俗想表达对于每个新创建的协程来说:他们都独立运行与一块新的栈，`这块栈是从堆(基于mmap维护了整个内存管理)上面申请的`，没用共用系统栈，那么这个协程的生命周期和上下文都能够被完整保存，可以被任意时间和任意线程独立执行


## go 创建协程
在golang语言中，直接通过`go`关键字可以轻松创建一个协程，并传递一个待执行的`func函数`,在此之后整个func和当前主线程再无瓜葛，它会被任意调度到任意线程去执行或多次执行

```golang
go func(){
    fmt.Println("this is a goroutine print!")
}
```
实际在编译后被换成了调用`runtime.newproc`方法
```
// 创建一个协程，用来运行传入的带有siz字节参数的函数
// 将协程push到队列里 等待调度运行
// 不能进行栈切分，因为函数参数需要拷贝，如果栈分裂的话可能fn后面的参数不完整了
//go:nosplit
func newproc(siz int32, fn *funcval) {
	//总的来说这里是编译器来调用的newproc方法，第一个参数siz 指明了调用函数fn的参数大小
	//NOTICE: 参数全部是存放在栈上的，所以通过fn后面的偏移量+参数大小就可以完整的拷贝函数参数了
	argp := add(unsafe.Pointer(&fn), sys.PtrSize)
	//stack: [size,fn,arg1,arg2,arg3....] size 指明了arg1..argn的栈范围大小
	gp := getg()
	//获取调用方的下个指令地址，一般用于设置ip寄存器用于表示下一行代码该执行哪
	pc := getcallerpc()
	systemstack(func() {
		newproc1(fn, (*uint8)(argp), siz, gp, pc)
	})
}
```
- `fn`就是上文我们传递的闭包函数，待异步执行的函数方法
- 如果在创建协程的时候，带上了参数，也能通过栈偏移量获取到函数参数`add(unsafe.Pointer(&fn),sys.PtrSize)`.需要马上将参数拷贝到协程空间内，因为这些参数仍然是存放在主线程栈上的(go嵌套则不一样)

## 分配栈
这里任然是处于创建协程的收尾部分，主要处理两件事:
1. 给协程分配2k的内存作为函数运行的占空间(会复用其他已经释放的协程栈)
2. 将协程丢给全局队列等待释放。在rust.tokio中相当于丢给`worker.shared.inject.push(task)`全局队列等待调度
3. 设定`exit`函数，实现永不返回的循环调度（不同的协程栈之间切换已经没有return的概念了，直接永不停歇的往前走）
```golang
// 在系统栈、g0栈上创建                                                                                                                                      uu一个协程
// 1. 拷贝参数到协程里
// 2. 初始化基本信息如，调用方的下一行代码地址，ip寄存器
// 3. 将协程推入全局列表等待调度
func newproc1(fn *funcval, argp *uint8, narg int32, callergp *g, callerpc uintptr) {
	//从tls中获取线程对应的协程
	_g_ := getg()

	//复用已经被释放了的之前的协程栈
	newg := gfget(_p_)
	if newg == nil {
		//立即创建一个协程+ 2k协程栈
		newg = malg(_StackMin)
		//将g转换为dead状态
		casgstatus(newg, _Gidle, _Gdead)
		//添加到allg全局队列管理
		allgadd(newg) // publishes with a g->status of Gdead so GC scanner doesn't look at uninitialized stack.
	}
    //....
	sp := newg.stack.hi - totalSize
    //...
	memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
	//初始化时 记录了协程内当前栈顶 和 基栈
	newg.sched.sp = sp
	newg.stktopsp = sp
    //..
	//协程内需要执行的代码指令地址，初始化时指向了函数的首地址,而在后面的生命周期中 会不断调度切换后会变化
	newg.startpc = fn.fn
    //协程id
	newg.goid = int64(_p_.goidcache)
	_p_.goidcache++
    //..
	//将协程投递到本地队列或者全局队列等待调度器调度
	runqput(_p_, newg, true)
    //..
	//顺便检查下，如果当前需要抢占则处理抢占
	if _g_.m.locks == 0 && _g_.preempt { // restore the preemption request in case we've cleared it in newstack
		//编译器在函数调用的时候会检查是否栈溢出，这里巧妙的利用栈溢出来实现抢占
		_g_.stackguard0 = stackPreempt
	}
}

```

## 栈的上下文
理想情况下面的函数在单线程中会顺序调用和执行，那么根据这种情况可以绘制一个理想的栈使用情况
```golang
package main
func go1_1(){var var11 uint64}
func go1(){
	var var1 uint64
	go1_1()
	//go1_e 函数结束的指令地址
}
func main(){
	runtime.GOMAXPROCS();
	go go1()
	go go2()
	select{}
}
func go2(){go2_1()
}
func go2_1(){}
```
![image](/images/blog/rust/golang-stack.png)
golang的协程永不return，不停的在协程间切换
```golang
// One round of scheduler: find a runnable goroutine and execute it.
// Never returns.
func schedule() {
}
```
在研究调度器的时候会有一个疑问，注释明明写的`never returns`，但是却没有看到死循环的操作，那么是如何实现永不return的呢

总的来说可以总结以下几个函数的调用顺序来概览到这种循环机制
```golang
schedule()      g0栈上:开始执行一轮调度,找到需要唤醒的G
execute()       g0栈上:开始唤醒协程G，切换到协程栈
mcall(goexit1)  g栈   :开始切换到g0栈上回收以及结束的G
schedule()      g0栈上:开始执行一轮调度，找到需要唤醒的G
```
主要是三个函数就能描述了整个调度的生命周期，但其实还有一个问题在，上面的循环很像一个递归调用，那可不可能发生爆栈呢

核心就在于g0栈是复用的,也就是在每次从g0栈切换到g栈的时候是不保存g0栈的，这么就会导致g0栈始终会从默认的地方在下次继续执行
```golang
TEXT runtime·gogo(SB), NOSPLIT, $16-8
	MOVQ	buf+0(FP), BX		// bx = gobuf
	MOVQ	gobuf_g(BX), DX     // dx = g
	MOVQ	0(DX), CX		// make sure g != nil
	get_tls(CX)
	MOVQ	DX, g(CX)       // 将目标g 设置为当前线程 tls->g
	MOVQ	gobuf_sp(BX), SP	// restore SP  恢复sp栈顶指针 $rsp = gobuf.sp 实现栈切换
	MOVQ	gobuf_ret(BX), AX  // ax = gobuf.ret
	MOVQ	gobuf_ctxt(BX), DX // dx = gobuf.ctxt 上下文信息
	MOVQ	gobuf_bp(BX), BP    // 恢复bp寄存器  $rbp = gobuf->bp 栈基指针 执行当前函数开始位置
```
可以看到从g0栈切换到g栈的核心方法`gogo`中并没有保存当前g0的上下文，也就是说`g0->sched`上下文信息始终没有发生变化，在下次通过`mcall`等切换回g0时不会导致g0栈空间的开辟


# 无栈协程->rust
上面描述了两种栈，一种是c的全局系统栈，另外一个就是基于堆的golang协程栈

可以明显感受到golang的栈会复杂很多，而且开销非常大
1. 所有的协程都会默认分配2k内存
2. 随着协程内函数调用的嵌套层级增大，2k栈明显不够用，那么会触发栈的扩容
3. 栈扩容又会引发一系列引用问题

但协程的实现又要保存上下文，不依赖单独的栈如何做到呢？。带着这个疑问来分析rust的黑魔法吧

还是先来感受下golang和rust关于协程的语法区别吧
```golang
//golang
func main(){
    go fun(){
        fmt.Println("ready to sleep!")
        time.Sleep(8 * time.Second) 
        //会暂停当前函数执行,给其他协程继续执行
        //等待睡眠时间到后重新调度后继续从当前位置向下执行
        fmt.Println("hello world 1!")
    }
}

```
```rust
//rust
tokio::spawn(async {
    println!("ready to sleep");
     tokio::time::sleep(time::Duration::from_secs(2)).await;
     //注意： 一定要加await！
     //当前函数会在这里暂停，等待睡眠时间到后继续恢复执行
    println!("hello world!");
});
```
两种语言的协程实现都能表现同样的功能，但rust已经能够感受到需要注意规范是比较多的
1. rust没有自带的运行时，所有协程的调度、执行、切换都需要依赖三方实现，比较好的就是`tokio`
2. 编译器只干了一件事情: `在有await语句的地方检测是否ready，否则挂起函数，等待下次运行`


既然rust不像golang那样有单独的栈，那他怎么实现上下文保存和栈的重入呢？，比较都是依赖主线程栈，怎么切换呢，不得甚解

在了解rust通过await实现协程(特别强调await)前来一起看看什么叫做状态机吧

## rust 状态机
理解协程的核心就是暂停和恢复，rust的协程通过状态机做到这一点，golang通过独立的栈做到这一点。理解这一点很重要

看个例子:(为了模拟暂停状态，需要自己实现一个future(async语句块))
```rust
pub struct Task {
    ready: bool,
}
impl Future for Task {
    type Output = bool;
    fn poll(self: Pin<&mut Self>, cx: &mut Context<'_>) -> Poll<Self::Output> {
        let mut ms = unsafe { self.get_unchecked_mut() };
        if !ms.ready {
            ms.ready = true;//下一次再进来就是true
            println!("task not ready!");
            Poll::Pending    
        }else{
            println!("task is ready!");
            Poll::Ready(ms.ready)
        }

    }
}
async fn test(){
    let task1 = Task{ready:false};
    task1.await; // 发生第一次暂停，因为task1返回了Poll::Pending,当前test也是一个协程，会保存上下文，结束当前函数
    println!("test done!");
}
fn main(){
    tokio::spawn(test());
}

```
1. 上面的实例永远只会打印一次`task is ready!`就结束了test（）函数的执行，下面的`test done!`永远不会被执行
    - 因为对于调度器tokio来说，他永远不知道task会在何时转换为Poll::Ready状态
    - 所以一般真正的阻塞的需要切换的地方tokio都覆盖完了，比如(网络io，sleep等等api)，都会在发生阻塞的时候切换出去，在唤醒后主动在来调用一次
    - 那么自己实现的阻塞的方法，那就需要注册唤醒器让tokio有能力重新调度了
2. 假定我们实现了唤醒器,并且将`ms.ready=true`注释掉`
    - 那么依然test函数会被重复调度运行，但用于只会执行 `task not ready`那段逻辑
    - 因为：rust在编译阶段做了手脚，通过状态(Pending,Ready)来区分该执行哪段逻辑

现在还是有点晕，状态机到底是个什么东西，那么我们就从编译器的视角来看看上面的rust代码被编译器魔改后到底实际执行的是什么代码吧

## 编译器生成的状态机代码
还是上面的那个例子，让我们来看看编译器最终生成的是什么代码吧(如下伪码)
```rust
/*
//原始代码
async fn test(){
    let task = Task{ready:false};
    task.await;
    println!("test done!");
}
*/

enum test {
    Enter,
    Yield1 {
        task: Task,
    },
    Exit,
}

impl test {
    fn start() -> Self {
        test::Enter
    }
    fn execute(&mut self) {
        match self {
            test::Enter => {
                let task = Task{ready:false}; //源代码
                *self = test::Yield1{task:task} //保存上下文
            }
            test::Yield1 {task} => {
                if task.poll() == Poll::Ready {//task.await 伪码
                    println!("test done!");
                    *self = test::Exit; //await结束 
                    return;
                } else {
                    return;
                }
            }
            test::Exit => panic!("Can't do this again!"),
        }    
    }
}
pub fn main(){
    let t = test::start();
    //tokio::spawn(test())
    //背后其实就是多次t.execute()
    //第一次 t.execute() ready false 打印: not ready
    //第二次 t.execute() ready true  打印: test done
    // 如果task.ready状态一直为false，那么会一直执行test::Yield1这个分支
    //编译器将async fn test() 生成一个enum带有状态的状态机，从而实现了直接在系统栈上就能够实现协程的暂停与恢复
}
```
1. `重点`: rust编译器会将带有.await的代码快转换为一个enum 状态机，就像上文一下test函数被改成了`enum test`
2. 对于每个await的代码都实现为一个enum的分支
3. 每次协程的暂停和恢复只是进入不同的代码分支罢了

虽然没有了额外的占的开销，但实际上编译器会生成很多指令和分支来支持这个状态机

不过相比需要额外的栈内存来实现协程，这种方式已经非常棒了

## 协程调度器(tokio runtime)
到上面为止我们只分析到了函数的暂停与恢复（协程的基本要素）。但何时暂停何时恢复这个rust并没有实现

调度器这种运行时功能目前比较好的三方实现是tokio

### 1. 启动多线程
初始化一个多线程的tokio运行时
```rust
fn main(){
    let rt = tokio::runtime::Builder::new_multi_thread().enable_all().build().unwrap();

}
```
builder会区分是多线程版本还是单线程版本
```rust
//tokio/tokio/src/runtime/builder.rs:514
pub fn build(&mut self) -> io::Result<Runtime> {
    match &self.kind {
        Kind::CurrentThread => self.build_basic_runtime(),
        #[cfg(feature = "rt-multi-thread")]
        Kind::MultiThread => self.build_threaded_runtime(),
    }
}
```
创建系统多线程
```rust
//tokio/tokio/src/runtime/builder.rs:663
cfg_rt_multi_thread! {
    impl Builder {
        fn build_threaded_runtime(&mut self) -> io::Result<Runtime> {
        //省略参数初始化。。。
                    // Spawn the thread pool workers
        let _enter = crate::runtime::context::enter(handle.clone());
        //开始派生系统线程
        launch.launch();
        Ok(Runtime {
            kind: Kind::ThreadPool(scheduler),
            handle,
            blocking_pool,
        })
        
        }
    }
}
```
### 2. worker线程调度协程任务
每个worker线程进入轮训模式
```rust
//tokio/tokio/src/runtime/thread_pool/worker.rs:382
impl Context {
    fn run(&self, mut core: Box<Core>) -> RunResult {
        while !core.is_shutdown {
            // Increment the tick
            core.tick();

            // Run maintenance, if needed
            core = self.maintenance(core);

            // First, check work available to the current worker.
            if let Some(task) = core.next_task(&self.worker) {
                core = self.run_task(task, core)?;
                continue;
            }

            // There is no more **local** work to process, try to steal work
            // from other workers.
            if let Some(task) = core.steal_work(&self.worker) {
                core = self.run_task(task, core)?;
            } else {
                // Wait for work
                core = self.park(core);
            }
        }

        core.pre_shutdown(&self.worker);

        // Signal shutdown
        self.worker.shared.shutdown(core);
        Err(())
    }
}
```
短小精干，整个调度轮训代码就这么多，有点借鉴了golang的调度器，基本都是
1. LIFO slot: 从优先队列获取协程任务执行
2. local queue: 从本地队列获取协程任务执行
3. global queue: 从全局队列获取协程任务执行
4. steal: 从其他线程队列窃取任务来执行

### 3. 没有任务时进入事件轮训(epoll_wait)
可以看到上面4个队列都没有获取到任务后，会调用`sel.park(core)`让出cpu，等待有事件到来时恢复线程后继续寻找下一个等待执行的协程任务

park被层层包装后最终的实现是通过`epoll_wait`来等待事件
```rust
impl Park for runtime::Driver::Driver {
    fn park(&mut self) -> Result<(), Self::Error> {
        self.inner.park()      // call time driver's park
    }
}

impl<P> Park for time::driver::Driver<P> {
    fn park(&mut self) -> Result<(), Self::Error> {
        // ... preprocess for time
        // may call self.park.park_timeout(duration)?;
        self.park.park()?;     // call process driver's park

        self.handle.process();
    }
}

impl Park for process::unix::driver::Driver {
    fn park(&mut self) -> Result<(), Self::Error> {
        self.park.park()?;      // call signal driver's park
        self.inner.process();
        Ok(())
    }
}

impl Park for signal::unix::driver::Driver {
    fn park(&mut self) -> Result<(), Self::Error> {
        self.park.park()?;      // call io driver's park
        self.process();
        Ok(())
    }
}
impl Park for io::Driver {
    fn park(&mut self) -> io::Result<()> {
        self.turn(None)?;
        Ok(())
    }
}
```
继续调用
```rust

fn turn(&mut self, max_wait: Option<Duration>) -> io::Result<()> {
    // 省略了一些events处理
    // 重点就是这个
    match self.poll.poll(&mut events, max_wait) {
        Ok(_) => {}
        Err(ref e) if e.kind() == io::ErrorKind::Interrupted => {}
        Err(e) => return Err(e),
    }
    //省略了一些处理

    Ok(())
}

```
poll的实际实现是调用了mio的方法
```rust
pub fn poll(&mut self, events: &mut Events, timeout: Option<Duration>) -> io::Result<()> {
    self.registry.selector.select(events.sys(), timeout)
}
```
select函数里通过syscall调用了epoll_wait,并且传入的timeout=-1，会一直阻塞直到有事件到来
```rust
pub fn select(&self, events: &mut Events, timeout: Option<Duration>) -> io::Result<()> {
    //省略一些。。。

    let timeout = timeout
        .map(|to| cmp::min(to.as_millis(), MAX_SAFE_TIMEOUT) as libc::c_int)
        .unwrap_or(-1);

    events.clear();
    syscall!(epoll_wait(
        self.ep,
        events.as_mut_ptr(),
        events.capacity() as i32,
        timeout,
    ))
    .map(|n_events| {
        // This is safe because `epoll_wait` ensures that `n_events` are
        // assigned.
        unsafe { events.set_len(n_events as usize) };
    })
}

```

### 4. 事件到来后的协程唤醒
上面epoll_wait过后，有事件到来则会进行事件分发
```rust

#![allow(unused)]
fn main() {
match self.poll.poll(&mut events, max_wait)

for event in events.iter() {
    let token = event.token();

    if token != TOKEN_WAKEUP {
        self.dispatch(token, Ready::from_mio(event));
    }
}
}

```
接着就会进行事件处理，唤醒对应的线程，将task投递到队列中
```rust

#![allow(unused)]
fn main() {
// set_readiness:
let mut current = self.readiness.load(Acquire);

loop {
    let current_generation = GENERATION.unpack(current);

    // 1.
    if let Some(token) = token {
        if GENERATION.unpack(token) != current_generation {
            return Err(());
        }
    }

    // 2.
    let current_readiness = Ready::from_usize(current);
    let new = f(current_readiness);

    // 3.
    let packed = match tick {
        Tick::Set(t) => TICK.pack(t as usize, new.as_usize()),
        Tick::Clear(t) => {
            if TICK.unpack(current) as u8 != t {
                // Trying to clear readiness with an old event!
                return Err(());
            }

            TICK.pack(t as usize, new.as_usize())
        }
    };

    // 4.
    let next = GENERATION.pack(current_generation, packed);

    match self
        .readiness
        .compare_exchange(current, next, AcqRel, Acquire)
    {
        Ok(_) => return Ok(()),
        // we lost the race, retry!
        Err(actual) => current = actual,
    }
}
}

```
所有都处理完后，park函数返回到最开始调度循环，重新开始新的一轮任务处理



# 总结
rust的协程主要是靠编译器的状态机实现 + 三方的调度器实现

tokio调度器的逻辑也有些借鉴了golang的调度器

大体分析的设计就是这样，当然细节没有深挖都是非常多的，但不妨碍我们对rust协程实现的理解