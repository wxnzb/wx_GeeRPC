package geerpc

import (
	"context"
	"log"
	"net"
	"os"
	"runtime"
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

// 是不是有病，这两个测试都过不了，恶心死我了
func Test_call(t *testing.T) {
	t.Parallel()
	addr := make(chan string)
	go startServer(addr)
	Addr := <-addr
	time.Sleep(time.Second) //这里为啥要停顿一秒，上面不是已经阻塞等待了吗
	t.Run("客户端超时", func(t *testing.T) {
		cl, _ := Dial("tcp", Addr)
		var reply int
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		err := cl.Call(ctx, "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "client timeout")
	})
	t.Run("服务器超时", func(t *testing.T) {
		cl, _ := Dial("tcp", Addr, &Option{HandleTimeout: time.Second})
		var reply int
		err := cl.Call(context.Background(), "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), "handle timeout"), "handle timeout")
	})
}

// 最终证明 XDial 不仅可以支持 TCP，还可以用于 Unix Socket 连接，适用于本机的高效 IPC 通信！
// Unix Socket 连接 中，通信不通过网络，而是通过本地的 Socket 文件，addr 就是这个文件的路径
func Test_Xdial(t *testing.T) {
	if runtime.GOOS == "linux" {
		ch := make(chan struct{})
		addr := "/tmp/geerpc.sock"
		go func() {
			//为啥要删除
			//在 Unix 系统中，Unix Socket 文件（/tmp/geerpc.sock）是持久化的，如果服务器进程上次运行崩溃或意外退出，Socket 文件可能还留在系统中。
			//如果不删除，可能会失败
			net.Listen("unix", addr)
			_ = os.Remove(addr)
			l, err := net.Listen("unix", addr)
			if err != nil {
				log.Fatal("listen error:", err)
			}
			ch <- struct{}{}
			Accept(l)
		}()
		<-ch
		_, err := XDial("unix@" + addr)
		_assert(err == nil, "xdial error")
	}
}
