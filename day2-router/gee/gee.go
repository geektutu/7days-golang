package gee

import (
	"fmt"
	"net/http"
)

type Params map[string]string

func (ps *Params) Get(key string) string {
	if value, ok := (*ps)[key]; ok {
		return value
	}
	return ""
}

func (ps *Params) set(key string, value string) {
	(*ps)[key] = value
}

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(http.ResponseWriter, *http.Request, *Params)

// Engine is defined to handle all requests
type Engine struct {
	router *router
}

// New is constructor of Engine
func New() *Engine {
	return &Engine{
		router: &router{root: &node{}},
	}
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.router.addRoute(pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if n, params := engine.router.getRoute(req.URL.Path); n != nil {
		n.handler(w, req, params)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
