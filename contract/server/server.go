package server

import "jinx/contract/router"

type IServer interface {

	// Start 启动服务器
	Start()

	// Stop 停止服务器
	Stop()

	// Serve 开始服务
	Serve()

	// AddRouter 给当前服务添加路由处理业务
	AddRouter(router router.IRouter)
}
