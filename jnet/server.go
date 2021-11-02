package jnet

import (
	"fmt"
	"jinx/contract"
	"net"
)

type Server struct {
	Name      string
	IPVersion string
	IP        string
	Port      int
}

func (s *Server) Start() {
	fmt.Printf("[Jinx Start] Server Listener at IP: %s, Port: %d\n", s.IP, s.Port)
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
			connect, err := tcpListener.AcceptTCP()
			if err != nil {
				fmt.Println("[Jinx Server] Accept err:", err)
				continue
			}
			go func() {
				for {
					buf := make([]byte, 512)
					cnt, err := connect.Read(buf)
					if err != nil {
						fmt.Println("[Jinx Server] read err", err)
					}
					fmt.Printf("[Jinx Server] read data:%s \n", buf[:cnt])
					// 回显
					if _, err := connect.Write(buf[:cnt]); err != nil {
						fmt.Println("[Jinx Server] write back err", err)
						continue
					}
				}
			}()
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

func NewServer(name string) contract.IServer {
	s := &Server{
		Name:      name,
		IPVersion: "tcp4",
		IP:        "0.0.0.0",
		Port:      8999,
	}
	return s
}
