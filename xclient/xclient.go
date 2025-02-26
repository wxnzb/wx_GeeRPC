package xclient

//支持负载均衡的客户端
import (
	geerpc "My_Geerpc"
	"context"
	"io"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discovery
	mu      sync.Mutex
	clients map[string]*geerpc.Client
	mod     SelectMode
	opt     *geerpc.Option
}

func Newxc(di Discovery, mod SelectMode, opt *geerpc.Option) *XClient {
	return &XClient{
		d:       di,
		mod:     mod,
		clients: make(map[string]*geerpc.Client),
	}
}

// 先获取一个服务实例,关键突破点是根据xc.mod获得一个服务器地址，在xc的mao中获得这个地址对应的客户端实例，然后就是正常调用
func (xc *XClient) Call(ctx context.Context, ServiceMethod string, args, reply interface{}) error {
	rpcaddr, err := xc.d.Get(xc.mod)
	if err != nil {
		return err
	}
	return xc.call(rpcaddr, ctx, ServiceMethod, args, reply)
}
func (xc *XClient) call(rpcaddr string, ctx context.Context, ServiceMethod string, args, reply interface{}) error {
	//根据服务地址获得客户端实例
	client, err := xc.dial(rpcaddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, ServiceMethod, args, reply)
}
func (xc *XClient) dial(rpcaddr string) (*geerpc.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[rpcaddr]
	if !ok && client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, rpcaddr)
		client = nil
	}
	if client == nil {
		client, err := geerpc.XDial(rpcaddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcaddr] = client
	}
	return client, nil
}

// 这个和上面一样
func (xc *XClient) Broadcast(ctx context.Context, ServiceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error = nil
	replydone := reply == nil
	ctx, cancel := context.WithCancel(ctx)
	for _, serveraddr := range servers {
		wg.Add(1)
		go func(serveraddr string) {
			defer wg.Done()
			var cloneReply interface{}
			if reply != nil {
				cloneReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(serveraddr, ctx, ServiceMethod, args, cloneReply)
			mu.Lock() //这里加锁是为了防止多个协程同时修改e
			if err != nil && e == nil {
				e = err
				cancel()
			}
			if !replydone && err == nil {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(cloneReply).Elem())
				replydone = true
			}
			mu.Unlock()
		}(serveraddr)
	}
	wg.Wait()
	return e
}

// 顺便设置一下多个关闭多个客户端
var _ io.Closer = (*XClient)(nil)

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}
