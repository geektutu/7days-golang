package gee

import "testing"

func TestNestingGroup(t *testing.T) {
	v1 := &RouterGroup{
		prefix: "/v1",
	}
	v2 := &RouterGroup{
		prefix: "/v2",
		parent: v1,
	}
	v3 := &RouterGroup{
		prefix: "/v3",
		parent: v2,
	}
	if v2.getNestPrefix() != "/v1/v2" {
		t.Fatal("v2 prefix should be /v1/v2")
	}
	if v3.getNestPrefix() != "/v1/v2/v3" {
		t.Fatal("v3 prefix should be /v1/v2/v3")
	}
}

func TestGroup(t *testing.T) {
	r := New()
	v1 := r.Group("/v1")
	v2 := v1.Group("/v2")
	if v2.getNestPrefix() != "/v1/v2" {
		t.Fatal("v2 prefix should be /v1/v2")
	}
}
