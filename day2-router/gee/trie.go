package gee

import (
	"fmt"
)

type node struct {
	path     string
	part     string
	children []*node
	isWild   bool
	handler  HandlerFunc
}

func (n *node) String() string {
	return fmt.Sprintf("node{path=%s, part=%s, isWild=%t}", n.path, n.part, n.isWild)
}

func (n *node) insert(path string, parts []string, handler HandlerFunc, height int) {
	if len(parts) == height {
		n.path = path
		n.handler = handler
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':'}
		n.children = append(n.children, child)
	}
	child.insert(path, parts, handler, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height {
		if n.path == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

func (n *node) travel(list *([]*node)) {
	if n.path != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}

func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
