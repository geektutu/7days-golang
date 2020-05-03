package server

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"geerpc/protocol"
)

type Server struct {
	ln      net.Listener
	service map[string]*service
}

func NewServer() *Server {
	return &Server{
		service: make(map[string]*service),
	}
}

func (s *Server) Address() net.Addr {
	return s.ln.Addr()
}

func (s *Server) Serve(network, address string) (err error) {
	if network == "http" {
		if s.ln, err = net.Listen("tcp", address); err != nil {
			return err
		}
		http.Handle(protocol.DefaultRPCPath, s)
		return http.Serve(s.ln, nil)
	}
	panic(network + " not support")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m, err := protocol.Read(req.Body)
	if err != nil {
		log.Println("failed to read message from body")
		_, _ = w.Write([]byte("fail"))
		return
	}

	log.Println(req.Method, m.ServiceMethod)

	respMsg := s.call(m)
	_ = respMsg.Write(w)
}

func (s *Server) Register(receiver interface{}) {
	service := newService(receiver)
	s.service[service.name] = service
}

func (s *Server) call(req *protocol.Message) (resp *protocol.Message) {
	serviceName, methodName, err := req.GetServiceMethod()
	resp = req.Clone()
	if err != nil {
		return resp.HandleError(protocol.NotFoundError, err)
	}

	service := s.service[serviceName]
	if service == nil || service.method[methodName] == nil {
		return resp.HandleError(protocol.NotFoundError, fmt.Errorf("%s not found", req.ServiceMethod))
	}

	return service.call(methodName, req)
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf(msg, v...))
	}
}
