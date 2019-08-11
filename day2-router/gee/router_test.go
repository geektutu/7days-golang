package gee

import (
	"fmt"
	"testing"
)

func newTestRouter() *router {
	r := &router{root: &node{}}
	r.addRoute("/", nil)
	r.addRoute("/hello/:name", nil)
	r.addRoute("/hello/b/c", nil)
	r.addRoute("/hi/:name", nil)
	return r
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("/a/geektutu")

	fmt.Printf("matched path: %s, params['name']: %s\n", n.path, ps.Get("name"))

	if n == nil {
		t.Fatal("nil shouldn't be returned")
	}

	if n.path != "/a/:name" {
		t.Fatal("should match /a/:name")
	}

	if ps.Get("name") != "geektutu" {
		t.Fatal("name should be equal to 'geektutu'")
	}

}

func TestGetRoutes(t *testing.T) {
	r := newTestRouter()
	for i, n := range r.getRoutes() {
		fmt.Println(i+1, n)
	}

	if len(r.getRoutes()) != 4 {
		t.Fatal("the number of routes shoule be 4")
	}
}
