package jinx

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	fmt.Println("client start...")
	// 1.连接远程服务器
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("link err", err)
		return
	}
	for {
		// 2.调用write写数据
		if _, err := conn.Write([]byte("imlgw.top")); err != nil {
			fmt.Println("write err", err)
			return
		}
		buf := make([]byte, 512)
		// 3.读取服务端返回的内容
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read buf err", err)
		}
		fmt.Printf("tcpserver return: %s, cnt = %d\n", buf[:cnt], cnt)
		time.Sleep(1 * time.Second)
	}
}
