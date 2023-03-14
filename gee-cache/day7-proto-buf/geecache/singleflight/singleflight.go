package singleflight

import (
	"sync"
	"time"
)

// call is an in-flight or completed Do call
type call struct {
	val interface{}
	err error
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		return c.val, c.err //Subsequent requests get the lock and get data directly from the cache that is delayed-delete
	}
	c := new(call)
	g.m[key] = c
	c.val, c.err = fn() //Call fn, make a request, make sure fn is only called once
	go func() {
		time.Sleep(time.Second)
		delete(g.m, key) //Delay deletion for one second
	}()

	return c.val, c.err //返回结果
}
