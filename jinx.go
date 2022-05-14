package jinx

import (
	"fmt"
	"github.com/imlgw/jinx/codec"
	"github.com/imlgw/jinx/config"
	"log"
	"net"
	"runtime"
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

func Run(network, addr string, opts ...Option) error {
	options := LoadOptions(opts...)
	if options.Codec == nil {
		options.Codec = codec.NewDefaultLengthFieldCodec()
	}

	// 初始化 loopGroup
	loopGroup = NewEventGroup(options.Lb)

	// 创建并启动 acceptor
	acceptor, err := NewAcceptor(network, addr)
	if err != nil {
		return err
	}
	go acceptor.Run()

	loopNum := options.LoopNum
	if loopNum <= 0 {
		// 不设置默认是 cpu 个数
		loopNum = runtime.NumCPU()
	}

	// 创建并启动 loopNum 个事件循环
	for i := 0; i < loopNum; i++ {
		loop, err := NewLoop(i)
		if err != nil {
			return err
		}
		loopGroup.Register(loop)
		go func() {
			err := loop.Loop()
			if err != nil {
				log.Printf("create and run loop error, %v \n", err)
			}
		}()
	}
	return nil
}
