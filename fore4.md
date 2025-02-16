## 1
_=conn.Close()
这个不需要加:=
## 2
strings.Contains(err.Error(), "connect timeout") 代码解析
这个代码的作用是检查错误信息 err.Error() 是否包含 "connect timeout" 这个子字符串。