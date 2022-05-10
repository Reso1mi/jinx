package jinx

import (
	"github.com/imlgw/jinx/errors"
	"github.com/imlgw/jinx/internal"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"sync/atomic"
)

type Reactor interface {
	HandleEvent(fd int, eventType internal.EventType) error
}

type Acceptor struct {
	loop *eventloop
}

func (a *Acceptor) HandleEvent(fd int, _ internal.EventType) error {
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

	handler := newHandler(connfd, sa, addr, nextLoop)
	// 将 handler 绑定到该 loop 对应 fd 的回调上
	nextLoop.reactor[connfd] = handler
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

type Handler struct {
	fd         int
	sa         unix.Sockaddr
	loop       *eventloop
	remoteAddr net.Addr
	localAddr  net.Addr
}

func newHandler(fd int, sa unix.Sockaddr, remoteAddr net.Addr, loop *eventloop) *Handler {
	return &Handler{
		fd:         fd,
		sa:         sa,
		remoteAddr: remoteAddr,
		loop:       loop,
	}
}

func (h *Handler) HandleEvent(fd int, eventType internal.EventType) error {
	// TODO implement me
	panic("implement me")
}
