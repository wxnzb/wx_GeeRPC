// 创建Option结构体
// 创建默认Option结构体
package geerpc

import (
	"My_Geerpc/codec"
	"encoding/json"
	"fmt"
	"go/ast"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const Magicnumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   string
}

var DefaultOption = Option{
	MagicNumber: Magicnumber,
	CodecType:   "gob",
}

type methodType struct {
	Method reflect.Method
	Argv   reflect.Type
	Reply  reflect.Type
}

// 创见server结构体
// 创建一个函数，来处理客户端连接
// 创建默认server结构体,来处理默认的连接
type Server struct {
	name     string
	//要rv,rt有啥用阿
	rv       reflect.Value
	rt       reflect.Type
	serthods map[string]*methodType
}
func 
// sevicemethod eg:wx.nzb,wx.jy...就是关于wx的方法func(w *wx)nzb(s string,i int)erroe{}
func NewServer(smh interface{}) *Server {
	s := new(Server)
	//s.name=smh.(string)不能这样,为啥
	s.name = reflect.Indirect(reflect.ValueOf(smh)).Type().Name()
	//这个是必须的，他必须是Wx.nzb导出的才行
	//Go 的 net/rpc 规范 中，小写的结构体不能注册方法，因为小写的结构体字段在包外是不可见的，无法被 RPC 调用。
	if ast.IsExported(s.name) == false {
		log.Fatalf("rpc server: %s is not a valid name", s.name)
	}
	s.rv = reflect.ValueOf(smh)
	s.rt = reflect.TypeOf(smh)
	NewMethods(s)
	return s
	//return &Server{}
}
func NewMethods(s *Server) {
	s.serthods = make(map[string]*methodType)
	for i := 0; i < s.rt.NumMethod(); i++ {
		method := s.rt.Method(i)
		if method.Type.NumIn() != 3 || method.Type.NumOut() != 1 {
			continue
		}
		//方法里面的参数类型不能是自定义的小写，不能导出
		if isExportedCan(method.Type.In(1)) || isExportedCan(method.Type.In(2)) {
			continue
		}
		if method.Type.Out(0) != reflect.TypeOf((*error)(nil)) {
			continue
		}
		s.serthods[method.Name] = &methodType{
			Method: method, //其实我感觉不太需要这个呀
			Argv:   method.Type.In(1),
			Reply:  method.Type.In(2),
		}
		log.Printf("register method:%s.%s", s.name, method.Name)
	}
}
func isExportedCan(t reflect.Type) bool {
	//需要ast.IsExported(t.Name())吗
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}
func (s *Server) Call(m *methodType, argv, reply reflect.Value) error {
	f := m.Method.Func
	returnvalues := f.Call([]reflect.Value{s.rv, argv, reply})
	if errInter := returnvalues[0].Interface(); errInter != nil {
		//return errInter(error)
		return errInter.(error)
	}
	return nil
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
	s.ServerCodec(codec.NewCodecFuncMap[opt.CodecType](conn))
}

// 定义request结构体
type Request struct {
	h            *codec.Header
	argv, replyv reflect.Value
}

// 创建readRequest函数
func (s *Server) readRequest(c codec.Codec) (*Request, error) {
	var h codec.Header
	if err := c.ReadHeader(&h); err != nil {
		log.Println("read header error:", err)
		return nil, err
	}
	req := &Request{h: &h}
	req.argv = reflect.New(reflect.TypeOf(""))               //go
	if err := c.ReadBody(req.argv.Interface()); err != nil { //go
		log.Println("read body error:", err)
		//不应该返回nil,因为req不为nil,有header
		return req, err
	}
	return req, nil
}

// 创建ServerCodec函数
// ServerCodec方法用于处理客户端请求
var invaildRequest = struct{}{}

func (s *Server) ServerCodec(c codec.Codec) {
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
		req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp: %d", req.h.Seq))
		log.Println(req.h, req.argv.Elem())
		go s.sendHandle(c, req.h, req.replyv.Interface(), sending, wg)
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
func (s *Server) sendHandle(c codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	s.sendResponse(c, h, body, sending)
}
