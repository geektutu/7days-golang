# 7天用Go从零实现系列

7天能写什么呢？类似 gin 的 web 框架？类似 groupcache 的分布式缓存？或者一个简单的 Python 解释器？希望这个仓库能给你答案。

推荐先阅读 **[Go 语言简明教程](https://geektutu.com/post/quick-golang.html)**，一篇文章了解Go的基本语法、并发编程，依赖管理等内容

## 7天用Go从零实现Web框架Gee

Gee 的设计与实现参考了Gin，[Go Gin简明教程](https://geektutu.com/post/quick-go-gin.html)可以快速入门。

#### [教程目录](https://geektutu.com/post/gee.html)

- [第一天：前置知识(http.Handler接口)](https://geektutu.com/post/gee-day1.html)，[Code - Github](gee-web/day1-http-base)
- [第二天：上下文设计(Context)](https://geektutu.com/post/gee-day2.html)，[Code - Github](gee-web/day2-context)
- [第三天：Tire树路由(Router)](https://geektutu.com/post/gee-day3.html)，[Code - Github](gee-web/day3-router)
- [第四天：分组控制(Group)](https://geektutu.com/post/gee-day4.html)，[Code - Github](gee-web/day4-group)
- [第五天：中间件(Middleware)](https://geektutu.com/post/gee-day5.html)，[Code - Github](gee-web/day5-middleware)
- [第六天：HTML模板(Template)](https://geektutu.com/post/gee-day6.html)，[Code - Github](gee-web/day6-template)
- [第七天：错误恢复(Panic Recover)](https://geektutu.com/post/gee-day7.html)，[Code - Github](gee-web/day7-panic-recover)

## WebAssembly Demo

#### [教程地址](https://geektutu.com/post/quick-go-wasm.html)

- [1. Hello World](demo-wasm/hello-world)
- [2. 注册函数](demo-wasm/hello-world)
- [3. 操作 DOM](demo-wasm/hello-world)
- [4. 回调函数](demo-wasm/hello-world)