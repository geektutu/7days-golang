package geebolt

import (
	"bytes"
	"sort"
)

type kv struct {
	key   []byte
	value []byte
}

type node struct {
	isLeaf   bool
	key      []byte
	parent   *node
	children []*node
	kvs      []kv
}

func (n *node) root() *node {
	if n.parent == nil {
		return n
	}
	return n.parent.root()
}

func (n *node) index(key []byte) (index int, exact bool) {
	index = sort.Search(len(n.kvs), func(i int) bool {
		return bytes.Compare(n.kvs[i].key, key) != -1
	})
	exact = len(n.kvs) > 0 && index < len(n.kvs) && bytes.Equal(n.kvs[index].key, key)
	return
}

func (n *node) put(oldKey, newKey, value []byte) {
	index, exact := n.index(oldKey)
	if !exact {
		n.kvs = append(n.kvs, kv{})
		copy(n.kvs[index+1:], n.kvs[index:])
	}
	kv := &n.kvs[index]
	kv.key = newKey
	kv.value = value
}

func (n *node) del(key []byte)  {
	index, exact := n.index(key)
	if exact {
		n.kvs = append(n.kvs[:index], n.kvs[index+1:]...)
	}
}

