package jnet

import (
	"fmt"
	"jinx/config"
	"jinx/contract/router"
	"jinx/contract/server"
	"net"
)

type Server struct {
	Name      string
	IPVersion string
	IP        string
	Port      int
	Router    router.IRouter
}

func (s *Server) AddRouter(router router.IRouter) {
	s.Router = router
}

func (s *Server) Start() {
	fmt.Printf("[Config] ServerName: %s, IP: %s, Port: %d, IPVersion: %s, MaxConn: %d, MaxPackSize: %d byte\n",
		config.ServerConfig.Name, config.ServerConfig.Host, config.ServerConfig.Port,
		config.ServerConfig.IPVersion, config.ServerConfig.MaxConn, config.ServerConfig.MaxPackSize)
	fmt.Printf("[Jinx Start] Server Listener at IP: %s, Port: %d\n", s.IP, s.Port)
	var connID uint = 0
	go func() {
		// 1.获取tcp的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("[Jinx Server] resolve tcp addr error:", err)
		}
		// 2.监听服务器的地址
		tcpListener, err := net.ListenTCP(s.IPVersion, addr)
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
			connection := NewConnection(tcpConn, connID, s.Router)
			connection.Start()
			connID++
		}
	}()
}

func (s *Server) Stop() {

}

func (s *Server) Serve() {
	s.Start()

	// 阻塞
	select {}
}

func NewServer(path string) server.IServer {
	if err := config.InitConfig(path); err != nil {
		panic(err)
	}
	s := &Server{
		Name:      config.ServerConfig.Name,
		IPVersion: config.ServerConfig.IPVersion,
		IP:        config.ServerConfig.Host,
		Port:      config.ServerConfig.Port,
		Router:    nil,
	}
	return s
}
