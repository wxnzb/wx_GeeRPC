* 1
- 接口结构体
func (w *GobCodec) ReadHeader(h *Header) error
Codec接口函数的实现都是结构体指针*Gobcodec,因此Codec 类型变量必须传 &GobCodec{}，var _ Codec = (*GobCodec)(nil)
我自己感觉所有codec.Codec的地方之后都可以用&GobCodec{}代替
因此func(s *Server)sendResponse(c *codec.Codec,h *codec.Header,body interface{},sending *sync.Mutex)是错误的，
应该是func(s *Server)sendResponse(c codec.Codec,h *codec.Header,body interface{},sending *sync.Mutex)
* 2
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
* 3
- defer func(){
		_=sending.Unlock()
	}()
可以将匿名函数简化为 defer sending.Unlock()
* 4
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