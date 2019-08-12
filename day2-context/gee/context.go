package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Context struct {
	Path   string
	Method string
	Writer http.ResponseWriter
	Req    *http.Request
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Path:   req.URL.Path,
		Method: req.Method,
		Req:    req,
		Writer: w,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.Status(code)
	c.SetHeader("Content-Type", "text/plain")
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) HTML(code int, html string) {
	c.Writer.WriteHeader(code)
	c.SetHeader("Content-Type", "text/html")
	c.Writer.Write([]byte(html))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.Writer.WriteHeader(code)
	c.SetHeader("Content-Type", "application/json")
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Writer.WriteHeader(code)
	c.Writer.Write(data)
}
