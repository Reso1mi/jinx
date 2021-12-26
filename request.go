package jinx

type Request interface {
	// GetConnection 获取当前请求的连接
	GetConnection() Connection
	// GetReqData 获取请求数据
	GetReqData() []byte
}

type request struct {
	conn Connection
	data []byte
}

func (r *request) GetConnection() Connection {
	return r.conn
}

func (r *request) GetReqData() []byte {
	return r.data
}

func NewRequest(conn Connection, data []byte) Request {
	r := &request{
		conn: conn,
		data: data,
	}
	return r
}
