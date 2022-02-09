解决问题：

`http.HandlerFunc和http.Handle的区别？`

``为什么只要传入任何实现了 ServerHTTP接口的实例，所有的HTTP请            求，就都会交给了该实例处理？`

`go语言中原生的一个web服务器启动的流程？`



## 使用go语言开启一个最基本的web服务

下面是使用go语言内置的net包开启的一个web服务： 

```go

func main() {
   
   //注册一个服务
   http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
      fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
   })

   // 监听8080端口
   log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 分析web服务的流程

1. 当我们开启上面的服务的时候`go run main.go`，此时在ListenAndServe方法会先创建一个`Server`结构。


```go
// ListenAndServe always returns a non-nil error.
func ListenAndServe(addr string, handler Handler) error {
   server := &Server{Addr: addr, Handler: handler}
   return server.ListenAndServe()
}
```

  2. 开启`net.Listen`进行监听，调用 `srv.Serve(ln)`

```go
// ListenAndServe always returns a non-nil error. After Shutdown or Close,
// the returned error is ErrServerClosed.
func (srv *Server) ListenAndServe() error {
   if srv.shuttingDown() {
      return ErrServerClosed
   }
   addr := srv.Addr
   if addr == "" {
      addr = ":http"
   }
   ln, err := net.Listen("tcp", addr)
   if err != nil {
      return err
   }
   return srv.Serve(ln)
}
```

3.`Serve` 函数中，用了一个 for 循环，通过 `l.Accept`不断接收从客户端传进来的请求连接。当接收到了一个新的请求连接的时候，通过 srv.NewConn创建了一个连接结构（`http.conn`），并创建一个 Goroutine 为这个请求连接对应服务（`c.serve`）。也就是说只要服务端监听到了服务，就会开启一个goroutine。
(这里只贴上了主要的代码，因为这个函数的逻辑较多)

```go
for {
   rw, err := l.Accept()
   if err != nil {
      select {
      case <-srv.getDoneChan():
         return ErrServerClosed
      default:
      }
      if ne, ok := err.(net.Error); ok && ne.Temporary() {
         if tempDelay == 0 {
            tempDelay = 5 * time.Millisecond
         } else {
            tempDelay *= 2
         }
         if max := 1 * time.Second; tempDelay > max {
            tempDelay = max
         }
         srv.logf("http: Accept error: %v; retrying in %v", err, tempDelay)
         time.Sleep(tempDelay)
         continue
      }
      return err
   }
   connCtx := ctx
   if cc := srv.ConnContext; cc != nil {
      connCtx = cc(connCtx, rw)
      if connCtx == nil {
         panic("ConnContext returned nil")
      }
   }
   tempDelay = 0
   c := srv.newConn(rw)
   c.setState(c.rwc, StateNew, runHooks) // before Serve can return
   go c.serve(connCtx)
```
4.`c.serve`代码量很大，但是只要知道它的功能是判断本次 HTTP 请求是否需要升级为 HTTPs，接着创建读文本的 reader 和写文本的 buffer，再进一步读取本次请求数据。
（`下面代码只是很少的一部分，因为serve的函数量太大，最最重要的就是serverHandler{c.server}.ServeHTTP(w, w.req)，大家可以在此函数的大概1930行左右看到下面的代码，go1.17`）

```go
serverHandler{c.server}.ServeHTTP(w, w.req)
w.cancelCtx()
if c.hijacked() {
   return
}
w.finishRequest()
if !w.shouldReuseConnection() {
   if w.requestBodyLimitHit || w.closedRequestBodyEarly() {
      c.closeWriteAndWait()
   }
   return
}
c.setState(c.rwc, StateIdle, runHooks)
c.curReq.Store((*response)(nil))
```

5.`serverHandler{c.server}.ServeHTTP(w, w.req)`这个是最重要的函数，也就是说如果你在服务开始的时候自定义了handler，那么就使用你自定义的，如果没有，就使用go默认的。`也就是说，只要传入任何实现了 ServerHTTP接口的实例，所有的HTTP请求，就都交给了该实例处理了。`

```go

type Handler interface {
   ServeHTTP(ResponseWriter, *Request)
}

type serverHandler struct {
   srv *Server
}

func (sh serverHandler) ServeHTTP(rw ResponseWriter, req *Request) {
   handler := sh.srv.Handler
   if handler == nil {
      handler = DefaultServeMux
   }
   if req.RequestURI == "*" && req.Method == "OPTIONS" {
      handler = globalOptionsHandler{}
   }

   if req.URL != nil && strings.Contains(req.URL.RawQuery, ";") {
      var allowQuerySemicolonsInUse int32
      req = req.WithContext(context.WithValue(req.Context(), silenceSemWarnContextKey, func() {
         atomic.StoreInt32(&allowQuerySemicolonsInUse, 1)
      }))
      defer func() {
         if atomic.LoadInt32(&allowQuerySemicolonsInUse) == 0 {
            sh.srv.logf("http: URL query contains semicolon, which is no longer a supported separator; parts of the query may be stripped when parsed; see golang.org/issue/25192")
         }
      }()
   }

   handler.ServeHTTP(rw, req)
}
```

