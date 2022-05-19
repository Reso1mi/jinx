package jinx

// Handler event callback
type Handler interface {

	// OnBoot server 启动
	OnBoot(f func(s *server))

	// OnOpen newConn 连接建立
	OnOpen(f func(c *connection))

	// OnClose closeConn 连接关闭
	OnClose(f func(c *connection))

	// OnRead 可读事件，收到客户端发送的数据
	OnRead(f func(c *connection))

	// OnWrite 可写事件，在服务端发送数据到客户端之前
	OnWrite(f func(c *connection))

	// OnShutdown 服务关闭
	OnShutdown(f func(s *server))
}
