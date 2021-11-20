package router

import "jinx/contract/request"

type IRouter interface {
	// BeforeHandle 处理业务前的方法
	BeforeHandle(request request.IRequest)
	// Handle 具体执行的方法
	Handle(request request.IRequest)
	// AfterHandle 处理业务后的方法
	AfterHandle(request request.IRequest)
}
