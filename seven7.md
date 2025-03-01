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