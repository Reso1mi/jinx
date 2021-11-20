package jnet

import "jinx/contract/request"

type BaseRouter struct {
}

func (b *BaseRouter) BeforeHandle(request request.IRequest) {
}

func (b *BaseRouter) Handle(request request.IRequest) {
}

func (b *BaseRouter) AfterHandle(request request.IRequest) {
}
