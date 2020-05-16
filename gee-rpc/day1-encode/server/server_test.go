package server

import (
	"bytes"
	"fmt"
	"geerpc/protocol"
	"net"
	"net/http"
	"testing"
	"time"
)

type Calc struct{}

type Req struct {
	Num1 int
	Num2 int
}

func (c *Calc) Add(req Req, reply *int) error {
	*reply = req.Num1 + req.Num2
	return nil
}

func TestServer_Register(t *testing.T) {
	s := NewServer()
	s.Register(&Calc{})

	service := s.service["Calc"]
	if service == nil || service.method["Add"] == nil {
		t.Fatal("failed to register")
	}
}

func TestServer_Call(t *testing.T) {
	s := NewServer()
	s.Register(&Calc{})
	req := &Req{Num1: 10, Num2: 20}

	reqMsg := protocol.NewMessage()
	reqMsg.SetServiceMethod("Calc.Add")
	_ = reqMsg.SetPayload(req)

	respMsg := s.call(reqMsg)
	var ans int
	_ = respMsg.GetPayload(&ans)
	if ans != req.Num1+req.Num2 {
		t.Fatal("failed to call Calc.Add")
	}
}

func TestServer_Serve(t *testing.T) {
	s := NewServer()
	s.Register(&Calc{})
	go func() { _ = s.Serve("http", ":0") }()

	time.Sleep(time.Second)
	port := s.Address().(*net.TCPAddr).Port
	addr := fmt.Sprintf("http://localhost:%d%s", port, protocol.DefaultRPCPath)

	reqMsg := protocol.NewMessage()
	reqMsg.SetServiceMethod("Calc.Add")
	_ = reqMsg.SetPayload(&Req{1, 2})

	var buf bytes.Buffer
	_ = reqMsg.Write(&buf)
	resp, _ := http.Post(addr, "application/octet-stream", &buf)

	respMsg, _ := protocol.Read(resp.Body)

	var ans int
	_ = respMsg.GetPayload(&ans)
	if respMsg.Status != protocol.OK || ans != 3 {
		t.Fatal("failed to call Calc.Add")
	}
}
