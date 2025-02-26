package main

import (
	geerpc "My_Geerpc"
	"My_Geerpc/xclient"
	"context"
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
func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

// 创建startSrever函数
// 创建main函数
// startServer函数用于启动服务器
func startServer(addr chan string) {
	l, _ := net.Listen("tcp", ":0") //fk.key
	//声明一个Foo类型的变量
	var foo Foo
	s := geerpc.NewServer()
	_ = s.Register(&foo)
	addr <- l.Addr().String() //fk.key
	log.Println("Server started on ", l.Addr())
	//这两步不应该和geerpc.Accept(l)一样吗
	s.Accept(l) //fk.key
	//geerpc.HandleHttp()
	//_ = http.Serve(l, nil)
}

func call(addr1, addr2 string) {
	d := xclient.NewDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.Newxc(d, xclient.RandomSelect, nil)
	defer func() { xc.Close() }()
	// time.Sleep(time.Second)
	// cl, _ := geerpc.DialHttp("tcp", <-addr)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			args := &Args{Num1: i, Num2: i * i}
			foo(xc, context.Background(), "call", "Foo.Add", args)
			// 		var reply int
			// 		ctx, _ := context.WithTimeout(context.Background(), time.Second)
			// 		cl.Call(ctx, "Foo.Add", args, &reply)
			// 		log.Printf("%d + %d = %d\n", args.Num1, args.Num2, reply)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
func broadcast(addr1, addr2 string) {
	d := xclient.NewDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.Newxc(d, xclient.RandomSelect, nil)
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
	ch1 := make(chan string)
	ch2 := make(chan string)
	startServer(ch1)
	print("1")
	startServer(ch2)
	// addr1 := <-ch1
	// addr2 := <-ch2
	// time.Sleep(time.Second)
	// call(addr1, addr2)
	// broadcast(addr1, addr2)
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
