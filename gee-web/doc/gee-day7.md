---
title: Go语言动手写Web框架 - Gee第七天 错误恢复(Panic Recover)
date: 2020-01-09 01:00:00
description: 7天用 Go语言 从零实现Web框架教程(7 days implement golang web framework from scratch tutorial)，用 Go语言/golang 动手写Web框架，从零实现一个Web框架，以 Gin 为原型从零设计一个Web框架。本文介绍了如何为Web框架增加错误处理机制。
tags:
- Go
nav: 从零实现
categories:
- Web框架 - Gee
keywords:
- Go语言
- 从零实现Web框架
- 动手写Web框架
- Panic
- Recover
image: post/gee-day7/go-panic.png
github: https://github.com/geektutu/7days-golang
book: 七天用Go从零实现系列
book_title: Day7 错误恢复
---

本文是[7天用Go从零实现Web框架Gee教程系列](https://geektutu.com/post/gee.html)的第七篇。

- 实现错误处理机制。

## panic

Go 语言中，比较常见的错误处理方法是返回 error，由调用者决定后续如何处理。但是如果是无法恢复的错误，可以手动触发 panic，当然如果在程序运行过程中出现了类似于数组越界的错误，panic 也会被触发。panic 会中止当前执行的程序，退出。

下面是主动触发的例子：

```go
// hello.go
func main() {
	fmt.Println("before panic")
	panic("crash")
	fmt.Println("after panic")
}
```

```bash
$ go run hello.go

before panic
panic: crash

goroutine 1 [running]:
main.main()
        ~/go_demo/hello/hello.go:7 +0x95
exit status 2
```

下面是数组越界触发的 panic

```go
// hello.go
func main() {
	arr := []int{1, 2, 3}
	fmt.Println(arr[4])
}
```

```bash
$ go run hello.go
panic: runtime error: index out of range [4] with length 3
```

## defer

panic 会导致程序被中止，但是在退出前，会先处理完当前协程上已经defer 的任务，执行完成后再退出。效果类似于 java 语言的 `try...catch`。

```go
// hello.go
func main() {
	defer func() {
		fmt.Println("defer func")
	}()

	arr := []int{1, 2, 3}
	fmt.Println(arr[4])
}
```

```go
$ go run hello.go 
defer func
panic: runtime error: index out of range [4] with length 3
```

可以 defer 多个任务，在同一个函数中 defer 多个任务，会逆序执行。即先执行最后 defer 的任务。

在这里，defer 的任务执行完成之后，panic 还会继续被抛出，导致程序非正常结束。

## recover

Go 语言还提供了 recover 函数，可以避免因为 panic 发生而导致整个程序终止，recover 函数只在 defer 中生效。

```go
// hello.go
func test_recover() {
	defer func() {
		fmt.Println("defer func")
		if err := recover(); err != nil {
			fmt.Println("recover success")
		}
	}()

	arr := []int{1, 2, 3}
	fmt.Println(arr[4])
	fmt.Println("after panic")
}

func main() {
	test_recover()
	fmt.Println("after recover")
}
```

```go
$ go run hello.go 
defer func
recover success
after recover
```

我们可以看到，recover 捕获了 panic，程序正常结束。*test_recover()* 中的 *after panic* 没有打印，这是正确的，当 panic 被触发时，控制权就被交给了 defer 。就像在 java 中，`try`代码块中发生了异常，控制权交给了 `catch`，接下来执行 catch 代码块中的代码。而在 *main()* 中打印了 *after recover*，说明程序已经恢复正常，继续往下执行直到结束。

## Gee 的错误处理机制

对一个 Web 框架而言，错误处理机制是非常必要的。可能是框架本身没有完备的测试，导致在某些情况下出现空指针异常等情况。也有可能用户不正确的参数，触发了某些异常，例如数组越界，空指针等。如果因为这些原因导致系统宕机，必然是不可接受的。

我们在[第六天](https://geektutu.com/post/gee-day6.html)实现的框架并没有加入异常处理机制，如果代码中存在会触发 panic 的 BUG，很容易宕掉。

例如下面的代码：

```go
func main() {
	r := gee.New()
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})
	r.Run(":9999")
}
```
在上面的代码中，我们为 gee 注册了路由 `/panic`，而这个路由的处理函数内部存在数组越界 `names[100]`，如果访问 *localhost:9999/panic*，Web 服务就会宕掉。

今天，我们将在 gee 中添加一个非常简单的错误处理机制，即在此类错误发生时，向用户返回 *Internal Server Error*，并且在日志中打印必要的错误信息，方便进行错误定位。

我们之前实现了中间件机制，错误处理也可以作为一个中间件，增强 gee 框架的能力。

新增文件 **gee/recovery.go**，在这个文件中实现中间件 `Recovery`。

```go
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}
```

`Recovery` 的实现非常简单，使用 defer 挂载上错误恢复的函数，在这个函数中调用 *recover()*，捕获 panic，并且将堆栈信息打印在日志中，向用户返回 *Internal Server Error*。

你可能注意到，这里有一个 *trace()* 函数，这个函数是用来获取触发 panic 的堆栈信息，完整代码如下：

[day7-panic-recover/gee/recovery.go](https://github.com/geektutu/7days-golang/tree/master/gee-web/day7-panic-recover)

```go
package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// print stack trace for debug
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}
```

在 *trace()* 中，调用了 `runtime.Callers(3, pcs[:])`，Callers 用来返回调用栈的程序计数器, 第 0 个 Caller 是 Callers 本身，第 1 个是上一层 trace，第 2 个是再上一层的 `defer func`。因此，为了日志简洁一点，我们跳过了前 3 个 Caller。

接下来，通过 `runtime.FuncForPC(pc)` 获取对应的函数，在通过 `fn.FileLine(pc)` 获取到调用该函数的文件名和行号，打印在日志中。

至此，gee 框架的错误处理机制就完成了。

## 使用 Demo

[day7-panic-recover/main.go](https://github.com/geektutu/7days-golang/tree/master/gee-web/day7-panic-recover)

```go
package main

import (
	"net/http"

	"gee"
)

func main() {
	r := gee.Default()
	r.GET("/", func(c *gee.Context) {
		c.String(http.StatusOK, "Hello Geektutu\n")
	})
	// index out of range for testing Recovery()
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}
```

接下来进行测试，先访问主页，访问一个有BUG的 `/panic`，服务正常返回。接下来我们再一次成功访问了主页，说明服务完全运转正常。

```bash
$ curl "http://localhost:9999"
Hello Geektutu
$ curl "http://localhost:9999/panic"
{"message":"Internal Server Error"}
$ curl "http://localhost:9999"
Hello Geektutu
```

我们可以在后台日志中看到如下内容，引发错误的原因和堆栈信息都被打印了出来，通过日志，我们可以很容易地知道，在*day7-panic-recover/main.go:47* 的地方出现了 `index out of range` 错误。

```bash
2020/01/09 01:00:10 Route  GET - /
2020/01/09 01:00:10 Route  GET - /panic
2020/01/09 01:00:22 [200] / in 25.364µs
2020/01/09 01:00:32 runtime error: index out of range
Traceback:
        /usr/local/Cellar/go/1.12.5/libexec/src/runtime/panic.go:523
        /usr/local/Cellar/go/1.12.5/libexec/src/runtime/panic.go:44
        /tmp/7days-golang/day7-panic-recover/main.go:47
        /tmp/7days-golang/day7-panic-recover/gee/context.go:41
        /tmp/7days-golang/day7-panic-recover/gee/recovery.go:37
        /tmp/7days-golang/day7-panic-recover/gee/context.go:41
        /tmp/7days-golang/day7-panic-recover/gee/logger.go:15
        /tmp/7days-golang/day7-panic-recover/gee/context.go:41
        /tmp/7days-golang/day7-panic-recover/gee/router.go:99
        /tmp/7days-golang/day7-panic-recover/gee/gee.go:130
        /usr/local/Cellar/go/1.12.5/libexec/src/net/http/server.go:2775
        /usr/local/Cellar/go/1.12.5/libexec/src/net/http/server.go:1879
        /usr/local/Cellar/go/1.12.5/libexec/src/runtime/asm_amd64.s:1338

2020/01/09 01:00:32 [500] /panic in 395.846µs
2020/01/09 01:00:38 [200] / in 6.985µs
```

## 参考

- [Package runtime - golang.org](https://golang.org/pkg/runtime/)
- [Is it possible get information about caller function in Golang? - StackOverflow](https://stackoverflow.com/questions/35212985/is-it-possible-get-information-about-caller-function-in-golang)

