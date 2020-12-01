---
title: 动手写RPC框架 - GeeRPC第六天 负载均衡(load balance)
date: 2020-10-08 14:00:00
description: 7天用 Go语言/golang 从零实现 RPC 框架 GeeRPC 教程(7 days implement golang remote procedure call framework from scratch tutorial)，动手写 RPC 框架，参照 golang 标准库 net/rpc 的实现，实现了服务端(server)、支持异步和并发的客户端(client)、消息编码与解码(message encoding and decoding)、服务注册(service register)、支持 TCP/Unix/HTTP 等多种传输协议。第六天实现了2种简单的负载均衡(load balance)算法，随机选择和 Round Robin 轮询调度算法。
tags:
- Go
nav: 从零实现
categories:
- RPC框架 - GeeRPC
keywords:
- Go语言
- 从零实现RPC框架
- 负载均衡
- 轮询调度
image: post/geerpc/geerpc.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day6 负载均衡
---

![golang RPC framework](geerpc/geerpc.jpg)

本文是[7天用Go从零实现RPC框架GeeRPC](https://geektutu.com/post/geerpc.html)的第六篇。

- 通过随机选择和 Round Robin 轮询调度算法实现服务端负载均衡，代码约 250 行

## 负载均衡策略

假设有多个服务实例，每个实例提供相同的功能，为了提高整个系统的吞吐量，每个实例部署在不同的机器上。客户端可以选择任意一个实例进行调用，获取想要的结果。那如何选择呢？取决了负载均衡的策略。对于 RPC 框架来说，我们可以很容易地想到这么几种策略：

- 随机选择策略 - 从服务列表中随机选择一个。
- 轮询算法(Round Robin) - 依次调度不同的服务器，每次调度执行 i = (i + 1) mode n。
- 加权轮询(Weight Round Robin) - 在轮询算法的基础上，为每个服务实例设置一个权重，高性能的机器赋予更高的权重，也可以根据服务实例的当前的负载情况做动态的调整，例如考虑最近5分钟部署服务器的 CPU、内存消耗情况。
- 哈希/一致性哈希策略 - 依据请求的某些特征，计算一个 hash 值，根据 hash 值将请求发送到对应的机器。一致性 hash 还可以解决服务实例动态添加情况下，调度抖动的问题。一致性哈希的一个典型应用场景是分布式缓存服务。感兴趣可以阅读[动手写分布式缓存 - GeeCache第四天 一致性哈希(hash)](https://geektutu.com/post/geecache-day4.html)
- ...

## 服务发现

负载均衡的前提是有多个服务实例，那我们首先实现一个最基础的服务发现模块 Discovery。为了与通信部分解耦，这部分的代码统一放置在 xclient 子目录下。

定义 2 个类型：

- SelectMode 代表不同的负载均衡策略，简单起见，GeeRPC 仅实现 Random 和 RoundRobin 两种策略。
- Discovery 是一个接口类型，包含了服务发现所需要的最基本的接口。
    - Refresh() 从注册中心更新服务列表
    - Update(servers []string) 手动更新服务列表
    - Get(mode SelectMode) 根据负载均衡策略，选择一个服务实例
    - GetAll() 返回所有的服务实例

[day6-load-balance/xclient/discovery.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day6-load-balance)

```go
package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect     SelectMode = iota // select randomly
	RoundRobinSelect                   // select using Robbin algorithm
)

type Discovery interface {
	Refresh() error // refresh from remote registry
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}
```

紧接着，我们实现一个不需要注册中心，服务列表由手工维护的服务发现的结构体：MultiServersDiscovery

```go
// MultiServersDiscovery is a discovery for multi servers without a registry center
// user provides the server addresses explicitly instead
type MultiServersDiscovery struct {
	r       *rand.Rand   // generate random number
	mu      sync.RWMutex // protect following
	servers []string
	index   int // record the selected position for robin algorithm
}

// NewMultiServerDiscovery creates a MultiServersDiscovery instance
func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}
```

- r 是一个产生随机数的实例，初始化时使用时间戳设定随机数种子，避免每次产生相同的随机数序列。
- index 记录 Round Robin 算法已经轮询到的位置，为了避免每次从 0 开始，初始化时随机设定一个值。

然后，实现 Discovery 接口

```go
var _ Discovery = (*MultiServersDiscovery)(nil)

// Refresh doesn't make sense for MultiServersDiscovery, so ignore it
func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

// Update the servers of discovery dynamically if needed
func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

// Get a server according to mode
func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n] // servers could be updated, so mode n to ensure safety
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

// returns all servers in discovery
func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	// return a copy of d.servers
	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
```

## 支持负载均衡的客户端

接下来，我们向用户暴露一个支持负载均衡的客户端 XClient。

[day6-load-balance/xclient/xclient.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day6-load-balance)

```go
package xclient

import (
	"context"
	. "geerpc"
	"io"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *Option
	mu      sync.Mutex // protect following
	clients map[string]*Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt, clients: make(map[string]*Client)}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		// I have no idea how to deal with error, just ignore it.
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}
```

XClient 的构造函数需要传入三个参数，服务发现实例 Discovery、负载均衡模式 SelectMode 以及协议选项 Option。为了尽量地复用已经创建好的 Socket 连接，使用 clients 保存创建成功的 Client 实例，并提供 Close 方法在结束后，关闭已经建立的连接。

接下来，实现客户端最基本的功能 `Call`。

```go
func (xc *XClient) dial(rpcAddr string) (*Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}
	if client == nil {
		var err error
		client, err = XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	return client, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

// Call invokes the named function, waits for it to complete,
// and returns its error status.
// xc will choose a proper server.
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}
```

我们将复用 Client 的能力封装在方法 `dial` 中，dial 的处理逻辑如下：

1) 检查 `xc.clients` 是否有缓存的 Client，如果有，检查是否是可用状态，如果是则返回缓存的 Client，如果不可用，则从缓存中删除。
2) 如果步骤 1) 没有返回缓存的 Client，则说明需要创建新的 Client，缓存并返回。

