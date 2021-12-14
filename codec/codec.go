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
	// lengthAdjustment 修正 lengthFieldLength 指定的消息体长度，和Netty中该字段的定义有点差别
	lengthAdjustment int
	// lengthIncludesLengthFieldLength 和 lengthFieldLength 配合使用
	lengthIncludesLengthFieldLength bool
}

func NewLengthFieldCodec(opts ...Option) *LengthFieldCodec {
	codec := &LengthFieldCodec{
		byteOrder:                       binary.LittleEndian,
		lengthFieldOffset:               0,
		initialBytesToStrip:             0,
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

type inputByte []byte

func (in *inputByte) readN(n uint64) ([]byte, error) {
	if n <= 0 || n > uint64(len(*in)) {
		return nil, errorset.ErrReadLengthInvalid
	}
	// 读取前n个字节的数据
	head := (*in)[0:n]
	// slice指针也同步到n
	*in = (*in)[n:]
	return head, nil
}

func (lc *LengthFieldCodec) Decode(data []byte) ([]byte, error) {
	var (
		err    error
		in     inputByte = data
		header []byte
	)
	if lc.lengthFieldOffset > 0 {
		header, err = in.readN(uint64(lc.lengthFieldOffset))
		if err != nil {
			return nil, err
		}
	}

	// 需要修改in指针指向，传入指针
	lenField, length, err := lc.getUnadjustedFrameLength(&in)
	if err != nil {
		return nil, err
	}

	// adjusted frame length
	frameLength := length + uint64(lc.lengthAdjustment)
	if lc.lengthIncludesLengthFieldLength {
		frameLength -= uint64(lc.lengthFieldLength)
	}

	frame, err := in.readN(frameLength)
	if err != nil {
		return nil, err
	}

	msg := make([]byte, uint64(len(header)+len(lenField))+frameLength)
	copy(msg, header)
	copy(msg[len(header):], lenField)
	copy(msg[(len(header)+len(lenField)):], frame)

	return msg[lc.initialBytesToStrip:], err
}

// 获取未调整前原始的数据帧长度， LengthField 中指定的长度
func (lc *LengthFieldCodec) getUnadjustedFrameLength(in *inputByte) ([]byte, uint64, error) {
	switch lc.lengthFieldLength {
	case 1:
		lenField, err := in.readN(1)
		if err != nil {
			return nil, 0, err
		}
		return lenField, uint64(lenField[0]), nil
	case 2:
		lenField, err := in.readN(2)
		if err != nil {
			return nil, 0, err
		}
		return lenField, uint64(lc.byteOrder.Uint16(lenField)), nil
	case 4:
		lenField, err := in.readN(4)
		if err != nil {
			return nil, 0, err
		}
		return lenField, uint64(lc.byteOrder.Uint32(lenField)), nil
	case 8:
		lenField, err := in.readN(8)
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

func withByteOrder(order binary.ByteOrder) Option {
	return byteOrderOpt{order: order}
}

type lengthFieldOffsetOpt int

func (opt lengthFieldOffsetOpt) apply(lc *LengthFieldCodec) {
	lc.lengthFieldOffset = int(opt)
}

func withLengthFieldOffset(offset int) Option {
	return lengthFieldOffsetOpt(offset)
}

type initialBytesToStripOpt int

func (opt initialBytesToStripOpt) apply(lc *LengthFieldCodec) {
	lc.initialBytesToStrip = int(opt)
}

func withInitialBytesToStrip(strip int) Option {
	return initialBytesToStripOpt(strip)
}

type lengthFieldLengthOpt int

func (opt lengthFieldLengthOpt) apply(lc *LengthFieldCodec) {
	lc.lengthFieldLength = int(opt)
}

func withLengthFieldLength(length int) Option {
	return lengthFieldLengthOpt(length)
}

type lengthAdjustmentOpt int

func (opt lengthAdjustmentOpt) apply(lc *LengthFieldCodec) {
	lc.lengthAdjustment = int(opt)
}

func withLengthAdjustment(length int) Option {
	return lengthAdjustmentOpt(length)
}

type lengthIncludesLengthFieldLengthOpt bool

func (opt lengthIncludesLengthFieldLengthOpt) apply(lc *LengthFieldCodec) {
	lc.lengthIncludesLengthFieldLength = bool(opt)
}

func withLengthIncludesLengthFieldLength(b bool) Option {
	return lengthIncludesLengthFieldLengthOpt(b)
}
