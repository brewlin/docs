---
title: http_模块开发的步骤
toc: true
date: 2020-02-10 21:28:59
tags: [linux,c,nginx,ext]
---

# http 流程的生命周期
在开发模块前我们需要了解http流程的生命周期，然后确定我们需要扩展的功能在哪个阶段，最后才能在该阶段介入我们扩展的功能。

nginx 的http流程有11个阶段，每个阶段都可以认为是单独的模块在负责对应的职责，每个阶段的处理可能不止一次，可能会发生循环作用，这个完全由每个模块来决定对应的后续行为，如下为默认基础的11个生命周期:


```
# NGX_HTTP_POST_READ_PHASE = 0
在接收到完整的http头部后处理的http阶段

# NGX_HTTP_SERVER_REWRITE_PHASE 处理头部阶段
在还没有查询到uri匹配的location前，这时==rewrite重写url==也作为一个独立的http阶段

# NGX_HTTP_FIND_CONFIG_PHASE 寻找匹配的location
根据uri寻找匹配的location，这个阶段通常由ngx_http_core_module模块实现，不建议其他http模块重写定义这一阶段的行为

# NGX_HTTP_REWRITE_PHASE
在config_phase阶段之后重写url的意义与server_rewrite_phase阶段显然是不同的，因为这两者会导致查到不同的location快(location 是与uri进行匹配的)

# NGX_HTTP_POST_REWRITE_PHASE 重新查找到对应的uir匹配的location模块
这一阶段是由于rewrite重写url后会重新跳到ngx_http_find_config_phase阶段，找到与新的uri匹配的location，所以这一阶段是无法由第三方http模块处理的，而仅由ngx_http_core_module模块使用

# NGX_HTTP_PREACCESS_PHASE 处理access接入前的阶段
处理ngx_http_access_phase阶段前，http模块可以介入的处理阶段

# NGX_HTTP_ACCESS_PHASE 判断是否允许访问nginx服务器
这个阶段用于让http模块判断是否允许这个请求访问nginx服务器

# NGX_HTTP_POST_ACCESS_PHASE 构造拒绝服务响应给客户端，进行收尾
当ngx_http_access_phase阶段中http模块的handler处理方法返回不允许访问的错误码时，这个阶段将构造拒绝服务的用户响应。所以这个阶段实际上用于给ngx_http_access_phase阶段收尾

# NGX_HTTP_TRY_FILES_PHASE 专门针对try_files配置进行静态文件处理
这个阶段完全是为了try_files配置项而设立的，当http请求访问静态文件资源时,try_files配置项可以使这个请求顺序地访问多个静态文件资源，如果某一次访问失败，则继续访问try_files中指定的下一个静态资源，另外这个功能完全是在NGX_HTTP_TRY_FIELS_phase阶段中实现的

# NGX_HTTP_CONTENT_PHASE 处理http内容的阶段
用于处理http请求内容的阶段，这是大部分http模块最喜欢介入的阶段

# NGX_HTTP_LOG_PHASE 最后日志记录收尾的阶段
处理完请求后记录日志的阶段，例如NG_HTTP_LOG_MODULE 模块就在这个阶段中加入了一个handler处理方法，使得每个http请求处理完毕后惠济路access_log日志
```

# 一、扩展开发初始工作
扩展开发前，我们需要了解nginx编译的流程和工作模式

## 1.配置检查
在nginx编译的时候我们都会执行`./configure`命令，去检查平台编译环境后设置对应的宏命令支持对应的方法，并且`./configure`带有很多可选项参数，比如`--add-module=path` 就可以设定我们扩展目录的所在路径

