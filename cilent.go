package geerpc

import (
	"My_Geerpc/codec"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// 设置Client结构体
type Client struct {
	seq      uint64
	opt      *Option
	c        codec.Codec
	pending  map[uint64]*Call
	shutdown bool       //出错强制关掉
	closing  bool       //客户端主动关掉
	mu       sync.Mutex //客户端发送多个信息给服务端，会用到上面的变量，他们在一个客户端实例中是共享的，因此用的时候需要加锁
	sending  sync.Mutex
	h        codec.Header
}

// 设置Call结构体
type Call struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Done          chan *Call
	Error         error
	Seq           uint64 //我感觉这个最后加上是因为Call那里要是出现超时
}

// 记住，给接口弄实例的时候，最好加一个这个看是否这个结构体都实现了这个实例，其实我感觉也可以不加
var _ io.Closer = (*Client)(nil)

// Client*的Close()方法
// 关闭Client
func (cl *Client) Close() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if cl.closing {
		return errors.New("client has been closed")
	}
	cl.closing = true
	return cl.c.Close()
}

// Client*的IsAvailable()方法
func (cl *Client) IsAvailable() bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return !cl.shutdown && !cl.closing
}

type result struct {
	cl  *Client
	err error
}

// 为了写client_test.go,这里先将NewCilent替换一下可改变时间的
type newClientFunc func(conn net.Conn, opt *Option) (*Client, error)

func dialtimeout(f newClientFunc, network, address string, opts ...*Option) (cl *Client, err error) {
	var opt *Option
	if len(opts) == 0 || opts[0] == nil {
		opt = &DefaultOption
	}
	//这个现在是用不上的，因为Option用的就是json
	if len(opts) == 1 {
		opt = opts[0]
		opt.MagicNumber = DefaultOption.MagicNumber
		if opt.CodecType == "" {
			opt.CodecType = DefaultOption.CodecType
		}
	}
	if len(opts) > 1 {
		log.Fatalf("Dial: wrong number of arguments, want 1, got %v", len(opts))
	}
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
	defer func() {
		if cl == nil {
			conn.Close()
		}
	}()
	res := make(chan result)
	go func() {
		cl, err := f(conn, opt)
		if err != nil {
			log.Printf("client creation failed: %v", err)
		}
		//这里我本来不想创建result的，忽略掉error,但是我不知道指针可以传进管道吗
		res <- result{cl: cl, err: err}
	}()
	if opt.ConnectTimeout == 0 {
		s := <-res
		return s.cl, s.err
	}
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: %d", opt.ConnectTimeout)
	case s := <-res:
		if s.cl == nil {
			return nil, fmt.Errorf("client creation failed: %v", s.err)
		}
		return s.cl, s.err
	}
}

// 创建一个客户端实例
func DialHttp(network, address string, opts ...*Option) (cl *Client, err error) {
	return dialtimeout(NewHttpCilent, network, address, opts...)
}
func Dial(network, address string, opts ...*Option) (cl *Client, err error) {
	return dialtimeout(NewClient, network, address, opts...)
}

// 创建一个新的HTTP客户端
func NewHttpCilent(conn net.Conn, opt *Option) (*Client, error) {
	// _, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", defaultRpcPath))
	// resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	// if err == nil && resp.Status == connected {
	// 	return NewClient(conn, opt)
	// }
	// if err == nil {
	// 	err = errors.New("unexpected HTTP response: " + resp.Status)
	// }
	// return nil, err
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", defaultRpcPath))

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return NewClient(conn, opt)
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	return nil, err
}
func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	// if opt == nil {
	// 	return nil, fmt.Errorf("opt cannot be nil")
	// }
	// _ = json.NewEncoder(conn).Encode(opt)

	// return NewCilentCodec(codec.NewCodecFuncMap[opt.CodecType](conn), opt), nil
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}
	if opt == nil {
		return nil, fmt.Errorf("opt cannot be nil")
	}

	// 确保 opt 不为 nil 后继续
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		return nil, fmt.Errorf("failed to encode option: %v", err)
	}

	return NewCilentCodec(codec.NewCodecFuncMap[opt.CodecType](conn), opt), nil
}

