package geerpc

import (
	"fmt"
	"reflect"
	"testing"
)

// 要实现结构体的方法
type Foo int
type Args struct {
	num1, num2 int
}

func (f Foo) Add(args Args, reply *int) error {
	*reply = args.num1 + args.num2
	return nil
}
func (f Foo) add(args Args, reply *int) error {
	*reply = args.num1 + args.num2
	return nil
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := NewService(&foo)
	_assert(len(s.serthods) == 1, "Foo has 1 methods,but %d methods", len(s.serthods))
	MethodFunc := s.serthods["Add"]
	_assert(MethodFunc != nil, "Foo has Add method")
}

func TestDiaoMethodFunc(t *testing.T) {
	var foo Foo
	s := NewService(&foo)
	MethodFunc := s.serthods["Add"]
	Argv := MethodFunc.NewArgv()
	Reply := MethodFunc.NewReply()
	Argv.Set(reflect.ValueOf(Args{1, 2}))
	err := s.Call(MethodFunc, Argv, Reply)
	_assert(err == nil && *Reply.Interface().(*int) == 3, "Add(1,2) should return 3")
}
func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed:"+msg, v...))
	}
}
