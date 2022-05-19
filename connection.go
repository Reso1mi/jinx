package jinx

import (
	"github.com/imlgw/jinx/codec"
	"github.com/imlgw/jinx/errors"
	"github.com/imlgw/jinx/internal"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"sync"
	"time"
)

type Conn interface {
	net.Conn
	IsOpen() bool
}

type connection struct {
	mux        sync.Mutex
	fd         int
	sa         unix.Sockaddr
	loop       *eventloop
	remoteAddr net.Addr
	localAddr  net.Addr
	codec      codec.ICodec // 编解码器
	outBuffer  []byte       // 写缓存
	inBuffer   []byte       // 读缓存
	closed     bool
}

func newConnection(fd int, sa unix.Sockaddr, remoteAddr net.Addr, loop *eventloop) *connection {
	return &connection{
		fd:         fd,
		sa:         sa,
		remoteAddr: remoteAddr,
		loop:       loop,
		inBuffer:   make([]byte, 0xffff),
	}
}

// Read from client，将 inBuffer 或者内核中的数据写入 b
func (c *connection) Read(b []byte) (int, error) {
	if c.closed {
		return 0, errors.ErrConnClosed
	}
	return copy(b, c.inBuffer), nil
}

// Write b to client，将 b 中的数据写入 outBuffer 或者内核
func (c *connection) Write(b []byte) (int, error) {
	if c.closed {
		return 0, errors.ErrConnClosed
	}

	// 没有历史数据
	if len(c.outBuffer) == 0 {
		writen, err := unix.Write(c.fd, b)
		if err != nil {
			return writen, err
		}

		if writen <= 0 {
			writen = 0
		}
		if writen < len(b) {
			// 没写完，将剩余数据先存入 outBuffer 然后注册读写事件
			// TODO: 需要一个弹性扩容的结构
			c.outBuffer = make([]byte, writen)
			copy(c.outBuffer, b[writen:])
			if err := c.loop.epoll.ModReadWrite(c.fd); err != nil {
				log.Printf("conn write [RegReadWrite] error, %v \n", err)
				return 0, c.Close()
			}
		}
		return len(b), nil
	}

	// 有历史数据，先写入 outBuffer 等待可写事件
	c.outBuffer = append(c.outBuffer, b...)

	return len(b), nil
}

func (c *connection) LocalAddr() net.Addr                { return c.localAddr }
func (c *connection) RemoteAddr() net.Addr               { return c.remoteAddr }
func (c *connection) SetDeadline(t time.Time) error      { return errors.ErrUnsupportedOp }
func (c *connection) SetReadDeadline(t time.Time) error  { return errors.ErrUnsupportedOp }
func (c *connection) SetWriteDeadline(t time.Time) error { return errors.ErrUnsupportedOp }

// handleEvent 作为 reactor 响应 epoll 事件
func (c *connection) handleEvent(_ int, eventType internal.EventType) error {
	if eventType&unix.EPOLLIN != 0 {
		return c.loop.handleReadEvent(c)
	}

	if eventType&unix.EPOLLOUT != 0 && len(c.outBuffer) != 0 {
		return c.loop.handleWriteEvent(c)
	}
	return nil
}

// Close 关闭连接
func (c *connection) Close() error {
	if c.loop.ser.onClose != nil {
		c.loop.ser.onClose(c)
	}
	delete(c.loop.reactor, c.fd)
	// 关闭连接，不用关闭 loop
	c.loop = nil
	c.outBuffer = nil
	c.closed = true
	// 关闭 connfd
	if err := unix.Close(c.fd); err != nil {
		return err
	}
	return nil
}

func (c *connection) IsOpen() bool { return !c.closed }
