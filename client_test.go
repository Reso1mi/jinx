package jinx

import (
	"fmt"
	"github.com/imlgw/jinx/codec"
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
		lengthFieldCodec := codec.NewLengthFieldCodec(
			codec.WithLengthFieldLength(2),
		)
		data := []byte("imlgw.top")
		encoded, err := lengthFieldCodec.Encode(data)
		// 2.调用write写数据
		if _, err := conn.Write(encoded); err != nil {
			fmt.Println("write err", err)
			return
		}
		// 3.读取服务端返回的内容
		res, err := lengthFieldCodec.Decode(conn)
		if err != nil {
			fmt.Println("read buf err", err)
		}
		fmt.Printf("tcpserver return: %s\n", res)
		time.Sleep(1 * time.Second)
	}
}
