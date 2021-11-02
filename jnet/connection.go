package jnet

import (
	"fmt"
	"jinx/contract"
	"net"
)

type Connection struct {
	// tcp套接字
	Conn *net.TCPConn
	// 链接的ID
	ConnID uint
	// 链接的状态
	isClose bool
	// 和链接绑定的业务方法
	handle contract.HandleFunc
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
		buf := make([]byte, 512)
		cnt, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("[Jinx] Read from Client error", err)
			continue
		}

		// 调用回调函数
		if err := c.handle(c.Conn, buf, cnt); err != nil {
			fmt.Println("[Jinx] handle error!", err)
		}
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

func NewConnection(conn *net.TCPConn, connID uint, callback contract.HandleFunc) *Connection {
	c := &Connection{
		Conn:     conn,
		ConnID:   connID,
		isClose:  false,
		handle:   callback,
		ExitChan: make(chan bool, 1),
	}
	return c
}
