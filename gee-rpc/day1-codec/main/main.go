package main

import (
	"encoding/json"
	"fmt"
	"geerpc"
	"geerpc/codec"
	"log"
	"net"
	"net/http"
	"time"
)

func startServer() {
	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Panic("network error:", err)
	}
	geerpc.Accept(l)

}

func main() {
	log.SetPrefix("")
	go startServer()
	time.Sleep(time.Second)

	// In fact, following code is like a simple GeeRPC Client
	conn, _ := net.Dial("tcp", ":9999")
	defer func() { _ = conn.Close() }()

	// negotiate options
	_ = json.NewEncoder(conn).Encode(&geerpc.Options{
		MagicNumber: geerpc.MagicNumber,
		CodecType:   codec.GobType,
	})

	cc := codec.NewGobCodec(conn)
	// send request & receive response
	for i := 0; i < 3; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply:", reply)
	}
	log.Fatal(http.ListenAndServe(":9999", nil))
}
