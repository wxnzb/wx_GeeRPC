## 1
- if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() 这里可不可以写成if mType.Out(0) != reflect.TypeOf((error))
- 原因：不可以直接写成 reflect.TypeOf((error))，因为 error 是一个 接口类型（interface），他不是一个具体的变量，直接 reflect.TypeOf(error) 会得到 nil，这不是我们想要的结果。
## 2
if ast.IsExported(s.name)==false{
    log.Fatalf("rpc server: %s is not a valid name",s.name)
	}
-   //这个是必须的，他必须是Wx.nzb导出的才行
	//Go 的 net/rpc 规范 中，小写的结构体不能注册方法，因为小写的结构体字段在包外是不可见的，无法被 RPC 调用。
## 3
for i := 0; i < s.rt.NumMethod(); i++ {
		method := s.rt.Method(i)
		// if ast.IsExported(method.Name) == false {
		// 	continue
		// }
}
- 省略//，method.Name的大小写不影响方法的注册，因为 reflect.Type.NumMethod()只会返回可导出可大写开头的方法
## 4
func isExportedCan(t reflect.Type)bool{
	return ast.IsExported(t.Name()) && t.PkgPath() != ""
}
- t.PkgPath() == "" 只会出现在两种情况下：
- 内置类型（int, string, bool 等）
- 导出的类型（大写开头的 struct）