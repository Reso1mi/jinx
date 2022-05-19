package jinx

import (
	"log"
	"net"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestSimpleJinxServer(t *testing.T) {
	network := "tcp"
	addr := ":9876"

	server, err := NewServer(network, addr, WithLb(RoundRobin), WithLoopNum(4), WithServerName("Resolmi"))
	if err != nil {
		t.Fatal(err)
		return
	}

	server.OnBoot(func(s Server) {
		log.Printf("\nserver info: \nname: [%s] \nnetwork: [%s] \naddr:[%s]\n",
			s.ServerName(), s.Network(), s.ServerAddr())
	})

	server.OnOpen(func(c Conn) {
		log.Printf("\nnew conn establish \nisOpen: [%v] \nlocalAddr: [%v], \nremoteAddr: [%v]",
			c.IsOpen(), c.LocalAddr(), c.RemoteAddr())
	})

	server.OnRead(func(c Conn) {
		buf := make([]byte, 1024)
		readn, err := c.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		log.Println("server:", string(buf[:readn]))
		if string(buf[:readn]) != "hello,jinx" {
			t.Fatal("error")
		}

		if _, err := c.Write([]byte("hello,vi")); err != nil {
			t.Fatal(err)
		}
	})

	server.OnWrite(func(c Conn) {
		log.Println("on write")
	})

	server.OnClose(func(c Conn) {
		log.Printf("conn closed \n isOpen: [%v] \n localAddr: [%v] remoteAddr: [%v]",
			c.IsOpen(), c.LocalAddr(), c.RemoteAddr())
	})

	server.OnShutdown(func(s Server) {
		log.Printf("server shutdown: \n  name: [%s] \n network: [%s] \n addr:[%s]\n",
			s.ServerName(), s.Network(), s.ServerAddr())
	})

	if err := server.Run(); err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	for !server.Started() {
	}
	go startClient(network, addr)
	wg.Wait()
	wg.Add(1)
	if err := server.Stop(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}

func startClient(network string, addr string) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		log.Panicln(err)
		return
	}
	if _, err := conn.Write([]byte("hello,jinx")); err != nil {
		log.Panicln(err)
	}
	buf := make([]byte, 1024)
	readn, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("client:", string(buf[:readn]))
	if string(buf[:readn]) != "hello,vi" {
		log.Panicln("error server")
	}
	wg.Done()
}
