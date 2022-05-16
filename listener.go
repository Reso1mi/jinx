package jinx

import (
	"github.com/imlgw/jinx/internal"
	"log"
	"net"
	"sync/atomic"
)

type listener struct {
	addr net.Addr
	lnfd int
	loop *eventloop
}

func newListener(network, addr string) (*listener, error) {
	// 生成一个 Listener（主要是拿 listenerfd 加入 eventloop）
	// listen, err := net.Listen(network, addr)
	// 这里不使用 net.Listen，这个会将 fd 直接加入 netpoll 的 eventloop，不确定会不会有其他影响
	socketfd, err := internal.SocketListen(network, addr)
	if err != nil {
		return nil, err
	}
	mainLoop, err := newLoop(-1)
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

func (l *listener) run() {
	if err := l.loop.poll(); err != nil {
		log.Printf("loop error, %v \n", err)
		return
	}
}

func (l *listener) handleEvent(fd int, _ internal.EventType) error {
	return l.loop.handleAccept(fd)
}
