package request

import "jinx/contract/connection"

type IRequest interface {
	// GetConnection 获取当前请求的连接
	GetConnection() connection.IConnection

	// GetReqData 获取请求数据
	GetReqData() []byte
}
