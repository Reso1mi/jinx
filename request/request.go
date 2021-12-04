package request

import (
	"github.com/imlgw/jinx"
)

type Request interface {
	// GetConnection 获取当前请求的连接
	GetConnection() jinx.Connection
	// GetReqData 获取请求数据
	GetReqData() []byte
}

type request struct {
	conn jinx.Connection
	data []byte
}

func (r *request) GetConnection() jinx.Connection {
	return r.conn
}

func (r *request) GetReqData() []byte {
	return r.data
}

func NewRequest(conn jinx.Connection, data []byte) Request {
	r := &request{
		conn: conn,
		data: data,
	}
	return r
}
