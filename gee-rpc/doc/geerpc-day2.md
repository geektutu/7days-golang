---
title: 动手写RPC框架 - GeeRPC第二天 支持并发与异步的客户端
date: 2020-10-07 18:00:00
description: 7天用 Go语言/golang 从零实现 RPC 框架 GeeRPC 教程(7 days implement golang remote procedure call framework from scratch tutorial)，动手写 RPC 框架，参照 golang 标准库 net/rpc 的实现，实现了服务端(server)、支持异步和并发的客户端(client)、消息编码与解码(message encoding and decoding)、服务注册(service register)、支持 TCP/Unix/HTTP 等多种传输协议。第二天实现了一个支持异步(asynchronous)和并发(concurrent)的客户端。
tags:
- Go
nav: 从零实现
categories:
- RPC框架 - GeeRPC
keywords:
- Go语言
- 从零实现RPC框架
- 客户端
- 异步
- 并发
image: post/geerpc/geerpc.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day2 高性能客户端
---

![golang RPC framework](geerpc/geerpc.jpg)

本文是[7天用Go从零实现RPC框架GeeRPC](https://geektutu.com/post/geerpc.html)的第二篇。

- 实现一个支持异步和并发的高性能客户端，代码约 250 行


## Call 的设计

对 `net/rpc` 而言，一个函数需要能够被远程调用，需要满足如下五个条件：

- the method's type is exported.
- the method is exported.
- the method has two arguments, both exported (or builtin) types.
- the method's second argument is a pointer.
- the method has return type error.

更直观一些：

```go
func (t *T) MethodName(argType T1, replyType *T2) error
```

根据上述要求，首先我们封装了结构体 Call 来承载一次 RPC 调用所需要的信息。

[day2-client/client.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day2-client)

```go
// Call represents an active RPC.
type Call struct {
	Seq           uint64
	ServiceMethod string      // format "<service>.<method>"
	Args          interface{} // arguments to the function
	Reply         interface{} // reply from the function
	Error         error       // if error occurs, it will be set
	Done          chan *Call  // Strobes when call is complete.
}

func (call *Call) done() {
	call.Done <- call
}
```

为了支持异步调用，Call 结构体中添加了一个字段 Done，Done 的类型是 `chan *Call`，当调用结束时，会调用 `call.done()` 通知调用方。


## 实现 Client

接下来，我们将实现 GeeRPC 客户端最核心的部分 Client。

```go
// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	cc       codec.Codec
	opt      *Option
	sending  sync.Mutex // protect following
	header   codec.Header
	mu       sync.Mutex // protect following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

// Close the connection
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

// IsAvailable return true if the client does work
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}
```

Client 的字段比较复杂：

- cc 是消息的编解码器，和服务端类似，用来序列化将要发送出去的请求，以及反序列化接收到的响应。
- sending 是一个互斥锁，和服务端类似，为了保证请求的有序发送，即防止出现多个请求报文混淆。
- header 是每个请求的消息头，header 只有在请求发送时才需要，而请求发送是互斥的，因此每个客户端只需要一个，声明在 Client 结构体中可以复用。
- seq 用于给发送的请求编号，每个请求拥有唯一编号。
- pending 存储未处理完的请求，键是编号，值是 Call 实例。
- closing 和 shutdown 任意一个值置为 true，则表示 Client 处于不可用的状态，但有些许的差别，closing 是用户主动关闭的，即调用 `Close` 方法，而 shutdown 置为 true 一般是有错误发生。

紧接着，实现和 Call 相关的三个方法。

```go
func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing || client.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}
```

- registerCall：将参数 call 添加到 client.pending 中，并更新 client.seq。
- removeCall：根据 seq，从 client.pending 中移除对应的 call，并返回。
- terminateCalls：服务端或客户端发生错误时调用，将 shutdown 设置为 true，且将错误信息通知所有 pending 状态的 call。

对一个客户端端来说，接收响应、发送请求是最重要的 2 个功能。那么首先实现接收功能，接收到的响应有三种情况：

- call 不存在，可能是请求没有发送完整，或者因为其他原因被取消，但是服务端仍旧处理了。
- call 存在，但服务端处理出错，即 h.Error 不为空。
- call 存在，服务端处理正常，那么需要从 body 中读取 Reply 的值。

```go
func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			// it usually means that Write partially failed
			// and call was already removed.
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	// error occurs, so terminateCalls pending calls
	client.terminateCalls(err)
}
```

创建 Client 实例时，首先需要完成一开始的协议交换，即发送 `Option` 信息给服务端。协商好消息的编解码方式之后，再创建一个子协程调用 `receive()` 接收响应。

```go
func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}
	// send options with server
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // seq starts with 1, 0 means invalid call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}
```

还需要实现 `Dial` 函数，便于用户传入服务端地址，创建 Client 实例。为了简化用户调用，通过 `...*Option` 将 Option 实现为可选参数。

```go
func parseOptions(opts ...*Option) (*Option, error) {
	// if opts is nil or pass nil as parameter
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

// Dial connects to an RPC server at the specified network address
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}
```

此时，GeeRPC 客户端已经具备了完整的创建连接和接收响应的能力了，最后还需要实现发送请求的能力。

```go
func (client *Client) send(call *Call) {
	// make sure that the client will send a complete request
	client.sending.Lock()
	defer client.sending.Unlock()

	// register this call.
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// prepare request header
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	// encode and send the request
	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		// call may be nil, it usually means that Write partially failed,
		// client has received the response and handled
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go invokes the function asynchronously.
// It returns the Call structure representing the invocation.
func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	client.send(call)
	return call
}

// Call invokes the named function, waits for it to complete,
// and returns its error status.
func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}
```

- `Go` 和 `Call` 是客户端暴露给用户的两个 RPC 服务调用接口，`Go` 是一个异步接口，返回 call 实例。
- `Call` 是对 `Go` 的封装，阻塞 call.Done，等待响应返回，是一个同步接口。

至此，一个支持异步和并发的 GeeRPC 客户端已经完成。

## Demo

第一天 GeeRPC 只实现了服务端，因此我们在 main 函数中手动模拟了整个通信过程，今天我们就将 main 函数中通信部分替换为今天的客户端吧。

[day2-client/main/main.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day2-client)

startServer 没有发生变化。

```go
func startServer(addr chan string) {
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

在 main 函数中使用了 `client.Call` 并发了 5 个 RPC 同步调用，参数和返回值的类型均为 string。

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
			args := fmt.Sprintf("geerpc req %d", i)
			var reply string
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}
```

运行结果如下：

```bash
start rpc server on [::]:50658
&{Foo.Sum 5 } geerpc req 3
&{Foo.Sum 1 } geerpc req 0
&{Foo.Sum 3 } geerpc req 1
&{Foo.Sum 2 } geerpc req 4
&{Foo.Sum 4 } geerpc req 2
reply: geerpc resp 1
reply: geerpc resp 5
reply: geerpc resp 3
reply: geerpc resp 2
reply: geerpc resp 4
```

## 附 推荐阅读

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go 语言笔试面试题](https://geektutu.com/post/qa-golang.html)