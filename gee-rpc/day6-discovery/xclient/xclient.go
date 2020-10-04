package xclient

import (
	"context"
	. "geerpc"
	"io"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *Option
	clients sync.Map
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt}
}

func (xc *XClient) Close() error {
	xc.clients.Range(func(k, v interface{}) bool {
		// I have no idea how to deal with error, just ignore it.
		_ = v.(*Client).Close()
		return true
	})
	xc.clients = sync.Map{}
	return nil
}

// Call invokes the named function, waits for it to complete,
// and returns its error status.
// xc will choose a proper server.
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr := xc.d.Get(xc.mode)
	client, ok := xc.clients.Load(rpcAddr)
	if !ok {
		var err error
		client, err = XDial(rpcAddr, xc.opt)
		if err != nil {
			return err
		}
		xc.clients.Store(rpcAddr, client)
	}
	return client.(*Client).Call(ctx, serviceMethod, args, reply)
}
