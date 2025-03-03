package main

import (
	geerpc "My_Geerpc"
	"My_Geerpc/registry"
	"My_Geerpc/xclient"
	"context"
	"log"
	"net"
	"net/http"
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
func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

// 创建startSrever函数
// 创建main函数
// startServer函数用于启动服务器
func startServer(registryaddr string, wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":0") //fk.key
	//声明一个Foo类型的变量
	var foo Foo
	s := geerpc.NewServer()
	_ = s.Register(&foo)
	registry.HeartBeat(registryaddr, "tcp@"+l.Addr().String(), 0)
	//log.Println("Server started on ", l.Addr())
	wg.Done()
	s.Accept(l) //fk.key
}

func call(registryaddr string) {
	d := xclient.NewGeeRegistryDiscovery(registryaddr, 0)
	xc := xclient.Newxc(d, xclient.RandomSelect, nil)
	defer func() { xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			args := &Args{Num1: i, Num2: i * i}
			foo(xc, context.Background(), "call", "Foo.Add", args)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
func broadcast(registryaddr string) {
	d := xclient.NewGeeRegistryDiscovery(registryaddr, 0)
	xc := xclient.Newxc(d, xclient.RoundRobinSelect, nil)
	defer func() { xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			args := &Args{Num1: i, Num2: i * i}
			foo(xc, context.Background(), "broadcast", "Foo.Add", args)
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", args)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
func main() {
	registryaddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()
	wg.Add(2)
	go startServer(registryaddr, &wg)
	go startServer(registryaddr, &wg)
	wg.Wait()
	time.Sleep(time.Second)
	call(registryaddr)
	broadcast(registryaddr)
}

// 分类看是调用那个
func foo(xc *xclient.XClient, ctx context.Context, typ, ServiceMethod string, args *Args) {
	var err error
	var reply int
	switch typ {
	case "call":
		err = xc.Call(ctx, ServiceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, ServiceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error:%v", typ, ServiceMethod, err)
	} else {
		log.Printf("%s %s success:%d+%d=%d", typ, ServiceMethod, args.Num1, args.Num2, reply)
	}
}
