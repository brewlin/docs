---
title: php_引用计数与gc
toc: true
date: 2020-01-10 21:28:59
tags: [linux,c,php,ext,refrerence,gc]
---

进行php扩展开发的时候会遇到一些问题，就是php用户态空间将变量传递到扩展层面`c层调用`的时候，会出现一些问题，下面的例子是一个timer定时器的例子。
用户空间会传递一个回调函数给`timer扩展接口`，那么实际回调函数被调用的地方是`c层`。但是该回调函数变量本身是由用户空间`申请`并交由php`内核gc管理`的，
如果扩展函数内不做任何操作，那么当切换到用户空间时php内核会判断该变量`需要回收`，然后扩展函数就会空指针异常等

当扩展函数内该php变量生命周期使用结束后，任然需要考虑`垃圾回收`的问题，并不是在扩展函数内简单`free(data)`就可以的，需要调用php内核引用计数接口等
进行变量的回收以及gc等，最后交由php内核gc管理。当然扩展函数内由c自行申请管理的内存可以自己释放

- 扩展函数定义示例
```c
PHP_METHOD(timer_obj,tick)
{
    php_lib_timer_callback *fci = (php_lib_timer_callback *)malloc(sizeof(php_lib_timer_callback));
    //强制传入两个参数
    ZEND_PARSE_PARAMETERS_START(2, 2)
    Z_PARAM_LONG(fci->seconds)
    Z_PARAM_FUNC(fci->fci,fci->fcc)
    ZEND_PARSE_PARAMETERS_END_EX(RETURN_FALSE);

    long id = create_time_event(fci->seconds,tick,fci,del);
    zend_fci_cache_persist(&fci->fcc);
    RETURN_LONG(id);
}
```
- 扩展函数内执行php用户态回调函数示例
```c
int tick(long long id,void *data){

    php_lib_timer_callback *fci = (php_lib_timer_callback *)data;
    zval result;
    fci->fci.retval = &result;
    if(zend_call_function(&fci->fci,&fci->fcc) != SUCCESS){
        return NOMORE;
    }
    return fci->seconds;
}
```
## timer 中对回调函数变量进行引用计数+1
上面会发现timer::tick()函数在返回给用户空间时会做一个操作`    zend_fci_cache_persist(&fci->fcc);`，正是该调用
对传入的回调函数进行饮用计数管理，`告诉php内核该回调函数在c层会继续使用不用回收`。代码如下
```c
static void zend_fci_cache_persist(zend_fcall_info_cache *fci_cache)
{
    if (fci_cache->object)
    {
        GC_ADDREF(fci_cache->object);
    }
    if (fci_cache->function_handler->op_array.fn_flags & ZEND_ACC_CLOSURE)
    {
        GC_ADDREF(ZEND_CLOSURE_OBJECT(fci_cache->function_handler));
    }
}

```

其中`GC_ADDREF（）`函数很明显就是内核GC相关的api。`fci_cache->function_handler` 则为用户传递的回调函数真正的变量地址

如上前奏后就可以在c扩展中放心的对用户传递的变量进行操作了

## timer 中结束后变量的Gc回收
上面有看到`php_lib_timer_callback`变量实际是自己定义的结构体，包括内存也是有开发者自己分配的，可以放心的`free`。但是该结构体中
指向的`fci.fcc` 则实际是php用户空间申请的变量，不能直接`free`,如果直接free，会引发php gc泄漏，如下警告所示:
```c
> php timer.php
/php/src/Zend/zend_closures.c(459) :  Freeing 0x00007fc084e6d480 (304 bytes), script=/timer.php
=== Total 1 memory leaks detected ===
```
所以依然需要根据php内核GC的管理方式来处理用户空间的变量，也就是模拟用户空间那样对变量的管理：
```c++
static void zend_fci_cache_discard(zend_fcall_info_cache *fci_cache)
{
    if (fci_cache->object) {
        OBJ_RELEASE(fci_cache->object);
    }
    if (fci_cache->function_handler->op_array.fn_flags & ZEND_ACC_CLOSURE) {
        OBJ_RELEASE(ZEND_CLOSURE_OBJECT(fci_cache->function_handler));

    }
}
static void zend_fci_params_discard(zend_fcall_info *fci)
{
    if (fci->param_count > 0)
    {
        uint32_t i;
        for (i = 0; i < fci->param_count; i++)
        {
            zval_ptr_dtor(&fci->params[i]);
        }
        efree(fci->params);
    }
}
```
在不需要使用的时候，也需要对回调函数本身进行减引用，以及回调函数内的用户态的参数进行减引用以及变量的回收。只有做完上面这些基本的管理才能
开发一个安全的扩展函数
## 总结
总之用户空间申请的变量传递给扩展内函数使用，如果在返回给用户空间后依然会继续使用就要zval_copy或者引用计数+1,因为在返回给用户空间的时候本身用户空间gc会判断该变量是否有继续引用，否则就`refcount -= 1`，用户空间回收该变量，但是扩展函数内依然在访问该已经被销毁的变量。就会导致错误

只有将这些变量的引用与回收做好了才能开发出安全可靠的扩展函数