简单讲下`--add-module=path`的工作原理：
```sh
//这里是 ./configure --add-module=path 命令会调用的脚本检查
//主要就是遍历 传入的path的路径，并去调用该路径下所有的config配置脚本
//所以在这里，我们就需要在我们的扩展目录下有一个config配置文件
if test -n "$NGX_ADDONS"; then

    echo configuring additional modules

    for ngx_addon_dir in $NGX_ADDONS
    do
        echo "adding module in $ngx_addon_dir"

        if test -f $ngx_addon_dir/config; then
            . $ngx_addon_dir/config

            echo " + $ngx_addon_name was configured"

        else
            echo "$0: error: no $ngx_addon_dir/config was found"
            exit 1
        fi
    done
fi
```
## 2.配置定义
如下是扩展开发的编译配置`path/to/modules/print/config`的具体内容:
```
ngx_addon_name = ngx_http_print_module
HTTP_MODULES="$HTTP_MODULES ngx_http_print_module"
NGX_ADDON_SRCS="$NGX_ADDON_SRCS $ngx_addon_dir/ngx_http_print_module.c"
```
主要有两个需要注意的 `ngx_addon_name` 仅在configure期间使用，设置模块名称

`HTTP_MODULES,NGX_ADDON_SRCS` 这两个变量可以看出都是追加操作，用于将我们的扩展代码文件追加到待编译的列表中，而且`HTTP_MODULES`用于告诉nginx这是一个http扩展，且扩展的入口是`ngx_http_pirnt_module`这个我们在代码中定义的`全局结构体变量`(它会被加入到http生命周期中去，且包含了扩展的入口，后文会详细涉及到)

## 3.扩展目录
这里我们以一个print模块为例，后面我们将介绍一个print模块的示例
```
- nginx-src
----------- auto
----------- conf
----------- man
----------- html
----------- src
----------- modules
------------------- print
------------------------- ngx_http_print_module.c
------------------------- config
```
这里样的目录方便我们后期扩展，上面讲到了`./configure --add-module=path/modules` 会去遍历该目录下所有的config，因此后期只需要在该目录下增加扩展目录即可

# 二、 http 基础模块的开发

## 1.定义主模块入口
该结构体在`configure`阶段的时候就会被写入到nginx全局的一个链表里，用于nginx启动时解析`nginx.conf`时遇到`http`配置项时会遍历调用所有相关的http模块

所以如下结构是作为一个扩展的启动入口
```c
//ngx_http_print_module.c

ngx_module_t ngx_http_print_module = {
    NGX_MODULE_V1,
    &ngx_http_print_module_ctx,
    ngx_http_print_commands,
    NGX_HTTP_MODULE,
    
    NULL,NULL,NULL,NULL,NULL,NULL,NULL,
    NGX_MODULE_V1_PADDING
};

```
## 2.定义配置存储解析入口
commands是一个很重要的结构体，同时也被`ngx_http_print_module`引用，用于匹配`nginx.conf`中的配置项

例如本例中，如果在`nginx.conf`中中出现`print test 1;`配置项，则nginx会通过本模块`ngx_http_print_module`在找到commands,并且会调用`set`回调方法,也就是下面定义的`ngx_http_print_register`方法去将`nginx.conf`中的值保存到我们自定义的结构体中，方便我们在模块内使用
```c
//ngx_http_print_module.c

//当解析nginx.conf 时 没发现一个如下的print配置，就会调用该set回调方法(ngx_http_print)
static ngx_command_t ngx_http_print_commands[] = {
    {
        ngx_string("print"), // The command name
        NGX_HTTP_LOC_CONF | NGX_HTTP_SRV_CONF| NGX_CONF_TAKE2,
        ngx_http_print_register, // The command handler
        NGX_HTTP_LOC_CONF_OFFSET,
        0,
        // offsetof(ngx_http_print_conf_t, str),
        NULL
    },
    ngx_null_command
};
```

