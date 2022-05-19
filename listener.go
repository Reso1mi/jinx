package jinx

import (
	"github.com/imlgw/jinx/internal"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type listener struct {
	once sync.Once
	addr net.Addr
	lnfd int
	loop *eventloop
}

func newListener(network, addr string, ser *server) (*listener, error) {
	// 生成一个 Listener（主要是拿 listenerfd 加入 eventloop）
	// listen, err := net.Listen(network, addr)
	// 这里不使用 net.Listen，这个会将 fd 直接加入 netpoll 的 eventloop，不确定会不会有其他影响
	socketfd, naddr, err := internal.SocketListen(network, addr)
	if err != nil {
		return nil, err
	}
	mainLoop, err := newLoop(-1, ser)
	if err != nil {
		return nil, err
	}

	// 将 socketfd 加入 mainLoop 的 epoll 事件中监听可读事件。
	// 发生读事件说明有连接进入（监听套接字的可读事件就是tcp全连接队列非空）
	// https://zhuanlan.zhihu.com/p/399651675
	if err := mainLoop.epoll.RegRead(socketfd); err != nil {
		return nil, err
	}
	l := &listener{lnfd: socketfd, loop: mainLoop, addr: &net.TCPAddr{IP: naddr.IP, Port: naddr.Port}}
	// 绑定到 loop 的响应器上
	mainLoop.reactor[socketfd] = l
	atomic.AddUint64(&mainLoop.conncnt, 1)
	return l, nil
}

func (l *listener) run() error {
	if err := l.loop.poll(); err != nil {
		log.Printf("loop error, %v \n", err)
		return err
	}
	return nil
}

func (l *listener) Close() error {
	l.once.Do(
		func() {
			// 关闭 listener 同时关闭 mainLoop
			if err := l.loop.Close(); err != nil {
				log.Printf("close error %v \n", err)
				return
			}
			if err := unix.Close(l.lnfd); err != nil {
				log.Printf("close lnfd error %v \n", err)
				return
			}
			l.loop = nil
			l.addr = nil
		})
	return nil
}

func (l *listener) handleEvent(fd int, _ internal.EventType) error {
	return l.loop.handleAccept(fd)
}
