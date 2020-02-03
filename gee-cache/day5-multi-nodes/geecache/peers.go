package geecache

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

// NoPeers is an implementation of PeerPicker that never finds a peer.
type NoPeers struct{}

// PickPeer return nothing
func (NoPeers) PickPeer(key string) (peer PeerGetter, ok bool) { return }

var portPicker func() PeerPicker

// RegisterPeerPicker registers the peer initialization function.
// It is called once, when the first group is created.
func RegisterPeerPicker(fn func() PeerPicker) {
	if portPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	portPicker = fn
}

func getPeers() PeerPicker {
	if portPicker == nil {
		return NoPeers{}
	}
	pk := portPicker()
	if pk == nil {
		pk = NoPeers{}
	}
	return pk
}
