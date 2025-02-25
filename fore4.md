## 1
_=conn.Close()
这个不需要加:=
## 2
strings.Contains(err.Error(), "connect timeout") 代码解析
这个代码的作用是检查错误信息 err.Error() 是否包含 "connect timeout" 这个子字符串。
## 3
- select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}
- <-time.After(opt.ConnectTimeout)这个的作用是在opt.ConnectTimeout后发送一个信号