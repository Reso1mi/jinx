package jinx

import (
	"fmt"
	"github.com/imlgw/jinx/codec"
	"github.com/imlgw/jinx/internal"
	"golang.org/x/sys/unix"
	"net"
)

type connection struct {
	fd         int
	sa         unix.Sockaddr
	loop       *eventloop
	remoteAddr net.Addr
	localAddr  net.Addr
	codec      codec.ICodec // 编解码器
	out        []byte
}

func NewConnection(fd int, sa unix.Sockaddr, remoteAddr net.Addr, loop *eventloop) Reactor {
	return &connection{
		fd:         fd,
		sa:         sa,
		remoteAddr: remoteAddr,
		loop:       loop,
	}
}

func (c *connection) Run() {
	panic("no")
}

func (c *connection) HandleEvent(fd int, eventType internal.EventType) error {
	if eventType&unix.EPOLLIN != 0 {
		var buf []byte
		// TODO: 会有并发的问题吗?
		n, err := unix.Read(fd, c.loop.buffer)
		if err == unix.EAGAIN {
			return nil
		}
		buf = c.loop.buffer[:n]
		fmt.Println(buf)
		return nil
	}

	if eventType&unix.EPOLLOUT != 0 {

		if len(c.out) != 0 {
			// 当内核缓冲区满的时候可能无法完全写入，n < len(c.out)
			n, err := unix.Write(fd, c.out)
			if err != nil {
				return c.Close()
			}
			if n == len(c.out) {
				c.out = nil
			} else {
				// 剩余 c.out[n:]
				c.out = c.out[n:]
			}
		}

		// c.out 中的数据已经全部写入内核，暂时不再需要监听写事件
		if len(c.out) == 0 {
			if err := c.loop.epoll.ModRead(c.fd); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *connection) Close() error {
	c.loop = nil
	c.out = nil
	delete(c.loop.reactor, c.fd)
	if err := unix.Close(c.fd); err != nil {
		return err
	}
	return nil
}
