---
title: 动手写RPC框架 - GeeRPC第四天 超时处理(timeout)
date: 2020-10-07 23:00:00
description: 7天用 Go语言/golang 从零实现 RPC 框架 GeeRPC 教程(7 days implement golang remote procedure call framework from scratch tutorial)，动手写 RPC 框架，参照 golang 标准库 net/rpc 的实现，实现了服务端(server)、支持异步和并发的客户端(client)、消息编码与解码(message encoding and decoding)、服务注册(service register)、支持 TCP/Unix/HTTP 等多种传输协议。第四天为RPC框架提供了处理超时的能力(timeout processing)。
tags:
- Go
nav: 从零实现
categories:
- RPC框架 - GeeRPC
keywords:
- Go语言
- 从零实现RPC框架
- 连接超时
image: post/geerpc/geerpc.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day4 超时处理
---

![golang RPC framework](geerpc/geerpc.jpg)

本文是[7天用Go从零实现RPC框架GeeRPC](https://geektutu.com/post/geerpc.html)的第四篇。

- 增加连接超时的处理机制
- 增加服务端处理超时的处理机制，代码约 100 行

## 为什么需要超时处理机制

超时处理是 RPC 框架一个比较基本的能力，如果缺少超时处理机制，无论是服务端还是客户端都容易因为网络或其他错误导致挂死，资源耗尽，这些问题的出现大大地降低了服务的可用性。因此，我们需要在 RPC 框架中加入超时处理的能力。

纵观整个远程调用的过程，需要客户端处理超时的地方有：

- 与服务端建立连接，导致的超时
- 发送请求到服务端，写报文导致的超时
- 等待服务端处理时，等待处理导致的超时（比如服务端已挂死，迟迟不响应）
- 从服务端接收响应时，读报文导致的超时

需要服务端处理超时的地方有：

- 读取客户端请求报文时，读报文导致的超时
- 发送响应报文时，写报文导致的超时
- 调用映射服务的方法时，处理报文导致的超时


GeeRPC 在 3 个地方添加了超时处理机制。分别是：

1）客户端创建连接时
2）客户端 `Client.Call()` 整个过程导致的超时（包含发送报文，等待处理，接收报文所有阶段）
3）服务端处理报文，即 `Server.handleRequest` 超时。

## 创建连接超时

为了实现上的简单，将超时设定放在了 Option 中。ConnectTimeout 默认值为 10s，HandleTimeout 默认值为 0，即不设限。

```go
type Option struct {
	MagicNumber    int           // MagicNumber marks this's a geerpc request
	CodecType      codec.Type    // client may choose different Codec to encode body
	ConnectTimeout time.Duration // 0 means no limit
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
}
```

客户端连接超时，只需要为 Dial 添加一层超时处理的外壳即可。

[day4-timeout/client.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day4-timeout)

```go
type clientResult struct {
	client *Client
	err    error
}

type newClientFunc func(conn net.Conn, opt *Option) (client *Client, err error)

func dialTimeout(f newClientFunc, network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()
	ch := make(chan clientResult)
	go func() {
		client, err := f(conn, opt)
		ch <- clientResult{client: client, err: err}
	}()
	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}
}

// Dial connects to an RPC server at the specified network address
func Dial(network, address string, opts ...*Option) (*Client, error) {
	return dialTimeout(NewClient, network, address, opts...)
}
```

在这里实现了一个超时处理的外壳 `dialTimeout`，这个壳将 NewClient 作为入参，在 2 个地方添加了超时处理的机制。

1) 将 `net.Dial` 替换为 `net.DialTimeout`，如果连接创建超时，将返回错误。
2）使用子协程执行 NewClient，执行完成后则通过信道 ch 发送结果，如果 `time.After()` 信道先接收到消息，则说明 NewClient 执行超时，返回错误。

## Client.Call 超时

`Client.Call` 的超时处理机制，使用 context 包实现，控制权交给用户，控制更为灵活。

```go
// Call invokes the named function, waits for it to complete,
// and returns its error status.
func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}
```

用户可以使用 `context.WithTimeout` 创建具备超时检测能力的 context 对象来控制。例如：

```go
ctx, _ := context.WithTimeout(context.Background(), time.Second)
var reply int
err := client.Call(ctx, "Foo.Sum", &Args{1, 2}, &reply)
...
```

## 服务端处理超时

这一部分的实现与客户端很接近，使用 `time.After()` 结合 `select+chan` 完成。

[day4-timeout/server.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day4-timeout)

```go
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}
```

这里需要确保 `sendResponse` 仅调用一次，因此将整个过程拆分为 `called` 和 `sent` 两个阶段，在这段代码中只会发生如下两种情况：

1) called 信道接收到消息，代表处理没有超时，继续执行 sendResponse。
2) `time.After()` 先于 called 接收到消息，说明处理已经超时，called 和 sent 都将被阻塞。在 `case <-time.After(timeout)` 处调用 `sendResponse`。

## 测试用例

第一个测试用例，用于测试连接超时。NewClient 函数耗时 2s，ConnectionTimeout 分别设置为 1s 和 0 两种场景。

[day4-timeout/client_test.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day4-timeout)

```go
func TestClient_dialTimeout(t *testing.T) {
	t.Parallel()
	l, _ := net.Listen("tcp", ":0")

	f := func(conn net.Conn, opt *Option) (client *Client, err error) {
		_ = conn.Close()
		time.Sleep(time.Second * 2)
		return nil, nil
	}
	t.Run("timeout", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: time.Second})
		_assert(err != nil && strings.Contains(err.Error(), "connect timeout"), "expect a timeout error")
	})
	t.Run("0", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: 0})
		_assert(err == nil, "0 means no limit")
	})
}
```

第二个测试用例，用于测试处理超时。`Bar.Timeout` 耗时 2s，场景一：客户端设置超时时间为 1s，服务端无限制；场景二，服务端设置超时时间为1s，客户端无限制。

```go
type Bar int

func (b Bar) Timeout(argv int, reply *int) error {
	time.Sleep(time.Second * 2)
	return nil
}

func startServer(addr chan string) {
	var b Bar
	_ = Register(&b)
	// pick a free port
	l, _ := net.Listen("tcp", ":0")
	addr <- l.Addr().String()
	Accept(l)
}

func TestClient_Call(t *testing.T) {
	t.Parallel()
	addrCh := make(chan string)
	go startServer(addrCh)
	addr := <-addrCh
	time.Sleep(time.Second)
	t.Run("client timeout", func(t *testing.T) {
		client, _ := Dial("tcp", addr)
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		var reply int
		err := client.Call(ctx, "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "expect a timeout error")
	})
	t.Run("server handle timeout", func(t *testing.T) {
		client, _ := Dial("tcp", addr, &Option{
			HandleTimeout: time.Second,
		})
		var reply int
		err := client.Call(context.Background(), "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), "handle timeout"), "expect a timeout error")
	})
}
```

## 附 推荐阅读

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go 语言笔试面试题](https://geektutu.com/post/qa-golang.html)