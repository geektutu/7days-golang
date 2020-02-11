# 7 days golang apps from scratch

<details>
<summary><strong>README 中文版本</strong></summary>
<div>

## 7天用Go从零实现系列

7天能写什么呢？类似 gin 的 web 框架？类似 groupcache 的分布式缓存？或者一个简单的 Python 解释器？希望这个仓库能给你答案。

推荐先阅读 **[Go 语言简明教程](https://geektutu.com/post/quick-golang.html)**，一篇文章了解Go的基本语法、并发编程，依赖管理等内容

### 7天用Go从零实现Web框架 - Gee

[Gee](https://geektutu.com/post/gee.html) 是一个模仿 [gin](https://github.com/gin-gonic/gin) 实现的 Web 框架，[Go Gin简明教程](https://geektutu.com/post/quick-go-gin.html)可以快速入门。

- 第一天：[前置知识(http.Handler接口)](https://geektutu.com/post/gee-day1.html) | [Code](gee-web/day1-http-base)
- 第二天：[上下文设计(Context)](https://geektutu.com/post/gee-day2.html) | [Code](gee-web/day2-context)
- 第三天：[Tire树路由(Router)](https://geektutu.com/post/gee-day3.html) | [Code](gee-web/day3-router)
- 第四天：[分组控制(Group)](https://geektutu.com/post/gee-day4.html) | [Code](gee-web/day4-group)
- 第五天：[中间件(Middleware)](https://geektutu.com/post/gee-day5.html) | [Code](gee-web/day5-middleware)
- 第六天：[HTML模板(Template)](https://geektutu.com/post/gee-day6.html) | [Code](gee-web/day6-template)
- 第七天：[错误恢复(Panic Recover)](https://geektutu.com/post/gee-day7.html) | [Code](gee-web/day7-panic-recover)

### 7天用Go从零实现分布式缓存 GeeCache

[GeeCache](https://geektutu.com/post/geecache.html) 是一个模仿 [groupcache](https://github.com/golang/groupcache) 实现的分布式缓存系统

- [第一天：LRU 缓存淘汰策略](https://geektutu.com/post/geecache-day1.html) | [Code](gee-cache/day1-lru)
- 第二天：单机并发缓存 | [Code](gee-cache/day2-single-node)
- 第三天：HTTP 服务端 | [Code](gee-cache/day3-http-server)
- 第四天：一致性哈希(Hash) | [Code](gee-cache/day4-consistent-hash)
- 第五天：分布式节点 | [Code](gee-cache/day5-multi-nodes)
- 第六天：防止缓存击穿 | [Code](gee-cache/day6-single-flight)
- 第七天：使用 Protobuf 通信 | [Code](gee-cache/day7-proto-buf)

### WebAssembly 使用示例

具体的实践过程记录在 [Go WebAssembly 简明教程](https://geektutu.com/post/quick-go-wasm.html)。

- 示例一：Hello World | [Code](demo-wasm/hello-world)
- 示例二：注册函数 | [Code](demo-wasm/register-functions)
- 示例三：操作 DOM | [Code](demo-wasm/manipulate-dom)
- 示例四：回调函数 | [Code](demo-wasm/callback)

</div>
</details>

What can I write in 7 days? A gin-like web framework? A distributed cache like groupcache? Or a simple Python interpreter? Hope this repo can give you the answer.

## Web Framework - Gee

[Gee](https://geektutu.com/post/gee.html) is a [gin](https://github.com/gin-gonic/gin)-like framework

- Day 1 - http.Handler Interface Basic [Code](gee-web/day1-http-base)
- Day 2 - Design a Flexiable Context [Code](gee-web/day2-context)
- Day 3 - Router with Tire-Tree Algorithm [Code](gee-web/day3-router)
- Day 4 - Group Control [Code](gee-web/day4-group)
- Day 5 - Middleware Mechanism [Code](gee-web/day5-middleware)
- Day 6 - Embeded Template Support [Code](gee-web/day6-template)
- Day 7 - Panic Recover & Make it Robust [Code](gee-web/day7-panic-recover)

## Distributed Cache - Geecache

[GeeCache](https://geektutu.com/post/geecache.html) is a [groupcache](https://github.com/golang/groupcache)-like distributed cache

- Day 1 - LRU (Least Recently Used) Caching Strategy [Code](gee-cache/day1-lru)
- Day 2 - Single Machine Concurrent Cache [Code](gee-cache/day2-single-node)
- Day 3 - Launch a HTTP Server [Code](gee-cache/day3-http-server)
- Day 4 - Consistent Hash Algorithm [Code](gee-cache/day4-consistent-hash)
- Day 5 - Communication between Distributed Nodes [Code](gee-cache/day5-multi-nodes)
- Day 6 - Cache Breakdown & Single Flight  | [Code](gee-cache/day6-single-flight)
- Day 7 - Use Protobuf as RPC Data Exchange Type | [Code](gee-cache/day7-proto-buf)

## Golang WebAssembly Demo

- Demo 1 - Hello World [Code](demo-wasm/hello-world)
- Demo 2 - Register Functions [Code](demo-wasm/register-functions)
- Demo 3 - Manipulate DOM [Code](demo-wasm/manipulate-dom)
- Demo 4 - Callback [Code](demo-wasm/callback)