package gee

import (
	"strings"
)

type router struct {
	root *node
}

func (r *router) addRoute(path string, handler HandlerFunc) {
	parts := filterNonEmpty(strings.Split(path, "/"))
	r.root.insert(path, parts, handler, 0)
}

func (r *router) getRoute(path string) (*node, *Params) {
	searchParts := filterNonEmpty(strings.Split(path, "/"))
	n := r.root.search(searchParts, 0)

	if n.path != "" {
		parts := filterNonEmpty(strings.Split(n.path, "/"))
		params := &Params{}

		for index, part := range parts {
			if part[0] == ':' {
				params.set(part[1:], searchParts[index])
			}
		}
		return n, params
	}

	return nil, nil
}

func (r *router) getRoutes() []*node {
	list := make([]*node, 0)
	r.root.travel(&list)
	return list
}

func filterNonEmpty(vs []string) []string {
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
		}
	}
	return parts
}
