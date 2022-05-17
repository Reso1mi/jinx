package jinx

import (
	"github.com/imlgw/jinx/errors"
	"github.com/imlgw/jinx/internal"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"sync/atomic"
)

type eventloop struct {
	idx int // idx

	epoll *internal.Epoll

	reactor map[int]Reactor // fd 对应的 Reactor

	conncnt uint64

	ser *server
}

// NewLoop 创建一个事件循环，idx 为循环序号
func newLoop(idx int) (*eventloop, error) {
	epoll, err := internal.CreateEpoll()
	if err != nil {
		return nil, err
	}

	return &eventloop{
		epoll:   epoll,
		idx:     idx,
		conncnt: 0,
		reactor: make(map[int]Reactor),
	}, nil
}

// Loop 开始事件循环
func (loop *eventloop) poll() error {
	if err := loop.epoll.Polling(
		func(fd int, eventType internal.EventType) error {
			r, ok := loop.reactor[fd]
			if ok {
				if err := r.handleEvent(fd, eventType); err != nil {
					return err
				}
			} else {
				// 可能是不再使用的fd，netty 中的处理是直接在 epoll 事件上删除了这个 fd
				if err := loop.epoll.Delete(fd); err != nil {
					log.Printf("remove unused fd err, %v \n", err)
				}
			}
			return nil
		}); err != nil {
		return err
	}
	return nil
}

// read eventloop 可读事件处理，将内核中的数据写入 inBuffer
func (loop *eventloop) handleReadEvent(c *connection) error {
	// TODO: 动态调整 inBuffer 大小 （RingBuffer?）
	n, err := unix.Read(c.fd, c.inBuffer)
	if err != nil || n == 0 {
		if err == unix.EAGAIN {
			// https://stackoverflow.com/questions/14370489/what-can-cause-a-resource-temporarily-unavailable-on-sock-send-command
			return nil
		}
		log.Printf("handleReadEvent err, %v \n", err)
		return c.Close()
	}
	return nil
}

// write eventloop 可写事件处理，将 outBuffer 中的数据写入内核（flush）
func (loop *eventloop) handleWriteEvent(c *connection) error {
	if len(c.outBuffer) != 0 {
		// 当内核缓冲区满的时候可能无法完全写入，writen < len(c.out)
		writen, err := unix.Write(c.fd, c.outBuffer)
		if err != nil {
			return c.Close()
		}
		if writen == len(c.outBuffer) {
			c.outBuffer = nil
		} else {
			// 剩余 c.out[writen:]，等待下次可写事件触发再 flush 到内核
			c.outBuffer = c.outBuffer[writen:]
		}
	}

	// c.out 中的数据已经全部写入内核，暂时不再需要监听写事件，当用户通过 conn 写入的时候再开启 write 事件
	if len(c.outBuffer) == 0 {
		if err := c.loop.epoll.ModRead(c.fd); err != nil {
			return err
		}
	}
	return nil
}

// handleAccept accept 事件处理
func (loop *eventloop) handleAccept(fd int) error {
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

	addr := sockaddrToTCPOrUnixAddr(sa)
	nextLoop := loop.ser.loopGroup.Next(addr)

	// 将 connfd 的读写事件注册到 epoll 的 evnet_list
	if err := nextLoop.epoll.RegReadWrite(connfd); err != nil {
		log.Printf("reg connfd event rw error, %v \n", err)
		_ = unix.Close(connfd)
		return err
	}

	conn := newConnection(connfd, sa, addr, nextLoop)
	// 将 conn 绑定到该 loop 对应 fd 的回调上
	nextLoop.reactor[connfd] = conn
	atomic.AddUint64(&nextLoop.conncnt, 1)
	return nil
}

func sockaddrToTCPOrUnixAddr(sa unix.Sockaddr) net.Addr {
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

// Close 停止事件循环
func (loop *eventloop) Close() error {
	if err := loop.epoll.Close(); err != nil {
		return err
	}
	return nil
}
