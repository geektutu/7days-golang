package main

import (
	"fmt"
	"geerpc"
	"log"
	"net"
)

func startServer(addr chan string) {
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)
	client, _ := geerpc.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	// send request & receive response
	for i := 0; i < 3; i++ {
		args := fmt.Sprintf("geerpc req %d", i)
		var reply string
		if err := client.Call("Foo.Sum", args, &reply); err != nil {
			log.Fatal("call Foo.Sum error", err)
		}
		log.Println("reply:", reply)
	}
}
