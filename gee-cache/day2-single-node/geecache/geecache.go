package geecache

import "errors"

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

type GetterFunc func(key string, dest Sink) error

func (f GetterFunc) Get(key string, dest Sink) error {
	return f(key, dest)
}

type Getter interface {
	Get(key string, dest Sink) error
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	return &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
}

func (g *Group) load(key string, dest Sink) (ByteView, error) {
	value, err := g.getLocally(key, dest)
	if err != nil {
		return value, err
	}

	g.populateCache(key, value)
	return value, nil
}

func (g *Group) getLocally(key string, dest Sink) (ByteView, error) {
	if err := g.getter.Get(key, dest); err != nil {
		return ByteView{}, err
	}
	return dest.view()
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) Get(key string, dest Sink) error {
	if dest == nil {
		return errors.New("groupcache: nil dest Sink")
	}
	if v, ok := g.mainCache.get(key); ok {
		return dest.SetBytes(v.b)
	}

	_, err := g.load(key, dest)
	return err
}
