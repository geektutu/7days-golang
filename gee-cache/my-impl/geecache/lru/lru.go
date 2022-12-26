package lru

import "container/list"

type Cache struct {
	nBytes    int64
	maxBytes  int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{maxBytes: maxBytes, ll: list.New(), cache: make(map[string]*list.Element), OnEvicted: onEvicted}
}

type entry struct {
	key   string
	value Value
}

func (c *Cache) Add(key string, value Value) {
	if node, ok := c.cache[key]; ok {
		c.ll.MoveToFront(node)
		kv := node.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = c.ll.Front()
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if node, ok := c.cache[key]; ok {
		c.ll.MoveToFront(node)
		return node.Value.(*entry).value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	if c.ll.Len() == 0 {
		return
	}
	oldest := c.ll.Back()
	defer c.ll.Remove(oldest)

	kv := oldest.Value.(*entry)
	c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
