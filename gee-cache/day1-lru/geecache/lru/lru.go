package lru

import (
	"container/list"
	"unsafe"
)

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes int
	nbytes   int
	ll       *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value interface{})
}

type entry struct {
	key   string
	value interface{}
}

// New is the Constructor of Cache
func New(maxBytes int, onEvicted func(string, interface{})) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value interface{}) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		kv.value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	c.nbytes += len(key) + sizeof(value)

	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get look ups a key's value
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= len(kv.key) + sizeof(kv.value)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Value is optional interface for the value
// if it's not implemented, use unsafe.Sizeof to count
type Value interface {
	Len() int // count how many bytes it takes
}

func sizeof(value interface{}) int {
	if m, ok := value.(Value); ok {
		return m.Len()
	}
	return int(unsafe.Sizeof(value))
}
