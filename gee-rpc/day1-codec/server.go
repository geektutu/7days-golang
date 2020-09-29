package geerpc

import (
	"encoding/json"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct{}

const MagicNumber = 0x3bef5c

type Options struct {
	MagicNumber int        // MagicNumber marks this's a geerpc request
	CodecType   codec.Type // client may choose different Codec to encode body
}

func newServer() *Server {
	return &Server{}
}

var DefaultServer = newServer()

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Options
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.ServeCodec(f(conn))
}

// request stores all information of a call
type request struct {
	h    *codec.Header // header of request argv
	argv string        // TODO suppose argv is a string
}

var requestPool = sync.Pool{
	New: func() interface{} { return &request{} },
}

func (server *Server) readRequestHeader(cc codec.Codec) (req *request, keepReading bool, err error) {
	req, _ = requestPool.Get().(*request)
	h, _ := codec.HeaderPool.Get().(*codec.Header)
	if err = cc.ReadHeader(h); err != nil {
		log.Println("rpc server: read header error:", err)
		return
	}
	// We read the header successfully. If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true
	req.h = h
	return
}

func (server *Server) readRequest(cc codec.Codec) (req *request, keepReading bool, err error) {
	req, keepReading, err = server.readRequestHeader(cc)
	if err != nil {
		// discard argv
		_ = cc.ReadBody(nil)
		return
	}

	// We read the header successfully. If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true

	// TODO: suppose argv is a string, now we can't judge the type of request argv
	var str string
	if err = cc.ReadBody(&str); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	req.argv = str
	return
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) Handle(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	// TODO, should call registered rpc methods
	// day 1 just print argv and send a hello message
	defer wg.Done()
	log.Println(req.h, req.argv)
	server.sendResponse(cc, req.h, fmt.Sprintf("geerpc resp %d", req.h.Seq), sending)
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) ServeCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // ensure header and argv is not separated by other response
	wg := new(sync.WaitGroup)  // wait until all request are handled
	for {
		req, keepReading, err := server.readRequest(cc)
		if err != nil {
			if !keepReading {
				break // it's not possible to recover, so close the connection
			}
			if req != nil {
				req.h.Error = err.Error()
				server.sendResponse(cc, req.h, invalidRequest, sending)
			}
			continue
		}
		wg.Add(1)
		go server.Handle(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
