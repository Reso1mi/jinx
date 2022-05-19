package jinx

import (
	"testing"
	"time"
)

func TestEventloop(t *testing.T) {

	// server, err := NewServer("tcp", ":9876", WithLb(RoundRobin), WithLoopNum(4))
	// if err != nil {
	// 	t.Fatal(err)
	// 	return
	// }
	//
	// if err := server.Run(); err != nil {
	// 	t.Fatal(err)
	// 	return
	// }
	//
	// time.Sleep(3 * time.Second)
	//
	// go func() {
	// 	// client
	// 	conn, err := net.Dial("tcp", ":9876")
	// 	if err != nil {
	// 		t.Fatal(err)
	// 		return
	// 	}
	//
	// 	msg := "hello, jinx!"
	// 	if _, err := conn.Write([]byte(msg)); err != nil {
	// 		t.Fatal(err)
	// 		return
	// 	}
	// }()

	time.Sleep(5 * time.Second)
}
