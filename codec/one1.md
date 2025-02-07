## 1
- 接口结构体
 `func (w *GobCodec) ReadHeader(h *Header) error `
Codec接口函数的实现都是结构体指针`*Gobcodec`,因此Codec 类型变量必须传 `&GobCodec{}`，`var _ Codec = (*GobCodec)(nil)`
我自己感觉所有`codec.Codec`的地方之后都可以用`&GobCodec{}`代替
因此`func(s *Server)sendResponse(c *codec.Codec,h *codec.Header,body interface{},sending *sync.Mutex)`是错误的，
应该是`func(s *Server)sendResponse(c codec.Codec,h *codec.Header,body interface{},sending *sync.Mutex)`
## 2
```go
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn), //会将conn里面的数据解码到结构体里面
		enc:  gob.NewEncoder(buf),  //会将结构体编码到buf里面
	}
}
- gob.NewEncoder(buf)先把数据写入 buf，等到buf.Flush()才真正写入conn
## 3
-log不同用法
- log.Println("decode error:", err)
- log.Printf("invalid magic number %x", opt.MagicNumber)
- 为什么 log.Println(req.h, req.argv.Elem()) 直接打印到了终端？
- 在 Go 语言中，log.Println()默认情况下会把日志直接输出到标准错误流（stderr），而stderr默认显示在终端，这就是你在终端上看到日志的原因


## 3
- func(s *Server)sendResponse(c *codec.Codec,h *codec.Header,body interface{},sending *sync.Mutex){
	defer func(){
		_=sending.Unlock()
	}()
	sending.Lock()
	if err:=c.Write(h,body);err!=nil{
		log.Println("write response error:",err)
	}
}
错误原因：defer 语句本身的执行时机很重要，不能在 Lock 之前注册 Unlock，否则会引发 死锁或 panic
## 3
```go
- defer func(){
		_=sending.Unlock()
	}()
可以将匿名函数简化为 defer sending.Unlock()
## 4
```go
- func (s *Server)ServerCodec(c codec.Codec){
	sending:=new(sync.Mutex)
	defer c.Close()
	for{
        if req,err:=s.readRequest(c);err!=nil{
			if req==nil{
				break
			} else{//对应上面的不应该
                  req.h.Error=err.Error()
				  s.sendResponse(c,req.h,nil,sending)
				  continue
			}
		}
		//之前写成了c.h.Seq
		req.replyv=reflect.ValueOf(fmt.Sprintf("geerpc resp: %d",req.h.Seq))
		s.sendResponse(c,req.h,req.replyv.Interface(),sending)
	}
}
为啥这样下面的req就成了未定义的了
## 5
- var invalidRequest = struct{}{} 这行代码是 Go 语言中的一种声明和初始化空结构体的方式。
## 遇到的恶心bug
- func (s *Server) readRequest(c codec.Codec) (*Request, error) {
	var h *codec.Header
	if err := c.ReadHeader(h); err != nil {
		log.Println("read header error:", err)
		return nil, err
	}
	req := &Request{h: h}
	req.argv = reflect.New(reflect.TypeOf(""))               //go
	if err := c.ReadBody(req.argv.Interface()); err != nil { //go
		log.Println("read body error:", err)
		//不应该返回nil,因为req不为nil,有header
		return req, err
	}
	return req, nil
}
- h只是一个nil指针，没有分配内存，导致c.ReadHeader(h)在写入时发生崩溃
- 法1 不用指针，&，我写的就是这种
- 法2：h := &codec.Header{} 为h分配内存，这就是go给结构体分配内存的方法
## 6
-   go startServer(addr)
    time.Sleep(time.Second)作用？？
	conn, _ := net.Dial("tcp", <-addr) //fk.key
- time.Sleep(time.Second)让客户端等待1秒，确保服务器完全启动后，再去net.Dial("tcp", <-addr)