## 3.定义存储配置的结构体
自定义的配置可以任何的设计，根据自己的场景来，本次例子只是简单的将`nginx.conf`中自定义的配置保存到如下的配置中去
```c
//ngx_http_print_module.c

typedef struct {
    ngx_str_t str;
    ngx_int_t num;
}ngx_http_print_conf_t
```
## 4.实现`ngx_http_print_register`保存配置
当nginx在启动阶段解析`nginx.conf`时，匹配到我们自定义的`command`则会调用对应的`set`回调函数，也就是`ngx_http_print_register`方法
```c
//ngx_http_print_module.c

//nginx.conf解析阶段，没发现一个匹配项就会调用当前函数注册相关handler 也就是请求处理的真正函数
static char * ngx_http_print_register(ngx_conf_t *cf,ngx_command_t *cmd,void *conf){
    printf("sfdsfs");
    ngx_http_core_loc_conf_t *clcf;

    clcf = ngx_http_conf_get_module_loc_conf(cf,ngx_http_core_module);
    //当http请求命中该配置后，会指行如下函数
    clcf->handler = ngx_http_print_handler;

    //解析conf中的 配置
    ngx_http_print_conf_t *cur_conf = conf;
    //是一个ngx_array_t 数组 保存着ngx解析nginx.conf中的配置参数
    ngx_str_t *value = cf->args->elts;
    cur_conf->str = value[1];
    if(cf->args->nelts > 2){
        cur_conf->num = ngx_atoi(value[2].data,value[2].len);
        if(cur_conf->num == NGX_ERROR){
            return "invalid number";
        }
    }

    return NGX_CONF_OK;

}
```
其实到这里还有个重要的东西没有注意到，函数的`void *conf`指向的是我们自定义的结构体，但是内存需要我们自己申请，那么什么时候申请呢，这就需要用到`ngx_http_module_t`的特性了

该结构用于我们监听框架初始化事件，当框架启动扫描时会调用我们模块自定义的事件，并且会多次调用
```
例如：
http{
    print a 1;
    server {
        print b 2;
        location / {
            print c 3;
        }
        location /test {
            print d 4;
        }
    }
}
```
如上就会调用4次我们自定义的函数,用于处理相关参数，也包括我们会提前预分配好内存保存
### 4.1 绑定请求事件
上面可以看到我们设置了这一行代码
```
clcf->handler = ngx_http_print_handler;
```
其实这行就是重点，该函数是nginx运行时请求真正的处理函数

也就是当有配置命中了我们的模块，那么自定义的函数会被介入到http框架的11个生命周期中，进行http请求处理

详情请看下面 第三部分 `模块请求入口函数` 的实现
 
## 5.定义框架初始化事件
上面可以看到`ngx_http_print_module_ctx`是一个自定义结构体，如下
```c
//ngx_http_print_module.c

static ngx_http_module_t ngx_http_print_module_ctx = {
    NULL,//解析配置前
    NULL,//解析配置后
    
    NULL,//解析http配置
    NULL,//合并http配置

    NULL,//解析server配置
    NULL,//合并server配置

    create_loc_conf, //解析location配置
    NULL//合并location配置
};
```
其实一个`ngx_http_module`结构体，包含了8个回调函数，分别是`ngin.conf`被解析时会调用的函数，需要我们自己实现

其实上面我们只实现了一个`create_loc_conf`方法，因为在解析配置前其实是需要提前分配好需要解析的配置保存的内存，所以这就是我们准备要做的工作

## 6.预分配自定义结构体内存
```
//ngx_http_print_module.c

static void * create_loc_conf(ngx_conf_t *cf){
    ngx_http_print_conf_t *conf;
    conf = ngx_pcalloc(cf->pool,sizeof(ngx_http_print_conf_t));
    if(conf == NULL){
        return NGX_CONF_ERROR;
    }
    conf->str.data = NULL;
    conf->str.len = 0;
    conf->num = 0;
    return conf;
}
```

