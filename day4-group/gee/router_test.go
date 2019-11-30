package gee

import (
	"fmt"
	"reflect"
	"testing"
)

func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	testCases := [][]string{
		parsePattern("/p/:name"),
		parsePattern("/p/*"),
		parsePattern("/p/*name/*"),
	}
	wants := [][]string{
		[]string{"p", ":name"},
		[]string{"p", "*"},
		[]string{"p", "*name"},
	}
	for index, result := range testCases {
		if reflect.DeepEqual(result, wants[index]) {
			t.Fatal("test parsePattern failed")
		}
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/geektutu")

	if n == nil {
		t.Fatal("nil shouldn't be returned")
	}

	if n.pattern != "/hello/:name" {
		t.Fatal("should match /hello/:name")
	}

	if ps["name"] != "geektutu" {
		t.Fatal("name should be equal to 'geektutu'")
	}

	fmt.Printf("matched path: %s, params['name']: %s\n", n.pattern, ps["name"])

}

func TestGetRoute2(t *testing.T) {
	r := newTestRouter()
	n1, ps1 := r.getRoute("GET", "/assets/file1.txt")
	ok1 := n1.pattern == "/assets/*filepath" && ps1["filepath"] == "file1.txt"
	if !ok1 {
		t.Fatal("pattern shoule be /assets/*filepath & filepath shoule be file1.txt")
	}

	n2, ps2 := r.getRoute("GET", "/assets/css/test.css")
	ok2 := n2.pattern == "/assets/*filepath" && ps2["filepath"] == "css/test.css"
	if !ok2 {
		t.Fatal("pattern shoule be /assets/*filepath & filepath shoule be css/test.css")
	}

}

func TestGetRoutes(t *testing.T) {
	r := newTestRouter()
	nodes := r.getRoutes("GET")
	for i, n := range nodes {
		fmt.Println(i+1, n)
	}

	if len(nodes) != 5 {
		t.Fatal("the number of routes shoule be 4")
	}
}

func TestNestingGroup(t *testing.T) {
	r := &RouterGroup{
		prefix:      "/v1",
		middleWares: nil,
		engine:      nil,
		parent:      nil,
	}
	r2 := &RouterGroup{
		prefix:      "/v2",
		middleWares: nil,
		engine:      nil,
		parent:      r,
	}
	res := getNestPrefix(r2, "/hello")
	if res != "/v1/v2/hello" {
		t.Fatal("match failed")
	}
}
