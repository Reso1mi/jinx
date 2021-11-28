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
	// RemoteAddr 获取远程服务端的状态
	RemoteAddr() net.Addr
	// Send 发送数据
	Send([]byte) error
}
