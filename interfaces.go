package jinx

import "net"

type Connection interface {
	// Start 启动链接
	Start()
	// Stop 停止链接
	Stop()
	// GetTCPConnection 获取当前链接的socket conn
	GetTCPConnection() *net.TCPConn
	// GetConnID 获取当前链接的ID
	GetConnID() uint
	// Send 发送数据
	Send([]byte) error
}
