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
## 5
_assert(err==nil&&*Reply.Interface().(*int)==3,"Add(1,2) should return 3")
- 不能写成*Reply==3，Reply是reflect.Value类型的
- reflect.Value.Interface()方法的作用是将reflect.Value还原成普通的interface{}类型值，以便后续进行类型断言
## 6
func TestNewServer(t *testing.T){
	var foo Foo
	s:=NewServer(foo)
	_assert(len(s.serthods)==1,"Foo has 1 methods")
	MethodFunc:=s.serthods["Add"]
	_assert(MethodFunc!=nil,"Foo has Add method")
}
- 函数里面的参数有什么用，我感觉都没用上，可以删除不写吗
- 参数 t *testing.T 不能删除，它们是 Go 语言 testing 框架的标准写法,运行go test
##  7
    var reply reflect.Value //返回一定是个指针所以不用判断
	reply = reflect.New(m.Reply.Elem())
- 在Go语言中，变量声明和赋值应该在同一行进行，否则会违反代码风格指南
## 8
svi := sv.(*Service)
- sv是interface{}类型，不能直接用*Service类型转换，需要先断言成*Service类型，然后再用*Service类型转换成*Service类型，.不能省略
## 9
argvi := req.argv.Interface()
req.argv 是一个 reflect.Value 类型，它表示某个变量的值。
req.argv.Interface() 将 reflect.Value 转换成 interface{}，即获取其实际的 Go 值。