另外，我们为 XClient 添加一个常用功能：`Broadcast`。

```go
// Broadcast invokes the named function for every server registered in discovery
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex // protect e and replyDone
	var e error
	replyDone := reply == nil // if reply is nil, don't need to set value
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel() // if any call failed, cancel unfinished calls
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
```

Broadcast 将请求广播到所有的服务实例，如果任意一个实例发生错误，则返回其中一个错误；如果调用成功，则返回其中一个的结果。有以下几点需要注意：

1) 为了提升性能，请求是并发的。
2) 并发情况下需要使用互斥锁保证 error 和 reply 能被正确赋值。
3) 借助 context.WithCancel 确保有错误发生时，快速失败。

## Demo

又到了 Demo 环节，我们还是借助一个简单的 Demo 验证今天的成果吧。

首先，启动 RPC 服务的代码还是类似的，Sum 是正常的方法，Sleep 用于验证 XClient 的超时机制能否正常运作。

[day6-load-balance/main/main.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day6-load-balance)

```go
package main

import (
	"context"
	"geerpc"
	"geerpc/xclient"
	"log"
	"net"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addrCh chan string) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := geerpc.NewServer()
	_ = server.Register(&foo)
	addrCh <- l.Addr().String()
	server.Accept(l)
}
```

封装一个方法 `foo`，便于在 `Call` 或 `Broadcast` 之后统一打印成功或失败的日志。

```go
func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}
```

call 调用单个服务实例，broadcast 调用所有服务实例

```go
func call(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// expect 2 - 5 timeout
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}


func main() {
	log.SetFlags(0)
	ch1 := make(chan string)
	ch2 := make(chan string)
	// start two servers
	go startServer(ch1)
	go startServer(ch2)

	addr1 := <-ch1
	addr2 := <-ch2

	time.Sleep(time.Second)
	call(addr1, addr2)
	broadcast(addr1, addr2)
}
```

运行结果如下：

```go
rpc server: register Foo.Sleep
rpc server: register Foo.Sum
rpc server: register Foo.Sleep
rpc server: register Foo.Sum
call Foo.Sum success: 4 + 16 = 20
call Foo.Sum success: 0 + 0 = 0
call Foo.Sum success: 3 + 9 = 12
call Foo.Sum success: 2 + 4 = 6
call Foo.Sum success: 1 + 1 = 2
broadcast Foo.Sum success: 3 + 9 = 12
broadcast Foo.Sum success: 1 + 1 = 2
broadcast Foo.Sum success: 0 + 0 = 0
broadcast Foo.Sum success: 4 + 16 = 20
broadcast Foo.Sum success: 2 + 4 = 6
broadcast Foo.Sleep success: 0 + 0 = 0
broadcast Foo.Sleep success: 1 + 1 = 2
broadcast Foo.Sleep error: rpc client: call failed: context deadline exceeded
broadcast Foo.Sleep error: rpc client: call failed: context deadline exceeded
broadcast Foo.Sleep error: rpc client: call failed: context deadline exceeded
```

## 附 推荐阅读

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go 语言笔试面试题](https://geektutu.com/post/qa-golang.html)