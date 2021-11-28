package conn

import (
	"fmt"
	"jinx"
	"jinx/config"
	"jinx/request"
	"jinx/router"
	"net"
)

// HandleFunc 处理链接业务的方法
type HandleFunc func(*net.TCPConn, []byte, int) error

// 不对外暴露，强制使用New创建
type connection struct {
	// tcp套接字
	Conn *net.TCPConn
	// 链接的ID
	ConnID uint
	// 链接的状态
	isClose bool
	// Router绑定
	Router router.Router
	// 等待链接退出的channel
	ExitChan chan bool
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
	if err := c.Conn.Close(); err != nil {
		fmt.Println("Connection Close err", err)
		return
	}
	close(c.ExitChan)
}

func (c *connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *connection) GetConnID() uint {
	return c.ConnID
}

func (c *connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *connection) Send(bytes []byte) error {
	panic("implement me")
}

func (c *connection) Read() {
	fmt.Println("[Jinx] Reader goroutine is running")
	defer fmt.Println("[Jinx] Reader Stop")
	defer c.Stop()

	for {
		buf := make([]byte, config.ServerConfig.MaxPackSize)
		cnt, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("[Jinx] Read from Client error", err)
			break
		}
		req := request.NewRequest(c, buf[:cnt])
		// 执行路由绑定的方法
		go func() {
			c.Router.BeforeHandle(req)
			c.Router.Handle(req)
			c.Router.AfterHandle(req)
		}()
	}
}

func NewConnection(conn *net.TCPConn, connID uint, router router.Router) jinx.Connection {
	c := &connection{
		Conn:     conn,
		ConnID:   connID,
		isClose:  false,
		Router:   router,
		ExitChan: make(chan bool, 1),
	}
	return c
}