// http@10.0.0.1:7001
func XDial(rpcaddr string, opts ...*Option) (*Client, error) {
	parts := strings.Split(rpcaddr, "@")
	if len(parts) != 2 {
		return nil, errors.New("rpc client err:wrong @ nums")
	}
	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHttp("tcp", addr, opts...)
	default:
		//我他瞄把protocol写成了"tcp"就说unix这个咋通不过去
		return Dial(protocol, addr, opts...)
	}
}

//	func NewCilent(conn net.Conn, opt *Option) (*Client, error) {
//		_ = json.NewEncoder(conn).Encode(opt)
//		codecFunc, ok := codec.NewCodecFuncMap[opt.CodecType]
//		if !ok {
//			return nil, fmt.Errorf("invalid codec type: %s", opt.CodecType)
//		}
//		return NewCilentCodec(codecFunc(conn), opt), nil
//	}
func NewCilentCodec(c codec.Codec, opt *Option) *Client {
	cl := &Client{
		seq:     1,
		c:       c,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go cl.receive()
	return cl
}

// Client*的receive()方法，用来接收·服务器发来的消息
func (cl *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = cl.c.ReadHeader(&h); err != nil {
			log.Println("rpc client: read header error:", err)
			break
		}
		call := cl.removeCall(h.Seq)
		switch {
		case call == nil:
			err = cl.c.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = cl.c.ReadBody(nil)
			call.done()
		default:
			if err := cl.c.ReadBody(call.Reply); err != nil {
				call.Error = err
			}
			call.done()
		}

	}
	cl.terminateCall(err)
}
func (call *Call) done() {
	call.Done <- call
}

// Client*的terminateCall()方法
func (cl *Client) terminateCall(err error) {
	//这两把锁的意思到底是什么呀
	cl.sending.Lock()
	defer cl.sending.Unlock()
	cl.mu.Lock()
	defer cl.mu.Unlock()
	for _, call := range cl.pending {
		call.Error = err
		call.done()
	}
}

// 开始写关于发送信息给服务器
func (cl *Client) Call(ctx context.Context, ServiceMethod string, args, reply interface{}) error {
	call := cl.Go(ServiceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		cl.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
		//这里刚开始看的还不明白，其实就是你通过go将消息发给服务端，然后一直等，直到在receive函数中，call.done()函数将call传给call.done通道，此时受到了服务端发送过来的reply,此时<-call.done里面才不为空，不会堵塞
	case <-call.Done:
		return call.Error
	}
}

// 这个函数还不太明白
func (cl *Client) Go(servicemethod string, args, reply interface{}, done chan *Call) *Call {
	call := &Call{
		ServiceMethod: servicemethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	cl.send(call)
	return call
}
func (cl *Client) send(call *Call) {
	//sending解决并发问题
	cl.sending.Lock()
	defer cl.sending.Unlock()
	seq := cl.registerCall(call)
	// var h codec.Header
	// h.ServiceMethod = call.ServiceMethod
	//不能写成上面那个，有严重问题
	//加上h.Seq = seq,这样在收到信息时，才能找到对应的call，err := cl.c.Write(&cl.h, call.Args);seq和call是这样关联起来的
	cl.h.ServiceMethod = call.ServiceMethod
	cl.h.Seq = seq
	cl.h.Error = ""

	if err := cl.c.Write(&cl.h, call.Args); err != nil {
		cl.removeCall(seq)
	}
}

// Client*的registerCall()方法
func (cl *Client) registerCall(call *Call) uint64 {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	Seq := cl.seq
	call.Seq = Seq
	cl.pending[cl.seq] = call
	cl.seq++
	return Seq
}
func (cl *Client) removeCall(seq uint64) *Call {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	call := cl.pending[seq]
	delete(cl.pending, seq)
	return call
}
