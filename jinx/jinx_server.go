package jinx

type Server interface {

	// Start 启动服务器
	Start()

	Stop()

	Serve()
}
