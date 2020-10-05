package xclient

import (
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota // select randomly
	RobinSelect                    // select using Robbin algorithm
)

type Discovery interface {
	Get(mode SelectMode) string
	All() []string
}

var _ Discovery = (*MultiServersDiscovery)(nil)

// MultiServersDiscovery is a discovery for multi servers without a registry center
// user provides the server addresses explicitly instead
type MultiServersDiscovery struct {
	r       *rand.Rand   // generate random number
	mu      sync.RWMutex // protect following
	servers []string
	index   int // record the selected position for robin algorithm
}

// Update the servers of discovery dynamically if needed
func (d *MultiServersDiscovery) Update(servers []string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
}

func (d *MultiServersDiscovery) Get(mode SelectMode) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.servers) == 0 {
		return ""
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(len(d.servers))]
	case RobinSelect:
		s := d.servers[d.index]
		d.index = (d.index + 1) % len(d.servers)
		return s
	default:
		return ""
	}
}

func (d *MultiServersDiscovery) All() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	// return a copy of d.servers
	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers
}

// NewMultiServerDiscovery creates a MultiServersDiscovery instance
func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	return &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
