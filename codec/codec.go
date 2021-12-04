package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/imlgw/jinx/errors"
)

type Codec interface {
	Encode(data []byte) []byte

	Decode(data []byte) []byte
}

type LengthFieldCodec struct {
	byteOrder binary.ByteOrder
	// 解码时跳过字节数
	initialBytesToStrip int
	// 长度字段长度
	lengthFieldLength int
	// 修正数据长度
	lengthAdjustment int
	// 和 lengthAdjustment 配合使用
	lengthIncludesLengthFieldLength bool
}

func (cc *LengthFieldCodec) Encode(data []byte) ([]byte, error) {
	var out []byte
	// 1. 将content长度写入out
	length := len(data) + cc.lengthAdjustment
	if cc.lengthIncludesLengthFieldLength {
		length += cc.lengthFieldLength
	}

	if length < 0 {
		return nil, errorset.ErrTooLessLength
	}

	switch cc.lengthFieldLength {
	case 1:
		if length >= 1<<8 {
			return nil, fmt.Errorf("more than 1 byte in length, %d", length)
		}
		out = []byte{byte(length)}
	case 2:
		if length >= 1<<16 {
			return nil, fmt.Errorf("more than 2 byte in length, %d", length)
		}
		out = make([]byte, 2)
		cc.byteOrder.PutUint16(out, uint16(length))
	case 4:
		if length >= 1<<32 {
			return nil, fmt.Errorf("more than 4 byte in length, %d", length)
		}
		out = make([]byte, 4)
		cc.byteOrder.PutUint32(out, uint32(length))
	case 8:
		if length >= 1<<64 {
			return nil, fmt.Errorf("more than 8 byte in length, %d", length)
		}
		cc.byteOrder.PutUint64(out, uint64(length))
	default:
		return nil, errorset.ErrUnsupportedLength
	}

	// 2. 将数据写入out
	return append(out, data...), nil
}

func (cc *LengthFieldCodec) Decode(data []byte) ([]byte, error) {
	var err error

	return nil, err
}

type bytes []byte

func (bs *bytes) readN(n int) ([]byte, error) {
	if n <= 0 || n > len(*bs) {
		return nil, errorset.ErrReadLengthInvalid
	}
	// 读取前n个字节的数据
	head := (*bs)[0:n]
	// slice指针也同步到n
	*bs = (*bs)[n:]
	return head, nil
}
