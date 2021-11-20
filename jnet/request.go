package jnet

import (
	"jinx/contract/connection"
	"jinx/contract/request"
)

type Request struct {
	conn connection.IConnection
	data []byte
}

func (r *Request) GetConnection() connection.IConnection {
	return r.conn
}

func (r *Request) GetReqData() []byte {
	return r.data
}

func NewRequest(conn connection.IConnection, data []byte) request.IRequest {
	r := &Request{
		conn: conn,
		data: data,
	}
	return r
}
