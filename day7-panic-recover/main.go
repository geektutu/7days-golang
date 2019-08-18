package main

/*
$ curl "http://localhost:9999"
Hello Geektutu
$ curl "http://localhost:9999/panic"
{"message":"Internal Server Error"}

>>> log
2019/08/18 17:55:57 [200] / in 4.533µs
2019/08/18 17:55:58 runtime error: index out of range
Traceback:
	/usr/local/Cellar/go/1.12.5/libexec/src/runtime/panic.go:523
	/usr/local/Cellar/go/1.12.5/libexec/src/runtime/panic.go:44
	/Users/geektutu/7days-golang/day7-panic-recover/main.go:20
	/Users/geektutu/7days-golang/day7-panic-recover/gee/context.go:41
	/Users/geektutu/7days-golang/day7-panic-recover/gee/recovery.go:37
	/Users/geektutu/7days-golang/day7-panic-recover/gee/context.go:41
	/Users/geektutu/7days-golang/day7-panic-recover/gee/logger.go:15
	/Users/geektutu/7days-golang/day7-panic-recover/gee/context.go:41
	/Users/geektutu/7days-golang/day7-panic-recover/gee/router.go:99
	/Users/geektutu/7days-golang/day7-panic-recover/gee/gee.go:129
	/usr/local/Cellar/go/1.12.5/libexec/src/net/http/server.go:2775
	/usr/local/Cellar/go/1.12.5/libexec/src/net/http/server.go:1879
	/usr/local/Cellar/go/1.12.5/libexec/src/runtime/asm_amd64.s:1338

2019/08/18 17:55:58 [500] /panic in 143.086µs
*/

import (
	"net/http"

	"./gee"
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
