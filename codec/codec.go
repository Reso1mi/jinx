package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/imlgw/jinx/errors"
	"io"
	"net"
)

type ICodec interface {
	Encode(data []byte) ([]byte, error)

	Decode(conn net.Conn) ([]byte, error)
}

/*
  参考 netty 的 LengthFieldPrepender 以及 LengthFieldBasedFrameDecoder
*/

type LengthFieldCodec struct {
	// byteOrder
	byteOrder binary.ByteOrder
	// lengthFieldOffset 解码时长度字段偏移量
	lengthFieldOffset int
	// initialBytesToStrip 解码时跳过字节数
	initialBytesToStrip int
	// lengthFieldLength 长度字段长度
	lengthFieldLength int
	// lengthIncludesLengthFieldLength 和 lengthFieldLength 配合使用
	lengthIncludesLengthFieldLength bool
	// lengthAdjustment 修正 lengthFieldLength 指定的消息体长度，和Netty中该字段的定义有点差别，与 lengthIncludesLengthFieldLength 无关
	lengthAdjustment int
}

func NewLengthFieldCodec(opts ...Option) *LengthFieldCodec {
	codec := &LengthFieldCodec{
		byteOrder:                       binary.LittleEndian,
		lengthFieldOffset:               0,
		initialBytesToStrip:             2,
		lengthFieldLength:               2,
		lengthAdjustment:                0,
		lengthIncludesLengthFieldLength: false,
	}
	for _, opt := range opts {
		opt.apply(codec)
	}
	return codec
}

func (lc *LengthFieldCodec) Encode(data []byte) ([]byte, error) {
	var out []byte
	// 将content长度写入out
	length := len(data) - lc.lengthAdjustment
	if lc.lengthIncludesLengthFieldLength {
		length += lc.lengthFieldLength
	}

	if length < 0 {
		return nil, errorset.ErrTooLessLength
	}

	switch lc.lengthFieldLength {
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
		lc.byteOrder.PutUint16(out, uint16(length))
	case 4:
		if length >= 1<<32 {
			return nil, fmt.Errorf("more than 4 byte in length, %d", length)
		}
		out = make([]byte, 4)
		lc.byteOrder.PutUint32(out, uint32(length))
	case 8:
		if uint64(length) > uint64(1<<64-1) {
			return nil, fmt.Errorf("more than 8 byte in length, %d", length)
		}
		lc.byteOrder.PutUint64(out, uint64(length))
	default:
		return nil, errorset.ErrUnsupportedLength
	}

	// 2. 将数据写入out
	return append(out, data...), nil
}

func readN(c net.Conn, n int) ([]byte, error) {
	var buf = make([]byte, n)
	//c.Read()可能会超时，可能是隐藏的问题
	if _, err := io.ReadFull(c, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (lc *LengthFieldCodec) Decode(c net.Conn) ([]byte, error) {
	var (
		err    error
		header []byte
	)
	if lc.lengthFieldOffset > 0 {
		header, err = readN(c, lc.lengthFieldOffset)
		if err != nil {
			return nil, err
		}
	}

	lenField, length, err := lc.getUnadjustedFrameLength(c)
	if err != nil {
		return nil, err
	}

	// todo: 超大数据支持. 这里将length转成int, 传输超大payload肯定会丢失数据.
	//       所以实际上框架目前并不支持超大数据传输(感觉也没有必要，net.Conn一次Read返回的数据长度也是int)
	// adjusted frame length
	frameLength := int(length) + lc.lengthAdjustment
	if lc.lengthIncludesLengthFieldLength {
		// 减去包头长度，获取调整后的帧长度
		frameLength -= lc.lengthFieldLength
	}

	frame, err := readN(c, frameLength)
	if err != nil {
		return nil, err
	}

	msg := make([]byte, len(header)+len(lenField)+frameLength)
	copy(msg, header)
	copy(msg[len(header):], lenField)
	copy(msg[(len(header)+len(lenField)):], frame)

	return msg[lc.initialBytesToStrip:], err
}

// 获取未调整前原始的数据帧长度( LengthField 中指定的长度)
func (lc *LengthFieldCodec) getUnadjustedFrameLength(c net.Conn) ([]byte, uint64, error) {
	switch lc.lengthFieldLength {
	case 1:
		lenField, err := readN(c, 1)
		if err != nil {
			return nil, 0, err
		}
		return lenField, uint64(lenField[0]), nil
	case 2:
		lenField, err := readN(c, 2)
		if err != nil {
			return nil, 0, err
		}
		return lenField, uint64(lc.byteOrder.Uint16(lenField)), nil
	case 4:
		lenField, err := readN(c, 4)
		if err != nil {
			return nil, 0, err
		}
		return lenField, uint64(lc.byteOrder.Uint32(lenField)), nil
	case 8:
		lenField, err := readN(c, 8)
		if err != nil {
			return nil, 0, err
		}
		return lenField, lc.byteOrder.Uint64(lenField), nil
	default:
		return nil, 0, errorset.ErrUnsupportedLength
	}
}

// Option 功能选项
// ×××××××××××××××××××××××××××××××××××××××**************
type Option interface {
	apply(*LengthFieldCodec)
}

type byteOrderOpt struct {
	order binary.ByteOrder
}

func (opt byteOrderOpt) apply(lc *LengthFieldCodec) {
	lc.byteOrder = opt.order
}

func WithByteOrder(order binary.ByteOrder) Option {
	return byteOrderOpt{order: order}
}

type lengthFieldOffsetOpt int

func (opt lengthFieldOffsetOpt) apply(lc *LengthFieldCodec) {
	lc.lengthFieldOffset = int(opt)
}

func WithLengthFieldOffset(offset int) Option {
	return lengthFieldOffsetOpt(offset)
}

type initialBytesToStripOpt int

func (opt initialBytesToStripOpt) apply(lc *LengthFieldCodec) {
	lc.initialBytesToStrip = int(opt)
}

func WithInitialBytesToStrip(strip int) Option {
	return initialBytesToStripOpt(strip)
}

type lengthFieldLengthOpt int

func (opt lengthFieldLengthOpt) apply(lc *LengthFieldCodec) {
	lc.lengthFieldLength = int(opt)
}

func WithLengthFieldLength(length int) Option {
	return lengthFieldLengthOpt(length)
}

type lengthAdjustmentOpt int

func (opt lengthAdjustmentOpt) apply(lc *LengthFieldCodec) {
	lc.lengthAdjustment = int(opt)
}

func WithLengthAdjustment(length int) Option {
	return lengthAdjustmentOpt(length)
}

type lengthIncludesLengthFieldLengthOpt bool

func (opt lengthIncludesLengthFieldLengthOpt) apply(lc *LengthFieldCodec) {
	lc.lengthIncludesLengthFieldLength = bool(opt)
}

func WithLengthIncludesLengthFieldLength(b bool) Option {
	return lengthIncludesLengthFieldLengthOpt(b)
}
