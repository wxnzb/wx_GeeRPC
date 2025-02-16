package geerpc

import (
	"context"
	"log"
	"net"
	"strings"
	"testing"
	"time"
)

// 客户端连接超时
func Test_ClientDialTimeout(t *testing.T) {
	t.Parallel()
	l, _ := net.Listen("tcp", ":0")
	f := func(conn net.Conn, opt *Option) (*Client, error) {
		_ = conn.Close()
		time.Sleep(2 * time.Second)
		return nil, nil
	}
	t.Run("第一次超时", func(t *testing.T) {
		_, err := dialtimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: time.Second})
		_assert(err != nil && strings.Contains(err.Error(), "connect timeout"), "a timeout error")
	})
	t.Run("第二次无限", func(t *testing.T) {
		_, err := dialtimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: 0})
		_assert(err == nil, "nolimited timeout")
	})
}

type Bar int

func (b Bar) Timeout(argv int, reply *int) error {
	time.Sleep(time.Second * 2)
	return nil
}
func startServer(addr chan string) {
	var bar Bar
	err := Register(&bar)
	if err != nil {
		log.Println("register error:", err)
	}
	l, _ := net.Listen("tcp", ":0") //fk.key
	addr <- l.Addr().String()       //fk.key
	log.Println("Server started on ", l.Addr())
	Accept(l) //fk.key
}
func Test_call(t *testing.T) {
	t.Parallel()
	addr := make(chan string)
	go startServer(addr)
	Addr := <-addr
	time.Sleep(time.Second)
	t.Run("客户端超时", func(t *testing.T) {
		cl, _ := Dial("tcp", Addr)
		var reply int
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		err := cl.Call(ctx, "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "client timeout")
	})
	t.Run("客户端超时", func(t *testing.T) {
		cl, _ := Dial("tcp", Addr, &Option{HandleTimeout: time.Second})
		var reply int
		err := cl.Call(context.Background(), "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), "sever timeout"), "sever timeout")
	})
}
