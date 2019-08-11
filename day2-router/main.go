package main

// $ curl http://localhost:9999/
// URL.Path = "/"
// $ curl http://localhost:9999/hello
// Header["Accept"] = ["*/*"]
// Header["User-Agent"] = ["curl/7.54.0"]
// curl http://localhost:9999/world
// 404 NOT FOUND: /world

import (
	"fmt"
	"net/http"

	"./gee"
)

func main() {
	r := gee.New()
	r.GET("/", func(w http.ResponseWriter, req *http.Request, params *gee.Params) {
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})

	r.GET("/hello/:name", helloHandler)
	r.Run(":9999")
}

func helloHandler(w http.ResponseWriter, req *http.Request, params *gee.Params) {
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	fmt.Fprintf(w, "Parse params in path, name:  %s\n", params.Get("name"))
}
