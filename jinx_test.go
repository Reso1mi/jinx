package jinx

import (
	"fmt"
	"testing"
)

type TestRouter struct {
	BaseRouter
}

func (router *TestRouter) Handle(request Request) {
	// 原始数据
	data := request.GetReqData()
	fmt.Println("server receiver:", string(data))
	// 回显
	if err := request.GetConnection().Send(data); err != nil {
		fmt.Println("[Jinx Server] write back err", err)
	}
}

func TestJinxServer(t *testing.T) {

	s := NewServer("/home/resolmi/GolandProjects/jinx/config.json",
		WithRouter(&TestRouter{}),
	)
	s.Serve()
}
