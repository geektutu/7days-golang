package geecache

import (
	"errors"
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
	gee := NewGroup("demo", 2<<10, GetterFunc(func(key string, dest Sink) error {
		log.Printf("search key %s", key)
		if v, ok := db[key]; ok {
			return dest.SetBytes([]byte(v))
		}
		return errors.New(fmt.Sprintf("%s not exist", key))
	}))

	var dst []byte
	dest := AllocatingByteSliceSink(&dst)

	for k, v := range db {
		if err := gee.Get(k, dest); err != nil || string(dst) != v {
			t.Fatal("failed to get value of Tom")
		}
	}

	if err := gee.Get("unknown", dest); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", string(dst))
	}
}
