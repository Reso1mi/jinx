package jnet

import (
	"fmt"
	"jinx/config"
	"jinx/contract/connection"
	"jinx/contract/router"
	"net"
)

type Connection struct {
	// tcp套接字
	Conn *net.TCPConn
	// 链接的ID
	ConnID uint
	// 链接的状态
	isClose bool
	// Router绑定
	Router router.IRouter
	// 等待链接退出的channel
	ExitChan chan bool
}

func (c *Connection) Start() {
	fmt.Println("[Jinx] Connection Start... ConnID = ", c.GetConnID())
	go c.Read()
}

func (c *Connection) Read() {
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
		request := &Request{
			conn: c,
			data: buf[:cnt],
		}
		// 执行路由绑定的方法
		go func() {
			c.Router.BeforeHandle(request)
			c.Router.Handle(request)
			c.Router.AfterHandle(request)
		}()
	}
}

func (c *Connection) Stop() {
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

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnID() uint {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) Send(bytes []byte) error {
	panic("implement me")
}

func NewConnection(conn *net.TCPConn, connID uint, router router.IRouter) connection.IConnection {
	c := &Connection{
		Conn:     conn,
		ConnID:   connID,
		isClose:  false,
		Router:   router,
		ExitChan: make(chan bool, 1),
	}
	return c
}
