package main

import (
	. "github.com/imlgw/jinx"
	"log"
)

func main() {
	network := "tcp"
	addr := ":9876"

	server, err := NewServer(network, addr, WithLb(RoundRobin), WithLoopNum(4), WithServerName("Resolmi"))
	if err != nil {
		log.Fatal(err)
		return
	}

	server.OnBoot(func(s Server) {
		log.Printf("\nserver info: \nname: [%s] \nnetwork: [%s] \naddr:[%s]\n",
			s.ServerName(), s.Network(), s.ServerAddr())
	})

	server.OnOpen(func(c Conn) {
		log.Printf("\nnew conn establish \nisOpen: [%v]  \nremoteAddr: [%v]",
			c.IsOpen(), c.RemoteAddr())
	})

	server.OnRead(func(c Conn) {
		buf := make([]byte, 1024)
		readn, err := c.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := c.Write(buf[:readn]); err != nil {
			log.Fatal(err)
		}
	})

	server.OnClose(func(c Conn) {
		log.Printf("\nconn closed \nisOpen: [%v] \nlocalAddr: [%v] \nremoteAddr: [%v]",
			c.IsOpen(), c.LocalAddr(), c.RemoteAddr())
	})

	server.OnShutdown(func(s Server) {
		log.Printf("\nserver shutdown: \nname: [%s] \nnetwork: [%s] \naddr:[%s]",
			s.ServerName(), s.Network(), s.ServerAddr())
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
