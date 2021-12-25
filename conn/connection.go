package conn

import (
	"fmt"
	"github.com/imlgw/jinx"
	"github.com/imlgw/jinx/codec"
	errorset "github.com/imlgw/jinx/errors"
	"github.com/imlgw/jinx/request"
	"github.com/imlgw/jinx/router"
	"net"
)

// HandleFunc 处理链接业务的方法
type HandleFunc func(*net.TCPConn, []byte, int) error

// 不对外暴露，强制使用New创建
type connection struct {
	// tcp套接字
	conn *net.TCPConn
	// 链接的ID
	connID uint
	// 链接的状态
	isClose bool
	// Router绑定
	router router.Router
	// 等待链接退出的channel
	exitChan chan bool
	// 编解码器
	codec codec.ICodec
}

func (c *connection) Start() {
	fmt.Println("[Jinx] Connection Start... ConnID = ", c.GetConnID())
	go c.Read()
}

func (c *connection) Stop() {
	fmt.Println("[Jinx] Connection Stop... ConnID = ", c.GetConnID())
	if c.isClose {
		return
	}
	c.isClose = true
	if err := c.conn.Close(); err != nil {
		fmt.Println("Connection Close err", err)
		return
	}
	close(c.exitChan)
}

func (c *connection) GetTCPConnection() *net.TCPConn {
	return c.conn
}

func (c *connection) GetConnID() uint {
	return c.connID
}

func (c *connection) Send(data []byte) error {
	if c.isClose {
		return errorset.ErrConnectionClosed
	}
	// encode data
	encoded, err := c.codec.Encode(data)
	if err != nil {
		fmt.Println("encode data err", string(data))
		return err
	}
	if _, err := c.conn.Write(encoded); err != nil {
		fmt.Println("conn write msg err", err)
		return err
	}
	return nil
}

func (c *connection) Read() {
	fmt.Println("[Jinx] Reader goroutine is running")
	defer fmt.Println("[Jinx] Reader Stop")
	defer c.Stop()

	for {
		buf, err := c.codec.Decode(c.conn)
		if err != nil {
			fmt.Println("[Jinx] Read from Client errors", err)
			break
		}
		req := request.NewRequest(c, buf)
		// 执行路由绑定的方法
		go func() {
			c.router.BeforeHandle(req)
			c.router.Handle(req)
			c.router.AfterHandle(req)
		}()
	}
}

func NewConnection(conn *net.TCPConn, connID uint, router router.Router, codec codec.ICodec) jinx.Connection {
	c := &connection{
		conn:     conn,
		connID:   connID,
		isClose:  false,
		router:   router,
		exitChan: make(chan bool, 1),
		codec:    codec,
	}
	return c
}
