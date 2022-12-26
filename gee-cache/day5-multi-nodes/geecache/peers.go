package geecache

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
// PeerGetter 对应于 HTTP 客户端。
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
