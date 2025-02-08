package geerpc

import (
	"My_Geerpc/codec"
	"encoding/json"
	"log"
	"net"
)

// 设置Client结构体
type Client struct {
	seq     uint64
	opt     *Option
	c       codec.Codec
	pending map[uint64]*Call
}

// 设置Call结构体
type Call struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Done          chan *Call
}

// Client*的Close()方法
// 关闭Client
func (cl *Client) Close() {

}

// Client*的IsAvailable()方法
func (cl *Client) IsAvailable() bool {
	//这里1和true是不一样的
	return true
}

//Client*的terminateCall()方法
// func (cl *Client)terminateCall(){

// }
// 创建一个客户端实例
func Dial(network, address string, opts ...*Option) (cl *Client) {
	//这里先强制使用DefaultOption
	var opt *Option
	if len(opts) == 0 || opts[0] == nil {
		opt = &DefaultOption
	}
	if len(opts) == 1 {
		opt = opts[0]
	}
	if len(opts) != 1 {
		log.Fatalf("Dial: wrong number of arguments, want 1, got %v", len(opts))
	}
	conn, _ := net.Dial(network, address)
	defer func() {
		if cl == nil {
			conn.Close()
		}
	}()
	//这里
	return NewCilent(conn, opt)
}
func NewCilent(conn net.Conn, opt *Option) *Client {
	_ = json.NewEncoder(conn).Encode(opt)
	return NewCilentCodec(codec.NewCodecFuncMap[opt.CodecType](conn), opt)
}
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
			break
		}
		call := cl.removeCall(h.Seq)
		cl.c.ReadBody(call.Reply)
		call.done()
	}
}
func (call *Call) done() {
	call.Done <- call
}

// 开始写关于发送信息给服务器
func (cl *Client) Call(ServiceMethod string, args, reply interface{}) {
	cl.Go(ServiceMethod, args, reply, make(chan *Call, 1))
}
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
	seq := cl.registerCall(call)
	var h codec.Header
	h.ServiceMethod = call.ServiceMethod
	if err := cl.c.Write(&h, call.Args); err != nil {
		cl.removeCall(seq)
	}
}

// Client*的registerCall()方法
func (cl *Client) registerCall(call *Call) uint64 {
	Seq := cl.seq
	cl.pending[cl.seq] = call
	cl.seq++
	return Seq
}
func (cl *Client) removeCall(seq uint64) *Call {
	call := cl.pending[seq]
	delete(cl.pending, seq)
	return call
}
