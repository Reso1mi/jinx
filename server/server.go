package server

import (
	"fmt"
	"github.com/imlgw/jinx/codec"
	"github.com/imlgw/jinx/config"
	"github.com/imlgw/jinx/conn"
	"github.com/imlgw/jinx/router"
	"net"
)

type Server interface {

	// Start 启动服务器
	Start()

	// Stop 停止服务器
	Stop()

	// Serve 开始服务
	Serve()

	// AddRouter 给当前服务添加路由处理业务
	AddRouter(router router.Router)
}

type server struct {
	name      string
	ipVersion string
	ip        string
	port      int
	router    router.Router
	codec     codec.ICodec
}

func (s *server) AddRouter(router router.Router) {
	s.router = router
}

func (s *server) Start() {
	fmt.Printf("[Config] ServerName: %s, IP: %s, Port: %d, IPVersion: %s, MaxConn: %d\n",
		config.ServerConfig.Name, config.ServerConfig.Host, config.ServerConfig.Port,
		config.ServerConfig.IPVersion, config.ServerConfig.MaxConn)
	fmt.Printf("[Jinx Start] Server Listener at IP: %s, Port: %d\n", s.ip, s.port)
	var connID uint = 0
	go func() {
		// 1.获取tcp的Addr
		addr, err := net.ResolveTCPAddr(s.ipVersion, fmt.Sprintf("%s:%d", s.ip, s.port))
		if err != nil {
			fmt.Println("[Jinx Server] resolve tcp addr errors:", err)
		}
		// 2.监听服务器的地址
		tcpListener, err := net.ListenTCP(s.ipVersion, addr)
		if err != nil {
			fmt.Println("[Jinx Server] listen err ", err)
		}
		// 3.等待客户端链接
		for {
			tcpConn, err := tcpListener.AcceptTCP()
			if err != nil {
				fmt.Println("[Jinx Server] Accept err:", err)
				continue
			}
			connection := conn.NewConnection(tcpConn, connID, s.router, s.codec)
			connection.Start()
			connID++
		}
	}()
}

func (s *server) Stop() {

}

func (s *server) Serve() {
	s.Start()

	// 阻塞
	select {}
}

func NewServer(path string) Server {
	if err := config.InitConfig(path); err != nil {
		panic(err)
	}
	s := &server{
		name:      config.ServerConfig.Name,
		ipVersion: config.ServerConfig.IPVersion,
		ip:        config.ServerConfig.Host,
		port:      config.ServerConfig.Port,
		router:    nil,
		codec:     codec.NewLengthFieldCodec(),
	}
	return s
}