6.经过上面的分析我们也就知道了go语言web服务的大致流程，也就是说如果我们想要修改web服务，或者说定制web，那么我们只需要自定义handler就可以完成了。

go默认的DefaultServeMux只是简单的使用map来存放路由，key是路径，比如/hello，value是具体的处理逻辑。这也是一般开发不使用原生web的原因。


## 下面给一个拓展，为什么gin这么快呢？

gin框架使用的是定制版本的`httprouter`。我们来分析一下`httprouter`
下面是`htttprouter`的`demo`



```go

package main

import (
   "fmt"
   "log"
   "net/http"

   "github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
   fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
   fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
   router := httprouter.New()
   router.GET("/", Index)
   router.GET("/hello/:name", Hello)

   log.Fatal(http.ListenAndServe(":8080", router))
}
```

这是我在`httprouter`里面找到的代码,也就是httprouter的路由匹配的逻辑：

```go
// ServeHTTP makes the router implement the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
   if r.PanicHandler != nil {
      defer r.recv(w, req)
   }

   path := req.URL.Path

   if root := r.trees[req.Method]; root != nil {
      if handle, ps, tsr := root.getValue(path); handle != nil {
         handle(w, req, ps)
         return
      } else if req.Method != http.MethodConnect && path != "/" {
         code := 301 // Permanent redirect, request with GET method
         if req.Method != http.MethodGet {
            // Temporary redirect, request with same method
            // As of Go 1.3, Go does not support status code 308.
            code = 307
         }

         if tsr && r.RedirectTrailingSlash {
            if len(path) > 1 && path[len(path)-1] == '/' {
               req.URL.Path = path[:len(path)-1]
            } else {
               req.URL.Path = path + "/"
            }
            http.Redirect(w, req, req.URL.String(), code)
            return
         }

         // Try to fix the request path
         if r.RedirectFixedPath {
            fixedPath, found := root.findCaseInsensitivePath(
               CleanPath(path),
               r.RedirectTrailingSlash,
            )
            if found {
               req.URL.Path = string(fixedPath)
               http.Redirect(w, req, req.URL.String(), code)
               return
            }
         }
      }
   }

   if req.Method == http.MethodOptions && r.HandleOPTIONS {
      // Handle OPTIONS requests
      if allow := r.allowed(path, http.MethodOptions); allow != "" {
         w.Header().Set("Allow", allow)
         if r.GlobalOPTIONS != nil {
            r.GlobalOPTIONS.ServeHTTP(w, req)
         }
         return
      }
   } else if r.HandleMethodNotAllowed { // Handle 405
      if allow := r.allowed(path, req.Method); allow != "" {
         w.Header().Set("Allow", allow)
         if r.MethodNotAllowed != nil {
            r.MethodNotAllowed.ServeHTTP(w, req)
         } else {
            http.Error(w,
               http.StatusText(http.StatusMethodNotAllowed),
               http.StatusMethodNotAllowed,
            )
         }
         return
      }
   }

   // Handle 404
   if r.NotFound != nil {
      r.NotFound.ServeHTTP(w, req)
   } else {
      http.NotFound(w, req)
   }
}
```

了解了这些，我们就可以知道为什么gin这么快了，其路由的原理是大量使用公共前缀的树结构，它基本上是一个紧凑的[Trie tree](https://baike.sogou.com/v66237892.htm)。具有公共前缀的节点也共享一个公共父节点。

**所以我们可以根据不同的逻辑来改变web服务的匹配的逻辑，来实现我们的定制web服务**

## 最后来一个demo

重要的事再强调一遍，`也就是说，只要传入任何实现了 ServerHTTP接口的实例，所有的HTTP请求，就都交给了该实例处理了。`（这句话是从兔兔大佬那边抄过来的，也是最重要的一个结论）



```go
package main

import (
   "fmt"
   "log"
   "net/http"
)

// Engine is the uni handler for all requests
type Engine struct{}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
   switch req.URL.Path {
   case "/":
      fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
   case "/hello":
      for k, v := range req.Header {
         fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
      }
   default:
      fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
   }
}

func main() {
   engine := new(Engine)
   log.Fatal(http.ListenAndServe(":9999", engine))
}
```

在理解完这篇文章后可以去到兔兔大佬的博客进行web框架的实战，相信大家肯定更容易理解。


附 ： 部分代码摘自 go源码，极客兔兔，叶剑峰博客，httprouter源码