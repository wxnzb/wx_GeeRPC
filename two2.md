## 1
```go
func Dial(network, address string, opts ...*Option) (cl *Client) {
	//这里先强制使用DefaultOption
	if opts == nil || len(opts) == 1 {
		opts[0] = &DefaultOption
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
	return NewCilent(conn, opts[0])
}
- 这里当opts==nil或len(opts)==0时，opts[0]根本不存在，访问opts[0]会导致 数组越界访问
## 2
```go
func A()bool{
return 1}
- 错误，这里boo类型只能是true或false
## 3
func main() {
	addr := make(chan string)
	go startServer(addr)
	time.Sleep(time.Second)
	cl := geerpc.Dial("tcp", <-addr)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		go func(i int) {
			wg.Add(1)
			h := &codec.Header{
				ServiceMethod: "wx.nzb",
				Seq:           uint64(i),
			}
			args := fmt.Sprintf("geerpc req %d", h.Seq)
			var reply string
		_:
			cl.Call(h.ServiceMethod, args, &reply)
			log.Println("reply:", reply)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
-有问题，应该把wg.Add(1)放在go前面
-如果你在goroutine开始执行之后调用wg.Add(1)，那么wg.Done()的调用可能会在wg.Add(1)之前，从而导致wg.Wait()无法正确等待所有goroutine完成。
## 4
for i := 0; i < 5; i++ {
	wg.Add(1)
	go func() { // ❌ 没有显式传递 i，直接用外部的 i
		defer wg.Done()
		fmt.Println("i:", i)
	}(i)
}
由于 go 语句创建的 goroutine 可能会延迟执行，此时 i 的值可能