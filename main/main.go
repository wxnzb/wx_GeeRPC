package main

import (
	geerpc "My_Geerpc"
	"log"
	"net"
	"sync"
	"time"
)

type Foo int
type Args struct {
	Num1, Num2 int
}

func (f Foo) Add(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// 创建startSrever函数
// 创建main函数
func startServer(addr chan string) {
	var foo Foo
	err := geerpc.Register(&foo)
	if err != nil {
		log.Println("register error:", err)
	}
	l, _ := net.Listen("tcp", ":0") //fk.key
	addr <- l.Addr().String()       //fk.key
	log.Println("Server started on ", l.Addr())
	geerpc.Accept(l) //fk.key
}

func main() {
	addr := make(chan string)
	go startServer(addr)
	time.Sleep(time.Second)
	cl := geerpc.Dial("tcp", <-addr)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			cl.Call("Foo.Add", args, &reply)
			log.Printf("%d + %d = %d\n", args.Num1, args.Num2, reply)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
