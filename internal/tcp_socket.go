package internal

import (
	"golang.org/x/sys/unix"
	"net"
)

// SocketListen 参考：https://zhuanlan.zhihu.com/p/399651675
func SocketListen(network, addr string) (int, *net.TCPAddr, error) {
	// 创建一个 socketfd，暂时只支持 tcp4
	// https://man7.org/linux/man-pages/man2/socket.2.html
	socketfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_TCP)
	if err != nil {
		return -1, nil, err
	}

	// 解析地址
	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return -1, nil, err
	}

	// 转换为 unix.SockaddrInet4 用于 bind()
	inet4 := &unix.SockaddrInet4{Port: tcpAddr.Port}
	copy(inet4.Addr[:], tcpAddr.IP)

	// 绑定 socketfd 和地址，https://man7.org/linux/man-pages/man2/bind.2.html
	if err = unix.Bind(socketfd, inet4); err != nil {
		return -1, nil, err
	}

	// 转换为监听套接字 https://man7.org/linux/man-pages/man2/listen.2.html
	// 第二个参数为「全连接队列长度」--> /proc/sys/net/core/somaxconn 默认4096
	if err = unix.Listen(socketfd, unix.SOMAXCONN); err != nil {
		return -1, nil, err
	}

	return socketfd, tcpAddr, nil
}
