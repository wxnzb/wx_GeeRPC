package main

import (
	geerpc "My_Geerpc"
	"My_Geerpc/codec"
	"fmt"
	"log"
	"net"
	"time"
)

// 创建startSrever函数
// 创建main函数
func startServer(addr chan string) {
	l, _ := net.Listen("tcp", ":0") //fk.key
	addr <- l.Addr().String()       //fk.key
	log.Println("Server started on ", l.Addr())
	geerpc.Accept(l) //fk.key
}

//	func main() {
//		addr := make(chan string)
//		go startServer(addr)
//		time.Sleep(time.Second)
//		conn, _ := net.Dial("tcp", <-addr) //fk.key
//		//这里是把信息通过流conn使得服务器进行接收
//		_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
//		c := codec.NewGobCodec(conn)
//		for i := 0; i < 5; i++ {
//			h := &codec.Header{
//				ServiceMethod: "wx.nzb",
//				Seq:           uint64(i),
//			}
//			_ = c.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
//			//感觉下面这些不太需要，只是为了检测
//			_ = c.ReadHeader(h)
//			var reply string
//			_ = c.ReadBody(&reply)
//			log.Println("reply:", reply)
//		}
//		conn.Close()
//	}
func main() {
	addr := make(chan string)
	go startServer(addr)
	time.Sleep(time.Second)
	cl := geerpc.Dial("tcp", <-addr)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "wx.nzb",
			Seq:           uint64(i),
		}
		args := fmt.Sprintf("geerpc req %d", h.Seq)
		var reply string
	_:
		cl.Call(h.ServiceMethod, args, &reply)
		log.Println("reply:", reply)
	}
}
