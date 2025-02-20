// 创建Option结构体
// 创建默认Option结构体
package geerpc

import (
	"My_Geerpc/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const Magicnumber = 0x3bef5c

type Option struct {
	MagicNumber    int
	CodecType      string
	ConnectTimeout time.Duration
	HandleTimeout  time.Duration
}

var DefaultOption = Option{
	MagicNumber:    Magicnumber,
	CodecType:      "gob",
	ConnectTimeout: time.Second * 10,
}

// 创见server结构体
// 创建一个函数，来处理客户端连接
// 创建默认server结构体,来处理默认的连接
type Server struct {
	serviceMap sync.Map
}

func (s *Server) Register(rvcr interface{}) error {
	m := NewService(rvcr)
	if _, dup := s.serviceMap.LoadOrStore(m.name, m); dup {
		return errors.New("rpc server:service already defined:" + m.name)
	}
	return nil
}
func Register(rvcr interface{}) error {
	return DefaultServer.Register(rvcr)
}
func (s *Server) FindService(servicemethod string) (*Service, *methodType, error) {
	dot := strings.LastIndex(servicemethod, ".")
	if dot < 0 {
		return nil, nil, errors.New("no . ,fomat erroe")
	}
	serviceName, methodName := servicemethod[:dot], servicemethod[dot+1:]
	sv, ok := s.serviceMap.Load(serviceName)
	if !ok {
		return nil, nil, errors.New("can't find service ")
	}
	svi := sv.(*Service)
	method := svi.serthods[methodName]
	if method == nil {
		return nil, nil, errors.New("can't find method ")
	}
	return svi, method, nil

}
func NewServer() *Server {
	return &Server{}
}
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept() //fk.key
		if err != nil {
			log.Println("accept error:")
		}
		go s.ServerConn(conn)
	}
}

var DefaultServer = NewServer()

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

// 创建ServerConn函数，用json解析Option结构体
func (s *Server) ServerConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("decode error:", err)
		return
	}
	if opt.MagicNumber != Magicnumber {
		log.Printf("invalid magic number %x", opt.MagicNumber)
		return
	}
	if opt.CodecType != "gob" {
		log.Println("invalid codec type")
		return
	}
	//直接调用可能有问题，需要改
	s.ServerCodec(codec.NewCodecFuncMap[opt.CodecType](conn), &opt)
}

// 定义request结构体
type Request struct {
	h            *codec.Header
	argv, replyv reflect.Value
	svc          *Service
	mtype        *methodType
}

// 创建readRequest函数
func (s *Server) readRequest(c codec.Codec) (req *Request, err error) {
	var h codec.Header
	if err := c.ReadHeader(&h); err != nil {
		log.Println("read header error:", err)
		return nil, err
	}
	req = &Request{h: &h}
	// req.argv = reflect.New(reflect.TypeOf(""))               //go
	// if err := c.ReadBody(req.argv.Interface()); err != nil { //go
	// 	log.Println("read body error:", err)
	// 	//不应该返回nil,因为req不为nil,有header
	// 	return req, err
	// }
	req.svc, req.mtype, err = s.FindService(h.ServiceMethod)
	if err != nil {
		return
	}
	req.argv = req.mtype.NewArgv()
	req.replyv = req.mtype.NewReply()
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err = c.ReadBody(argvi); err != nil {
		log.Println("read body error:", err)
		return
	}
	return
}

// 创建ServerCodec函数
// ServerCodec方法用于处理客户端请求
var invaildRequest = struct{}{}

func (s *Server) ServerCodec(c codec.Codec, opt *Option) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	defer c.Close()
	for {
		req, err := s.readRequest(c)
		if err != nil {
			if req == nil {
				break
			} else {
				req.h.Error = err.Error()
				//错误应该立即返回给客户端
				s.sendResponse(c, req.h, invaildRequest, sending)
				continue
			}
		}
		//之前写成了c.h.Seq,你要始终记住此时的codec.Codec是*Gobcodec
		wg.Add(1)
		//req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp: %d", req.h.Seq))
		//log.Println(req.h, req.argv.Elem())
		go s.sendHandle(c, req.h, req, sending, wg, opt.HandleTimeout) //这里先随便设置一下
	}
	wg.Wait()
}

// 创建sendResponse函数
func (s *Server) sendResponse(c codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := c.Write(h, body); err != nil {
		log.Println("write response error:", err)
	}
}

// 这里只判断处理call是否超时，sent的作用是为了让sendResponse执行完后在退出程序
func (s *Server) sendHandle(c codec.Codec, h *codec.Header, req *Request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.Call(req.mtype, req.argv, req.replyv) //处理可能超时
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			s.sendResponse(c, h, invaildRequest, sending) //发送信息
			sent <- struct{}{}
			return
		}
		s.sendResponse(c, h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()
	if timeout == 0 {
		<-called
		<-sent
	}
	select {
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server: request timein: %d", timeout)
		s.sendResponse(c, h, invaildRequest, sending)
	case <-called:
		<-sent
	}
	//s.sendResponse(c, h, body, sending)
}

// 加入Https
const (
	connected      = "200 Connected to Gee RPc"
	defaultRpcPath = "/_geerpc_"
)

func (s *Server) ServeHttp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain;charset=uft-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must connect")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rcp hijacking:", r.RemoteAddr, ":", err.Error())
		return
	}
	_, _ = io.WriteString(conn, "HTTP/1.1"+connected+"\n\n")
	s.ServerConn(conn)

}
func (s *Server) HandleHttp() {
	http.Handle(defaultRpcPath, s)
}
func HandleHttp() {
	DefaultServer.HandleHttp()
}
