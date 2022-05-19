package internal

import (
	"golang.org/x/sys/unix"
	"net"
	"sync"
	"testing"
)

func TestSocketBind(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		lnfd, _, err := SocketListen("tcp", ":9876")
		if err != nil {
			t.Fatal(err)
		}
		wg.Done()
		connfd, _, err := unix.Accept(lnfd)
		buf := make([]byte, 50)
		// unix.Read(connfd, buf)
		// fmt.Println("recv:", string(buf))
		if n, err := unix.Read(connfd, buf); err != nil || string(buf[:n]) != "Resolmi" {
			t.Fatal(err)
		}
		wg.Done()
	}()
	wg.Wait()
	wg.Add(1)
	conn, err := net.Dial("tcp", ":9876")
	if err != nil {
		t.Fatal(err)
		return
	}

	if _, err := conn.Write([]byte("Resolmi")); err != nil {
		return
	}

	wg.Wait()
}