# 三、模块请求入口函数
这个便是本文的重点，充当了http 请求处理的角色，当有请求命中了我们定义的配置项，则如下函数会介入到请求处理生命周期中去
```c
//ngx_http_print_module.c

//作为http生命周期阶段的一部分 处理该请求
static ngx_int_t ngx_http_print_handler(ngx_http_request_t *r){
    if(!(r->method & NGX_HTTP_GET)){
        return NGX_HTTP_NOT_ALLOWED;
    }

    //不处理包体，直接通知不在接受客户端传递数据
    //这行看似可有可无，其实是当我们不处理缓存区数据，万一客户端继续发送可能会导致超时
    ngx_int_t rc = ngx_http_discard_request_body(r);
    if(rc != NGX_OK){
        return rc;
    }
    //返回响应
    ngx_str_t type = ngx_string("application/json");
    ngx_str_t response = ngx_string(" the print module");
    //设置状态码
    r->headers_out.status = NGX_HTTP_OK;
    //设置响应包长度
    r->headers_out.content_length_n = response.len;
    //设置content-type
    r->headers_out.content_type = type;
    
    //发送http头部
    rc = ngx_http_send_header(r);
    if(rc == NGX_ERROR || rc > NGX_OK || r->header_only){
        return rc;
    }

    //构造ngx_buf_t 结构体准备发送包体
    ngx_buf_t *b;
    b = ngx_create_temp_buf(r->pool,response.len)
    if(b == NULL){
        return NGX_HTTP_INTERNAL_SERVER_ERROR;
    }
    ngx_memcpy(b->pos,response.data,response.len);
    b->last = b->post + response.len;
    //表明这是最后一块缓冲区
    b->last_buf = 1;
    
    ngx_chain_t out;
    out.buf = b;
    out.next = NULL;

    //发送包体
    return ngx_http_output_filter(r,&out);
}
```
## 1.响应普通文本
这个就是普通的字符串数据响应给客户端方式,本文的案例也是用的这种，返回一个普通字符数据给客户端
```c
    //构造ngx_buf_t 结构体准备发送包体
    ngx_buf_t *b;
    b = ngx_create_temp_buf(r->pool,response.len)
    if(b == NULL){
        return NGX_HTTP_INTERNAL_SERVER_ERROR;
    }
    ngx_memcpy(b->pos,response.data,response.len);
    b->last = b->post + response.len;
    //表明这是最后一块缓冲区
    b->last_buf = 1;
    
    ngx_chain_t out;
    out.buf = b;
    out.next = NULL;
```
同时也可以将本地文件内容读取后返回给客户端，下面7.2的方法就可以做到


## 2.响应本地磁盘文件
分别定义了发送文件响应的方法

设置文件回收的事件方法，防止内存泄漏或者文件占用
```c
   u_char* filename = (u_char*)"/tmp/print.html";
    //告诉nginx 实际响应的内容从文件中获取
    b->in_file = 1;
    b->file = ngx_pcalloc(r->pool,sizeof(ngx_file_t));
    b->file->fd = ngx_open_file(filename,NGX_FILE_RDONLY|NGX_FILE_NONBLOCK,NGX_FILE_OPEN,0);
    b->file->log = r->connection->log;
    b->file->name.data = filename;
    b->file->name.len = sizeof(filename) -1;
    if(b->file->fd <= 0 ){
        return NGX_HTTP_NOT_FOUND;
    }
    if(ngx_file_info(filename,&b->file->info) == NGX_FILE_ERROR){
        return NGX_HTTP_INTERNAL_SERVER_ERROR;
    }
    r->headers_out.content_length_n = b->file->info.st_size;
    b->file_pos = 0;
    b->file_last = b->file->info.st_size;

    ngx_pool_cleanup_t* cl = ngx_pool_cleanup_add(r->pool,sizeof(ngx_pool_cleanup_file_t));
    if(cl == NULL){
        return  NGX_ERROR;
    }
    cl->handler = ngx_pool_cleanup_file;
    ngx_pool_cleanup_file_t *clnf = cl->data;
    clnf->fd = b->file->fd;
    clnf->name = b->file->name.data;
    clnf->log = r->pool->log;
```

# 四、code & 总结
总的来说，每个部分nginx都提供了非常多的功能和api，本文只是简单的实现了一个从配置定义、配置触发自定义函数、以及介入http请求，响应http等例子介绍了一个nginx扩展的开发

