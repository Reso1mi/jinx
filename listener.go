package jinx

import (
	"github.com/imlgw/jinx/errors"
	"github.com/imlgw/jinx/internal"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"sync/atomic"
)

type listener struct {
	addr net.Addr
	lnfd int
	loop *eventloop
}

func NewListener(network, addr string) (Reactor, error) {
	// 生成一个 Listener（主要是拿 listenerfd 加入 eventloop）
	// listen, err := net.Listen(network, addr)
	// 这里不使用 net.Listen，这个会将 fd 直接加入 netpoll 的 eventloop，不确定会不会有其他影响
	socketfd, err := internal.SocketListen(network, addr)
	if err != nil {
		return nil, err
	}
	mainLoop, err := NewLoop(-1)
	if err != nil {
		return nil, err
	}

	// 将 socketfd 加入 mainLoop 的 epoll 事件中监听可读事件。
	// 发生读事件说明有连接进入（监听套接字的可读事件就是tcp全连接队列非空）
	// https://zhuanlan.zhihu.com/p/399651675
	if err := mainLoop.epoll.RegRead(socketfd); err != nil {
		return nil, err
	}
	l := &listener{lnfd: socketfd, loop: mainLoop}
	// 绑定到 loop 的响应器上
	mainLoop.reactor[socketfd] = l
	atomic.AddUint64(&mainLoop.conncnt, 1)
	return l, nil
}

func (l *listener) Run() {
	if err := l.loop.Loop(); err != nil {
		log.Printf("loop error, %v \n", err)
		return
	}
}

func (l *listener) HandleEvent(fd int, _ internal.EventType) error {
	// 建立新链接
	connfd, sa, err := unix.Accept(fd)
	if err != nil {
		return errors.ErrAcceptSocket
	}

	// 将 connfd 设置为非阻塞模式
	if err := unix.SetNonblock(connfd, true); err != nil {
		_ = unix.Close(connfd)
		log.Printf("set nfd nonblock error, %v \n", err)
		return err
	}

	addr := SockaddrToTCPOrUnixAddr(sa)
	nextLoop := loopGroup.Next(addr)

	// 将 connfd 的读写事件注册到 epoll 的 evnet_list
	if err := nextLoop.epoll.RegReadWrite(connfd); err != nil {
		log.Printf("reg connfd event rw error, %v \n", err)
		_ = unix.Close(connfd)
		return err
	}

	conn := NewConnection(connfd, sa, addr, nextLoop)
	// 将 conn 绑定到该 loop 对应 fd 的回调上
	nextLoop.reactor[connfd] = conn
	atomic.AddUint64(&nextLoop.conncnt, 1)
	return nil
}

func SockaddrToTCPOrUnixAddr(sa unix.Sockaddr) net.Addr {
	switch sa := (sa).(type) {
	case *unix.SockaddrInet4:
		return &net.TCPAddr{IP: sa.Addr[:], Port: sa.Port}
		// TODO: IPV6
	case *unix.SockaddrUnix:
		return &net.UnixAddr{Name: sa.Name, Net: "unix"}
	default:
		log.Printf("(unknown - %T)", sa)
		return nil
	}
}
