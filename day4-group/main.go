package main

// (1) v1
// $ curl -i http://localhost:9999/v1/
// HTTP/1.1 200 OK
// Date: Mon, 12 Aug 2019 18:11:07 GMT
// Content-Length: 18
// Content-Type: text/html; charset=utf-8
// <h1>Hello Gee</h1>

// (2)
// $ curl "http://localhost:9999/v1/hello?name=geektutu"
// hello geektutu, you're at /v1/hello

// (3)
// $ curl "http://localhost:9999/v2/hello/geektutu"
// hello , you're at /v2/hello/geektutu

// (4)
// $ curl "http://localhost:9999/v2/login" -X POST -d 'username=geektutu&password=1234'
// {"password":"1234","username":"geektutu"}

// (5)
// $ curl "http://localhost:9999/hello"
// 404 NOT FOUND: /hello

import (
	"net/http"

	"./gee"
)

func main() {
	r := gee.New()
	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *gee.Context) {
			c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})

		v1.GET("/hello", func(c *gee.Context) {
			// expect /hello?name=geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *gee.Context) {
			c.JSON(http.StatusOK, &map[string]string{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}

	r.Run(":9999")
}
