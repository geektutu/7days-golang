---
title: 动手写RPC框架 - GeeRPC第一天 服务端与消息编码
date: 2020-10-06 17:00:00
description: 7天用 Go语言/golang 从零实现 RPC 框架 GeeRPC 教程(7 days implement golang remote procedure call framework from scratch tutorial)，动手写 RPC 框架，参照 golang 标准库 net/rpc 的实现，实现了服务端(server)、支持异步和并发的客户端(client)、消息编码与解码(message encoding and decoding)、服务注册(service register)、支持 TCP/Unix/HTTP 等多种传输协议。第一天实现了一个简单的服务端和消息的编码与解码。
tags:
- Go
nav: 从零实现
categories:
- RPC框架 - GeeRPC
keywords:
- Go语言
- 从零实现RPC框架
- Codec
- 序列化
- 反序列化
image: post/geerpc/geerpc.jpg
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day1 服务端与消息编码
---

![golang RPC framework](geerpc/geerpc.jpg)

本文是[7天用Go从零实现RPC框架GeeRPC](https://geektutu.com/post/geerpc.html)的第一篇。

- 使用 `encoding/gob` 实现消息的编解码(序列化与反序列化)
- 实现一个简易的服务端，仅接受消息，不处理，代码约 200 行


## 消息的序列化与反序列化

一个典型的 RPC 调用如下：

```go
err = client.Call("Arith.Multiply", args, &reply)
```

客户端发送的请求包括服务名 `Arith`，方法名 `Multiply`，参数 `args` 三个，服务端的响应包括错误 `error`，返回值 `reply` 2 个。我们将请求和响应中的参数和返回值抽象为 body，剩余的信息放在 header 中，那么就可以抽象出数据结构 Header：

[day1-codec/codec/codec.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day1-codec)

```go
package codec

import "io"

type Header struct {
	ServiceMethod string // format "Service.Method"
	Seq           uint64 // sequence number chosen by client
	Error         string
}
```

- ServiceMethod 是服务名和方法名，通常与 Go 语言中的结构体和方法相映射。
- Seq 是请求的序号，也可以认为是某个请求的 ID，用来区分不同的请求。
- Error 是错误信息，客户端置为空，服务端如果如果发生错误，将错误信息置于 Error 中。


我们将和消息编解码相关的代码都防到 codec 子目录中，在此之前，还需要在根目录下使用 `go mod init geerpc` 初始化项目，方便后续子 package 之间的引用。

进一步，抽象出对消息体进行编解码的接口 Codec，抽象出接口是为了实现不同的 Codec 实例：

```go
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}
```

紧接着，抽象出 Codec 的构造函数，客户端和服务端可以通过 Codec 的 `Type` 得到构造函数，从而创建 Codec 实例。这部分代码和工厂模式类似，与工厂模式不同的是，返回的是构造函数，而非实例。

```go
type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
```

我们定义了 2 种 Codec，`Gob` 和 `Json`，但是实际代码中只实现了 `Gob` 一种，事实上，2 者的实现非常接近，甚至只需要把 `gob` 换成 `json` 即可。

首先定义 `GobCodec` 结构体，这个结构体由四部分构成，`conn` 是由构建函数传入，通常是通过 TCP 或者 Unix 建立 socket 时得到的链接实例，dec 和 enc 对应 gob 的 Decoder 和 Encoder，buf 是为了防止阻塞而创建的带缓冲的 `Writer`，一般这么做能提升性能。

[day1-codec/codec/gob.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day1-codec)

```go
package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}
```

接着实现 `ReadHeader`、`ReadBody`、`Write` 和 `Close` 方法。

```go
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}
```

## 通信过程

客户端与服务端的通信需要协商一些内容，例如 HTTP 报文，分为 header 和 body 2 部分，body 的格式和长度通过 header 中的 `Content-Type` 和 `Content-Length` 指定，服务端通过解析 header 就能够知道如何从 body 中读取需要的信息。对于 RPC 协议来说，这部分协商是需要自主设计的。为了提升性能，一般在报文的最开始会规划固定的字节，来协商相关的信息。比如第1个字节用来表示序列化方式，第2个字节表示压缩方式，第3-6字节表示 header 的长度，7-10 字节表示 body 的长度。

对于 GeeRPC 来说，目前需要协商的唯一一项内容是消息的编解码方式。我们将这部分信息，放到结构体 `Option` 中承载。目前，已经进入到服务端的实现阶段了。

[day1-codec/server.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day1-codec)

```go
package geerpc

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int        // MagicNumber marks this's a geerpc request
	CodecType   codec.Type // client may choose different Codec to encode body
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}
```

一般来说，涉及协议协商的这部分信息，需要设计固定的字节来传输的。但是为了实现上更简单，GeeRPC 客户端固定采用 JSON 编码 Option，后续的 header 和 body 的编码方式由 Option 中的 CodeType 指定，服务端首先使用 JSON 解码 Option，然后通过 Option 得 CodeType 解码剩余的内容。即报文将以这样的形式发送：

```bash
| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|
```

在一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多个，即报文可能是这样的。

```go
| Option | Header1 | Body1 | Header2 | Body2 | ...
```

## 服务端的实现

通信过程已经定义清楚了，那么服务端的实现就比较直接了。

[day1-codec/server.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day1-codec)

```go
// Server represents an RPC Server.
type Server struct{}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
```

- 首先定义了结构体 `Server`，没有任何的成员字段。
- 实现了 `Accept` 方式，`net.Listener` 作为参数，for 循环等待 socket 连接建立，并开启子协程处理，处理过程交给了 `ServerConn` 方法。
- DefaultServer 是一个默认的 `Server` 实例，主要为了用户使用方便。

如果想启动服务，过程是非常简单的，传入 listener 即可，tcp 协议和 unix 协议都支持。

```go
lis, _ := net.Listen("tcp", ":9999")
geerpc.Accept(lis)
```

`ServeConn` 的实现就和之前讨论的通信过程紧密相关了，首先使用 `json.NewDecoder` 反序列化得到 Option 实例，检查 MagicNumber 和 CodeType 的值是否正确。然后根据 CodeType 得到对应的消息编解码器，接下来的处理交给 `serverCodec`。

```go
// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCodec(f(conn))
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)  // wait until all request are handled
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break // it's not possible to recover, so close the connection
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}
```

`serveCodec` 的过程非常简单。主要包含三个阶段

- 读取请求 readRequest
- 处理请求 handleRequest
- 回复请求 sendResponse

之前提到过，在一次连接中，允许接收多个请求，即多个 request header 和 request body，因此这里使用了 for 无限制地等待请求的到来，直到发生错误（例如连接被关闭，接收到的报文有问题等），这里需要注意的点有三个：

- handleRequest 使用了协程并发执行请求。
- 处理请求是并发的，但是回复请求的报文必须是逐个发送的，并发容易导致多个回复报文交织在一起，客户端无法解析。在这里使用锁(sending)保证。
- 尽力而为，只有在 header 解析失败时，才终止循环。

```go
// request stores all information of a call
type request struct {
	h            *codec.Header // header of request
	argv, replyv reflect.Value // argv and replyv of request
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	// TODO: now we don't know the type of request argv
	// day 1, just suppose it's string
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	// TODO, should call registered rpc methods to get the right replyv
	// day 1, just print argv and send a hello message
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
```

目前还不能判断 body 的类型，因此在 readRequest 和 handleRequest 中，day1 将 body 作为字符串处理。接收到请求，打印 header，并回复 `geerpc resp ${req.h.Seq}`。这一部分后续再实现。


## main 函数(一个简易的客户端)

day1 的内容就到此为止了，在这里我们已经实现了一个消息的编解码器 `GobCodec`，并且客户端与服务端实现了简单的协议交换(protocol exchange)，即允许客户端使用不同的编码方式。同时实现了服务端的雏形，建立连接，读取、处理并回复客户端的请求。

接下来，我们就在 main 函数中看看如何使用刚实现的 GeeRPC 吧。

[day1-codec/main/main.go](https://github.com/geektutu/7days-golang/tree/master/gee-rpc/day1-codec)

```go
package main

import (
	"encoding/json"
	"fmt"
	"geerpc"
	"geerpc/codec"
	"log"
	"net"
	"time"
)

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

func main() {
	addr := make(chan string)
	go startServer(addr)

	// in fact, following code is like a simple geerpc client
	conn, _ := net.Dial("tcp", <-addr)
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)
	// send options
	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
	cc := codec.NewGobCodec(conn)
	// send request & receive response
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
```

- 在 `startServer` 中使用了信道 `addr`，确保服务端端口监听成功，客户端再发起请求。
- 客户端首先发送 `Option` 进行协议交换，接下来发送消息头 `h := &codec.Header{}`，和消息体 `geerpc req ${h.Seq}`。
- 最后解析服务端的响应 `reply`，并打印出来。

执行结果如下：

```bash
start rpc server on [::]:63662
&{Foo.Sum 0 } geerpc req 0
reply: geerpc resp 0
&{Foo.Sum 1 } geerpc req 1
reply: geerpc resp 1
&{Foo.Sum 2 } geerpc req 2
reply: geerpc resp 2
&{Foo.Sum 3 } geerpc req 3
reply: geerpc resp 3
&{Foo.Sum 4 } geerpc req 4
reply: geerpc resp 4
```

## 附 推荐阅读

- [Go 语言简明教程](https://geektutu.com/post/quick-golang.html)
- [Go 语言笔试面试题](https://geektutu.com/post/qa-golang.html)