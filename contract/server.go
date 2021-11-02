package contract

type IServer interface {

	// Start 启动服务器
	Start()

	Stop()

	Serve()
}
