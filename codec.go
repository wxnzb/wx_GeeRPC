package codec

//定义头结构体
//定义Codec接口读写消息头体
//通过map进行选择编解码用的方法
import "io"

type Header struct {
	ServiceMethod string
	Seq           uint64
	Error         string
}
type Codec interface {
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
	io.Closer
}
type NewCodecFunc func(io.ReadWriteCloser) Codec

var NewCodecFuncMap map[string]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[string]NewCodecFunc)
	NewCodecFuncMap["gob"] = NewGobCodec
}
