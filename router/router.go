package router

import (
	"jinx/request"
)

type Router interface {
	// BeforeHandle 处理业务前的方法
	BeforeHandle(request request.Request)
	// Handle 具体执行的方法
	Handle(request request.Request)
	// AfterHandle 处理业务后的方法
	AfterHandle(request request.Request)
}

type BaseRouter struct {
}

func (b *BaseRouter) BeforeHandle(request request.Request) {
}

func (b *BaseRouter) Handle(request request.Request) {
}

func (b *BaseRouter) AfterHandle(request request.Request) {
}
