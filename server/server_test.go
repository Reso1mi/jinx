package server

import (
	"fmt"
	"jinx/request"
	"jinx/router"
	"testing"
)

type TestRouter struct {
	router.BaseRouter
}

func (router *TestRouter) Handle(request request.Request) {
	data := request.GetReqData()
	// 回显
	if _, err := request.GetConnection().GetTCPConnection().Write(data); err != nil {
		fmt.Println("[Jinx Server] write back err", err)
	}
}

func TestServerV1(t *testing.T) {
	s := NewServer("/home/resolmi/GolandProjects/jinx/config.json")
	// 一个Server绑定一个Router
	s.AddRouter(&TestRouter{})
	s.Serve()
}
