package xclient

//支持负载均衡的客户端
import (
	geerpc "My_Geerpc"
	"context"
	"io"
	"sync"
)

type XClient struct {
	d       Discovery
	mu      sync.Mutex
	clients map[string]*geerpc.Client
	mod     SelectMode
	opt     *geerpc.Option
}

// 先获取一个服务实例
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