## 编译
```
./configure --add-moudle=/path/to/print
make
make install
```
## nginx.conf
```
http{
    server{
        listen 8081;
        
        location / {
            print test 2;
        }
    }
}
```
## config
```
ngx_addon_name = ngx_http_print_module
HTTP_MODULES="$HTTP_MODULES ngx_http_print_module"
NGX_ADDON_SRCS="$NGX_ADDON_SRCS $ngx_addon_dir/ngx_http_print_module.c"
```
## code 
```
#include "nginx.h"
#include "ngx_config.h"
#include "ngx_core.h"
#include "ngx_http.h"


static char * ngx_http_print_register(ngx_conf_t *cf,ngx_command_t *cmd,void *conf);
static ngx_int_t ngx_http_print_handler(ngx_http_request_t *r);
static void * create_loc_conf(ngx_conf_t *cf);
static void * create_serv_conf(ngx_conf_t *cf);


typedef struct {
    ngx_str_t str;
    ngx_int_t num;
}ngx_http_print_conf_t;

//当解析nginx.conf 时 没发现一个如下的print配置，就会调用该set回调方法(ngx_http_print)
static ngx_command_t ngx_http_print_commands[] = {
    {
        ngx_string("print"), // The command name
        NGX_HTTP_LOC_CONF | NGX_HTTP_SRV_CONF| NGX_CONF_TAKE2,
        ngx_http_print_register, // The command handler
        NGX_HTTP_LOC_CONF_OFFSET,
        0,
        // offsetof(ngx_http_print_conf_t, str),
        NULL
    },
    ngx_null_command
};
//nginx.conf解析阶段，没发现一个匹配项就会调用当前函数注册相关handler 也就是请求处理的真正函数
static char * ngx_http_print_register(ngx_conf_t *cf,ngx_command_t *cmd,void *conf){
    printf("sfdsfs");
    ngx_http_core_loc_conf_t *clcf;

    clcf = ngx_http_conf_get_module_loc_conf(cf,ngx_http_core_module);
    //当http请求命中该配置后，会指行如下函数
    clcf->handler = ngx_http_print_handler;

    //解析conf中的 配置
    ngx_http_print_conf_t *cur_conf = conf;
    //是一个ngx_array_t 数组 保存着ngx解析nginx.conf中的配置参数
    ngx_str_t *value = cf->args->elts;
    cur_conf->str = value[1];
    if(cf->args->nelts > 2){
        cur_conf->num = ngx_atoi(value[2].data,value[2].len);
        if(cur_conf->num == NGX_ERROR){
            return "invalid number";
        }
    }


    return NGX_CONF_OK;

}
// 在nginx启动，也就是框架初始化时会调用如下的自定义模块的回调函数
//如果没有什么需要做的，就不需要实现相关函数
static ngx_http_module_t ngx_http_print_module_ctx = {
    NULL,
    NULL,
    NULL,
    NULL,

    NULL,
    NULL,

    create_loc_conf,
    NULL
};
ngx_module_t ngx_http_print_module = {
    NGX_MODULE_V1,
    &ngx_http_print_module_ctx,
    ngx_http_print_commands,
    NGX_HTTP_MODULE,
    
    NULL,NULL,NULL,NULL,NULL,NULL,NULL,
    NGX_MODULE_V1_PADDING
};
static void * create_loc_conf(ngx_conf_t *cf){
    ngx_http_print_conf_t *conf;
    conf = ngx_pcalloc(cf->pool,sizeof(ngx_http_print_conf_t));
    if(conf == NULL){
        return NGX_CONF_ERROR;
    }
    conf->str.data = NULL;
    conf->str.len = 0;
    conf->num = 0;
    return conf;
}
static void * create_serv_conf(ngx_conf_t *cf){
    ngx_http_print_conf_t *conf;
    conf = ngx_pcalloc(cf->pool,sizeof(ngx_http_print_conf_t));
    if(conf == NULL){
        return NGX_CONF_ERROR;
    }
    conf->str.data = NULL;
    conf->str.len = 0;
    conf->num = 0;
    return conf;
}
static ngx_int_t  response_file(ngx_http_request_t *r){
    ngx_str_t type = ngx_string("application/json");
    u_char* filename = (u_char*)"/tmp/print.html";
    ngx_buf_t *b = ngx_palloc(r->pool,sizeof(ngx_buf_t));

    //设置状态码
    r->headers_out.status = NGX_HTTP_OK;
    //设置content-type
    r->headers_out.content_type = type;

    //告诉nginx 实际响应的内容从文件中获取
    b->in_file = 1;
    b->file = ngx_pcalloc(r->pool,sizeof(ngx_file_t));
    b->file->fd = ngx_open_file(filename,NGX_FILE_RDONLY|NGX_FILE_NONBLOCK,NGX_FILE_OPEN,0);
    b->file->log = r->connection->log;
    b->file->name.data = filename;
    b->file->name.len = sizeof(filename) -1;
    if(b->file->fd <= 0 ){
        return NGX_HTTP_NOT_FOUND;
    }
    if(ngx_file_info(filename,&b->file->info) == NGX_FILE_ERROR){
        return NGX_HTTP_INTERNAL_SERVER_ERROR;
    }
    r->headers_out.content_length_n = b->file->info.st_size;
    b->file_pos = 0;
    b->file_last = b->file->info.st_size;
    //发送http头部
    ngx_int_t rc = ngx_http_send_header(r);
    if(rc == NGX_ERROR || rc > NGX_OK || r->header_only){
        return rc;
    }

    ngx_pool_cleanup_t* cl = ngx_pool_cleanup_add(r->pool,sizeof(ngx_pool_cleanup_file_t));
    if(cl == NULL){
        return  NGX_ERROR;
    }
    cl->handler = ngx_pool_cleanup_file;
    ngx_pool_cleanup_file_t *clnf = cl->data;
    clnf->fd = b->file->fd;
    clnf->name = b->file->name.data;
    clnf->log = r->pool->log;

    ngx_http_print_conf_t *cf = (ngx_http_print_conf_t*)r->loc_conf[0];

    ngx_chain_t out;
    out.buf = b;
    out.next = NULL;

    //发送包体
    return ngx_http_output_filter(r,&out);
}
static ngx_int_t response_str(ngx_http_request_t *r){



    ngx_str_t type = ngx_string("application/json");
    ngx_str_t response = ngx_string(" the print module");
        //设置状态码
    r->headers_out.status = NGX_HTTP_OK;
    //设置content-type
    r->headers_out.content_type = type;
    //设置状态码
    r->headers_out.status = NGX_HTTP_OK;
    //设置响应包长度
    r->headers_out.content_length_n = response.len;
    //设置content-type
    r->headers_out.content_type = type;
    //构造ngx_buf_t 结构体准备发送包体
            //发送http头部
    ngx_int_t rc = ngx_http_send_header(r);
    if(rc == NGX_ERROR || rc > NGX_OK || r->header_only){
        return rc;
    }



    ngx_buf_t *b = ngx_create_temp_buf(r->pool,response.len);
    if(b == NULL){
        return NGX_HTTP_INTERNAL_SERVER_ERROR;
    }
    ngx_memcpy(b->pos,response.data,response.len);
    b->last = b->pos + response.len;
    //表明这是最后一块缓冲区
    b->last_buf = 1;

    ngx_chain_t out;
    out.buf = b;
    out.next = NULL;

    //发送包体
    return ngx_http_output_filter(r,&out);
}

//作为http生命周期阶段的一部分 处理该请求
static ngx_int_t ngx_http_print_handler(ngx_http_request_t *r){
    if(!(r->method & NGX_HTTP_GET)){
        return NGX_HTTP_NOT_ALLOWED;
    }

    //不处理包体，直接通知不在接受客户端传递数据
    //这行看似可有可无，其实是当我们不处理缓存区数据，万一客户端继续发送可能会导致超时
    ngx_int_t rc = ngx_http_discard_request_body(r);
    if(rc != NGX_OK){
        return rc;
    }
    return response_str(r);
    // return response_file(r);
}
```