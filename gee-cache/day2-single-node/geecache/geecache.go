package geecache

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// NewGroup create a new instance of Group
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

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}

	return g.load(key)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (g *Group) load(key string) (value ByteView, err error) {
	bytes, err := g.getter.Get(key)
	if err == nil {
		value = ByteView{cloneBytes(bytes)}
		g.populateCache(key, value)
	}
	return
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
