package geerpc

import (
	"go/ast"
	"log"
	"reflect"
)

type methodType struct {
	Method reflect.Method
	Argv   reflect.Type
	Reply  reflect.Type
}
type Service struct {
	name string
	//要rv,rt有啥用阿
	rv       reflect.Value
	rt       reflect.Type
	serthods map[string]*methodType
}

func (m *methodType) NewArgv() reflect.Value {
	var argv reflect.Value
	if m.Argv.Kind() == reflect.Ptr {
		argv = reflect.New(m.Argv.Elem())
	} else {
		argv = reflect.New(m.Argv).Elem()
	}
	return argv
}
func (m *methodType) NewReply() reflect.Value {
	//返回一定是个指针所以不用判断
	reply := reflect.New(m.Reply.Elem())
	//这里为啥还要继续进行判断，感觉用处不大，那要是reply是个结构体里面包含了Map或者slice也判断不出来？？
	switch m.Reply.Elem().Kind() {
	case reflect.Map:
		reply.Elem().Set(reflect.MakeMap(m.Reply.Elem()))
	case reflect.Slice:
		reply.Elem().Set(reflect.MakeSlice(m.Reply.Elem(), 0, 0))
	}
	return reply
}

// sevicemethod eg:wx.nzb,wx.jy...就是关于wx的方法func(w *wx)nzb(s string,i int)erroe{}
func NewService(smh interface{}) *Service {
	s := new(Service)
	//s.name=smh.(string)不能这样,为啥
	s.name = reflect.Indirect(reflect.ValueOf(smh)).Type().Name()
	//这个是必须的，他必须是Wx.nzb导出的才行
	//Go 的 net/rpc 规范 中，小写的结构体不能注册方法，因为小写的结构体字段在包外是不可见的，无法被 RPC 调用。
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid name", s.name)
	}
	s.rv = reflect.ValueOf(smh)
	s.rt = reflect.TypeOf(smh)
	NewMethods(s)
	return s
}

func NewMethods(s *Service) {
	s.serthods = make(map[string]*methodType)
	for i := 0; i < s.rt.NumMethod(); i++ {
		method := s.rt.Method(i)
		if method.Type.NumIn() != 3 || method.Type.NumOut() != 1 {
			continue
		}
		//方法里面的参数类型不能是自定义的小写，不能导出
		if !isExportedCan(method.Type.In(1)) || !isExportedCan(method.Type.In(2)) {
			continue
		}
		if method.Type.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		s.serthods[method.Name] = &methodType{
			Method: method, //其实我感觉不太需要这个呀
			Argv:   method.Type.In(1),
			Reply:  method.Type.In(2),
		}
		log.Printf("register method:%s.%s", s.name, method.Name)
	}
}
func isExportedCan(t reflect.Type) bool {
	//需要ast.IsExported(t.Name())吗
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}
func (s *Service) Call(m *methodType, argv, reply reflect.Value) error {
	f := m.Method.Func
	returnvalues := f.Call([]reflect.Value{s.rv, argv, reply})
	if errInter := returnvalues[0].Interface(); errInter != nil {
		//return errInter(error)
		return errInter.(error)
	}
	return nil
}
