package jinx

import (
	"fmt"
	"github.com/imlgw/jinx/codec"
	"github.com/imlgw/jinx/config"
	"net"
)

type Server interface {

	// Start 启动服务器
	Start()

	// Stop 停止服务器
	Stop()

	// Serve 开始服务
	Serve()
}

type server struct {
	name      string
	ipVersion string
	ip        string
	port      int
	opts      *Options
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
			connection := NewConnection(tcpConn, connID, s.opts.Router, s.opts.Codec)
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

func NewServer(path string, opts ...Option) Server {
	options := LoadOptions(opts...)
	if err := config.InitConfig(path); err != nil {
		panic(err)
	}
	if options.Codec == nil {
		options.Codec = codec.NewDefaultLengthFieldCodec()
	}
	s := &server{
		name:      config.ServerConfig.Name,
		ipVersion: config.ServerConfig.IPVersion,
		ip:        config.ServerConfig.Host,
		port:      config.ServerConfig.Port,
		opts:      options,
	}
	return s
}
