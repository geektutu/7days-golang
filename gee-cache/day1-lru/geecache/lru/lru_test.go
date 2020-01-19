package lru

import (
	"testing"
)

func TestGet(t *testing.T) {
	lru := New(0, nil)
	lru.Add("key1", 1234)
	if v, ok := lru.Get("key1"); !ok || v != 1234 {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}
