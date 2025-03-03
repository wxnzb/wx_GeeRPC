## 1
- t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}这样调用和下面有区别吗 var err error
	err=SendHeartBeat(register,addr)
	for err==nil{
		<-time.After(duration)
		err=SendHeartBeat(register,addr)
	}
- 性能问题
## 2
- if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
- strings.TrimSpace(server) != "" 去掉空格
## 3
- const defaultTimeout = time.Second * time.Duration(10)和const defaultTimeout = time.Second * 10有什么区别吗
- 表达的效果其实是一样的
## 4
- 关于http的一些小知识
- l:=net.Listen("tcp",addr)
- http.Handle(defaultRegistryPath, r)//注册http服务器
- http.Serve(l,nil)//启动http服务器，他监听l端口，有的话就自动调用serveHTTP函数
- func (r *Registry) HandleHTTP() {
    http.Handle("/geerpc/registry", http.HandlerFunc(r.ServeHTTP))
}
HandleHTTP() 调用 http.Handle() 注册了 /geerpc/registry 路由，并指定 r.ServeHTTP 处理它。
http.Serve(l, nil) 开始监听请求。
当客户端访问 http://localhost:8080/geerpc/registry 时：
http.Serve 监听到请求，并找到 /geerpc/registry 这个路由。
自动调用 r.ServeHTTP(w, req) 来处理这个请求。
