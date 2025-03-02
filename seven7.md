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