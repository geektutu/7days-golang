package main

import (
	"context"
	"geerpc"
	"log"
	"net/http"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addr string) {
	var foo Foo
	_ = geerpc.Register(&foo)
	geerpc.HandleHTTP()
	log.Fatal(http.ListenAndServe(addr, nil))
}

func call() {
	// start server may cost some time
	time.Sleep(time.Second)
	client, _ := geerpc.DialHTTP("tcp", ":9999")
	defer func() { _ = client.Close() }()

	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	go call()
	startServer(":9999")
}
