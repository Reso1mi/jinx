package server

import (
	"fmt"
	"github.com/imlgw/jinx/request"
	"github.com/imlgw/jinx/router"
	"testing"
)

type TestRouter struct {
	router.BaseRouter
}

func (router *TestRouter) Handle(request request.Request) {
	// 原始数据
	data := request.GetReqData()
	fmt.Println("server receiver:", string(data))
	// 回显
	if err := request.GetConnection().Send(data); err != nil {
		fmt.Println("[Jinx Server] write back err", err)
	}
}

func TestServer(t *testing.T) {

	s := NewServer("/home/resolmi/GolandProjects/jinx/config.json")

	// 一个Server绑定一个Router
	s.AddRouter(&TestRouter{})
	s.Serve()
}
