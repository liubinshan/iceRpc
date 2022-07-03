package codec

import "io"

type Header struct {
	// 服务名和方法名
	ServiceMethod string // format "Service.Method"
	// 请求的序号，用来区分不同的请求
	Seq uint64 // sequence number chosen by client
	// 错误信息
	Error string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
