---
title: shamem.h
toc: true
date: 2019-12-11 21:28:59
tags: [linux,share,memory,syscall]
---
```c
#ifndef _SHMEM_H_INCLUDED_
#define _SHMEM_H_INCLUDED_

#define  OK          0
#define  ERROR      -1


#include <sys/mman.h>
#include "server.h"
typedef server serv;

typedef struct {
    u_char      *addr;
    size_t       size;
    u_char    name;
    serv   *server;
} shm_t;


int shm_alloc(shm_t *shm);
void shm_free(shm_t *shm);


#endif /* _SHMEM_H_INCLUDED_ */

```