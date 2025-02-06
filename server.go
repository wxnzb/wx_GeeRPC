// 创建Option结构体
// 创建默认Option结构体
package geerap

import (
	"My_Geerpc/codec"
	"encoding/json"
	"fmt"
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

// 创见server结构体
// 创建一个函数，来处理客户端连接
// 创建默认server结构体,来处理默认的连接
type Server struct{}

func NewServer() *Server {
	return &Server{}
}
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
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
		log.Println("invalid magic number")
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
	var h *codec.Header
	if err := c.ReadHeader(h); err != nil {
		log.Println("read header error:", err)
		return nil, err
	}
	req := Request{h: h}
	req.argv = reflect.New(reflect.TypeOf(""))               //go
	if err := c.ReadBody(req.argv.Interface()); err != nil { //go
		log.Println("read body error:", err)
		//不应该返回nil,因为req不为nil,有header
		return nil, err
	}
	return &req, nil
}

// 创建ServerCodec函数
// ServerCodec方法用于处理客户端请求
func (s *Server) ServerCodec(c codec.Codec) {
	sending := new(sync.Mutex)
	defer c.Close()
	for {
		req, err := s.readRequest(c)
		if err != nil {
			if req == nil {
				break
			} else { //对应上面的不应该
				req.h.Error = err.Error()
				s.sendResponse(c, req.h, nil, sending)
				continue
			}
		}
		//之前写成了c.h.Seq,你要始终记住此时的codec.Codec是*Gobcodec
		req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp: %d", req.h.Seq))
		s.sendResponse(c, req.h, req.replyv.Interface(), sending)
	}
}

// 创建sendResponse函数
func (s *Server) sendResponse(c codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := c.Write(h, body); err != nil {
		log.Println("write response error:", err)
	}
}
