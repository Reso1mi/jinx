package jinx

import (
	"fmt"
	"github.com/imlgw/jinx/codec"
	"github.com/imlgw/jinx/config"
	"golang.org/x/sys/unix"
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

func Run(network, addr string, opts ...Option) error {
	options := LoadOptions(opts...)
	if options.Codec == nil {
		options.Codec = codec.NewDefaultLengthFieldCodec()
	}

	// 生成一个 Listener（主要是拿 listenerfd 加入 eventloop）
	// listen, err := net.Listen(network, addr)
	// 这里不使用net.Listen，会将 fd 直接加入 gonetpoll 的 eventloop，不确定会不会有其他影响

	// 创建一个 socketfd，暂时只支持 tcp4
	// https://man7.org/linux/man-pages/man2/socket.2.html
	socketfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_TCP)
	if err != nil {
		return err
	}

	// 解析地址
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return err
	}

	// 转换为 unix.SockaddrInet4 用于 bind()
	inet4 := &unix.SockaddrInet4{Port: tcpAddr.Port}
	copy(inet4.Addr[:], tcpAddr.IP)

	// 绑定 socketfd 和地址，https://man7.org/linux/man-pages/man2/bind.2.html
	if err := unix.Bind(socketfd, inet4); err != nil {
		return err
	}

	mainLoop, err := NewLoop(-1)
	if err != nil {
		return err
	}

	//
	loopGroup = NewEventGroup(options.Lb)
	// 将 socketfd 加入 mainLoop 的 epoll 事件中监听可读事件。
	// 发生读事件说明有连接进入（监听套接字的可读事件就是tcp全连接队列非空）
	// https://zhuanlan.zhihu.com/p/399651675
	if err := mainLoop.epoll.RegRead(socketfd); err != nil {
		return err
	}

	return nil
}
