package codec

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestLengthFieldCodec2(t *testing.T) {
	codec := NewLengthFieldCodec(
		withLengthFieldLength(2),
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

	// 解码数据
	res, err := codec.Decode(out)
	if err != nil {
		t.Fatalf("decocde data err: %v", err)
	}
	if !bytes.Equal(res, out) {
		t.Fatalf("decode data[%s] shuould equal 2 src data[%s]", out, res)
	}
}

func TestLengthFieldCodec2WithJust2(t *testing.T) {
	codec := NewLengthFieldCodec(
		withLengthFieldLength(2),
		// 待编码数据包含长度字段
		withLengthIncludesLengthFieldLength(true),
		withLengthAdjustment(-2),
	)

	in := make([]byte, 1<<16-1)
	if _, err := rand.Read(in); err != nil {
		t.Fatal(err)
	}
	// | 2 | 65535 |
	out, _ := codec.Encode(in)
	// encode 将长度写入头部，数据部分不变
	if !bytes.Equal(out[2:], in) {
		t.Fatalf("encoded data should equal 2 src data")
	}

	// 解码数据
	res, err := codec.Decode(out)
	if err != nil {
		t.Fatalf("decocde data err: %v", err)
	}
	if !bytes.Equal(res, out) {
		t.Fatalf("decode data[%s] shuould equal 2 src data[%s]", out, res)
	}
}
