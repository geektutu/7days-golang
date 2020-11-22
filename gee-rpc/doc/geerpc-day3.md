---
title: 动手写RPC框架 - GeeRPC第三天 服务注册(service register)
date: 2020-10-07 19:00:00
description: 7天用 Go语言/golang 从零实现 RPC 框架 GeeRPC 教程(7 days implement golang remote procedure call framework from scratch tutorial)，动手写 RPC 框架，参照 golang 标准库 net/rpc 的实现，实现了服务端(server)、支持异步和并发的客户端(client)、消息编码与解码(message encoding and decoding)、服务注册(service register)、支持 TCP/Unix/HTTP 等多种传输协议。第三天实现了服务注册，即将 Go 语言结构体通过反射映射为服务。
tags:
- Go
nav: 从零实现
categories:
- RPC框架 - GeeRPC
keywords:
- Go语言
- 从零实现RPC框架
- 反射
- 服务
image: post/geerpc/geerpc.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day3 服务注册
---

![golang RPC framework](geerpc/geerpc.jpg)

本文是[7天用Go从零实现RPC框架GeeRPC](https://geektutu.com/post/geerpc.html)的第三篇。

- 通过反射实现服务注册功能
- 在服务端实现服务调用，代码约 150 行

## 结构体映射为服务

RPC 框架的一个基础能力是：像调用本地程序一样调用远程服务。那如何将程序映射为服务呢？那么对 Go 来说，这个问题就变成了如何将结构体的方法映射为服务。

对 `net/rpc` 而言，一个函数需要能够被远程调用，需要满足如下五个条件：

- the method's type is exported. -- 方法所属类型是导出的。
- the method is exported. -- 方式是导出的。
- the method has two arguments, both exported (or builtin) types. -- 两个入参，均为导出或内置类型。
- the method's second argument is a pointer. -- 第二个入参必须是一个指针。
- the method has return type error. -- 返回值为 error 类型。

更直观一些：

```go
func (t *T) MethodName(argType T1, replyType *T2) error
```

假设客户端发过来一个请求，包含 ServiceMethod 和 Argv。

```json
{
    "ServiceMethod"： "T.MethodName"
    "Argv"："0101110101..." // 序列化之后的字节流
}
```

通过 "T.MethodName" 可以确定调用的是类型 T 的 MethodName，如果硬编码实现这个功能，很可能是这样：

```go
switch req.ServiceMethod {
    case "T.MethodName":
        t := new(t)
        reply := new(T2)
        var argv T1
        gob.NewDecoder(conn).Decode(&argv)
        err := t.MethodName(argv, reply)
        server.sendMessage(reply, err)
    case "Foo.Sum":
        f := new(Foo)
        ...
}
```

也就是说，如果使用硬编码的方式来实现结构体与服务的映射，那么每暴露一个方法，就需要编写等量的代码。那有没有什么方式，能够将这个映射过程自动化呢？可以借助反射。

通过反射，我们能够非常容易地获取某个结构体的所有方法，并且能够通过方法，获取到该方法所有的参数类型与返回值。例如：

```go
func main() {
	var wg sync.WaitGroup
	typ := reflect.TypeOf(&wg)
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		argv := make([]string, 0, method.Type.NumIn())
		returns := make([]string, 0, method.Type.NumOut())
		// j 从 1 开始，第 0 个入参是 wg 自己。
		for j := 1; j < method.Type.NumIn(); j++ {
			argv = append(argv, method.Type.In(j).Name())
		}
		for j := 0; j < method.Type.NumOut(); j++ {
			returns = append(returns, method.Type.Out(j).Name())
		}
		log.Printf("func (w *%s) %s(%s) %s",
			typ.Elem().Name(),
			method.Name,
			strings.Join(argv, ","),
			strings.Join(returns, ","))
    }
}
```

运行的结果是：

```go
func (w *WaitGroup) Add(int)
func (w *WaitGroup) Done()
func (w *WaitGroup) Wait()
```

## 通过反射实现 service

前面两天我们完成了客户端和服务端，客户端相对来说功能是比较完整的，但是服务端的功能并不完整，仅仅将请求的 header 打印了出来，并没有真正地处理。那今天的主要目的是补全这部分功能。首先通过反射实现结构体与服务的映射关系，代码独立放置在 `service.go` 中。

[day3-service/service.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day3-service)

第一步，定义结构体 methodType：

```go
type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	// arg may be a pointer type, or a value type
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	// reply must be a pointer type
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}
```

每一个 methodType 实例包含了一个方法的完整信息。包括

- method：方法本身
- ArgType：第一个参数的类型
- ReplyType：第二个参数的类型
- numCalls：后续统计方法调用次数时会用到

另外，我们还实现了 2 个方法 `newArgv` 和 `newReplyv`，用于创建对应类型的实例。`newArgv` 方法有一个小细节，指针类型和值类型创建实例的方式有细微区别。

第二步，定义结构体 service：

```go
type service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value
	method map[string]*methodType
}
```

service 的定义也是非常简洁的，name 即映射的结构体的名称，比如 `T`，比如 `WaitGroup`；typ 是结构体的类型；rcvr 即结构体的实例本身，保留 rcvr 是因为在调用时需要 rcvr 作为第 0 个参数；method 是 map 类型，存储映射的结构体的所有符合条件的方法。

接下来，完成构造函数 `newService`，入参是任意需要映射为服务的结构体实例。

```go
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}
```

`registerMethods` 过滤出了符合条件的方法：

- 两个导出或内置类型的入参（反射时为 3 个，第 0 个是自身，类似于 python 的 self，java 中的 this）
- 返回值有且只有 1 个，类型为 error

最后，我们还需要实现 `call` 方法，即能够通过反射值调用方法。

```go
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
```

## service 的测试用例

为了保证 service 实现的正确性，我们为 service.go 写了几个测试用例。

[day3-service/service_test.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day3-service)

定义结构体 Foo，实现 2 个方法，导出方法 Sum 和 非导出方法 sum。

```go
type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// it's not a exported Method
func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}
```

测试 newService 和 call 方法。

```go
func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	_assert(len(s.method) == 1, "wrong service Method, expect 1, but got %d", len(s.method))
	mType := s.method["Sum"]
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]

	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.call(mType, argv, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 1, "failed to call Foo.Sum")
}
```

## 集成到服务端

通过反射结构体已经映射为服务，但请求的处理过程还没有完成。从接收到请求到回复还差以下几个步骤：第一步，根据入参类型，将请求的 body 反序列化；第二步，调用 `service.call`，完成方法调用；第三步，将 reply 序列化为字节流，构造响应报文，返回。

回到代码本身，补全之前在 `server.go` 中遗留的 2 个 TODO 任务 `readRequest` 和 `handleRequest` 即可。

在这之前，我们还需要为 Server 实现一个方法 `Register`。

[day3-service/server.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day3-service)

```go
// Server represents an RPC Server.
type Server struct {
	serviceMap sync.Map
}

// Register publishes in the server the set of methods of the
func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }
```

配套实现 `findService` 方法，即通过 `ServiceMethod` 从 serviceMap 中找到对应的 service

```go
func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}
```

`findService` 的实现看似比较繁琐，但是逻辑还是非常清晰的。因为 ServiceMethod 的构成是 "Service.Method"，因此先将其分割成 2 部分，第一部分是 Service 的名称，第二部分即方法名。现在 serviceMap 中找到对应的 service 实例，再从 service 实例的 method 中，找到对应的 methodType。

准备工具已经就绪，我们首先补全 readRequest 方法：

```go
// request stores all information of a call
type request struct {
	h            *codec.Header // header of request
	argv, replyv reflect.Value // argv and replyv of request
	mtype        *methodType
	svc          *service
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	// make sure that argvi is a pointer, ReadBody need a pointer as parameter
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}
```

readRequest 方法中最重要的部分，即通过 `newArgv()` 和 `newReplyv()` 两个方法创建出两个入参实例，然后通过 `cc.ReadBody()` 将请求报文反序列化为第一个入参 argv，在这里同样需要注意 argv 可能是值类型，也可能是指针类型，所以处理方式有点差异。

接下来补全 handleRequest 方法：

```go
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	err := req.svc.call(req.mtype, req.argv, req.replyv)
	if err != nil {
		req.h.Error = err.Error()
		server.sendResponse(cc, req.h, invalidRequest, sending)
		return
	}
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
```

相对于 readRequest，handleRequest 的实现非常简单，通过 `req.svc.call` 完成方法调用，将 replyv 传递给 sendResponse 完成序列化即可。

到这里，今天的所有内容已经实现完成，成功在服务端实现了服务注册与调用。

## Demo

最后，还是需要写一个可执行程序(main)验证今天的成果。

[day3-service/main/main.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day3-service)

第一步，定义结构体 Foo 和方法 Sum

```go
package main

import (
	"geerpc"
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
```

第二步，注册 Foo 到 Server 中，并启动 RPC 服务

```go
func startServer(addr chan string) {
	var foo Foo
	if err := geerpc.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
}
```

第三步，构造参数，发送 RPC 请求，并打印结果。

```go
func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	client, _ := geerpc.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}
```

运行结果如下：

```bash
rpc server: register Foo.Sum
start rpc server on [::]:57509
1 + 1 = 2
2 + 4 = 6
3 + 9 = 12
0 + 0 = 0
4 + 16 = 20
```

## 附 推荐阅读

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go 语言笔试面试题](https://geektutu.com/post/qa-golang.html)
