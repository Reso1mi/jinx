package codec

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestLengthFieldCodec2_Encode(t *testing.T) {
	codec := NewLengthFieldCodec(
		WithLengthFieldLength(2),
	)

	in := make([]byte, 1<<16-1)
	if _, err := rand.Read(in); err != nil {
		t.Fatal(err)
	}
	// 超出指定的长度
	if _, err := codec.Encode(in); err != nil {
		t.Fatal("encode data err, over the lengthField")
	}

	in = make([]byte, 1<<16-1)
	if _, err := rand.Read(in); err != nil {
		t.Fatal(err)
	}
	out, _ := codec.Encode(in)
	// encode 将长度写入头部，数据部分不变
	if !bytes.Equal(out[2:], in) {
		t.Fatalf("encoded data should equal 2 src data")
	}
}

func TestLengthFieldCodec2WithJust2_Encode(t *testing.T) {
	codec := NewLengthFieldCodec(
		WithLengthFieldLength(2),
		// 待编码数据包含长度字段
		WithLengthIncludesLengthFieldLength(true),
	)

	in := make([]byte, 1<<16-3)
	if _, err := rand.Read(in); err != nil {
		t.Fatal(err)
	}
	// | 2 | 65535 |
	out, err := codec.Encode(in)
	if err != nil {
		t.Fatalf("encode err: %v", err)
	}
	// encode 将长度写入头部，数据部分不变
	if !bytes.Equal(out[2:], in) {
		t.Fatalf("encoded data should equal 2 src data")
	}
}

func TestNewLengthFieldCodec(t *testing.T) {
	codec := NewLengthFieldCodec(
		WithLengthFieldLength(2),
		WithLengthIncludesLengthFieldLength(true),
	)

	go func() {
		addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8999")
		tcpListener, _ := net.ListenTCP("tcp", addr)
		for {
			tcpConn, _ := tcpListener.AcceptTCP()
			for {
				decoded, err := codec.Decode(tcpConn)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("server receive:", string(decoded[2:]))
			}
		}
	}()
	time.Sleep(3 * time.Second)
	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:8999")
		if err != nil {
			fmt.Println("link err", err)
			return
		}

		sendData := make([]byte, 0)

		imlgwSite := "imlgw.top"
		encoded, _ := codec.Encode([]byte(imlgwSite))

		for i := 0; i < 20; i++ {
			sendData = append(sendData, encoded...)
		}

		write, err := conn.Write(sendData)
		fmt.Println("client send:", sendData)
		fmt.Println(write)
	}()

	select {}
}
