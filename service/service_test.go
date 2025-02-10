//要实现结构体的方法
type Foo int
type Args struct{
	num1,num2 int
}
func (f Foo)Add(args Args,reply *int)error{
    *reply=args.num1+args.num2
	return nil
}
func (f Foo)add(args Args,reply *int)error{
    *reply=args.num1+args.num2
	return nil
}
TestNewServer(){
	var foo Foo
	s:=NewServer(foo)
	MethodFunc:=s.serthods["Add"]
}
TestDiaoMethodFunc(){
	var foo Foo
	s:=NewServer(foo)
	MethodFunc:=s.serthods["Add"]
	Argv:=MethodFunc.NewArgv()
	Reply:=MethodFunc.NewReply()
	Argv.Set(reflect.ValueOf(Args{1,2}))
    s.Call(MethodFunc,Argv,Reply)
}