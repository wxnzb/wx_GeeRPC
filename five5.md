## 1
- xc.mu.Lock()和xc.mu.Rclock()有什么区别吗
- 第一种在并发的情况下，无论是读写，都得是原子性，但是第二种情况，读的时候可以多个线程，写的时候是原子性
- 你只要在Xclient这个结构体中定义这个锁为mu sync.RWMutex,这两种都可以用
## 2
- 问题，在写xclient文件夹下的xclient.go文件时，要用到geerpc包下的内容，(eg:server.go中的Option和client.go中的Client和XDial函数)，及的
## 3
- 这里的 http.Handle(defaultRpcPath, s) 让 Go 内置的 HTTP 服务器将 /_geerpc_ 这个路径的请求交给 s（即 Server 类型的对象）处理。
Server 结构体实现了 http.Handler 接口，因此 ServeHTTP() 方法会被调用，所以客户端发送http的CONNECT在defaultRpcPath通道请求后会直接调用ServeHTTP()函数进行交互
## 4
- 当参数为 0 时，表示不使用任何默认的日志格式信息（如时间、文件名等），只打印日志内容本身