// 定义Godcodec结构体
// 使GodcodecGodcodec指针实现codec.go中的Codec接口
// 实现具体的NewGobFunc函数
// 用GodcodecGodcodec指针实现codec.go中的Codec接口里面的方法的具体函数
package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

var _ Codec = (*GobCodec)(nil) //go

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn), //会将conn里面的数据解码到结构体里面
		enc:  gob.NewEncoder(buf),  //会将结构体编码到buf里面
	}
}
func (w *GobCodec) ReadHeader(h *Header) error {
	return w.dec.Decode(h)
}
func (w *GobCodec) ReadBody(i interface{}) error {
	return w.dec.Decode(i)
}
func (w *GobCodec) Write(h *Header, i interface{}) (err error) {
	defer func() {
		_ = w.buf.Flush()
		if err != nil {
			_ = w.conn.Close()
		}
	}()
	if err := w.enc.Encode(h); err != nil {
		log.Println("encode header error:", err)
		return err
	}
	if err := w.enc.Encode(i); err != nil {
		log.Println("encode body error:", err)
		return err
	}
	return nil
}
func (w *GobCodec) Close() error {
	return w.conn.Close()
}
