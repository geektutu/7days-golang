package geecache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"Tom":  "123",
	"Jack": "456",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	gee := NewGroup("demo", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Printf("search key %s", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	for k, v := range db {
		view, err := gee.Get(k)
		if err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
