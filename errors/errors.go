package errors

import "errors"

var (
	// ================================================= codec errors =================================================.

	// ErrInvalidFixedLength occurs when the output data have invalid fixed length.
	ErrInvalidFixedLength = errors.New("invalid fixed length of bytes")
	// ErrUnexpectedEOF occurs when no enough data to read by codec.
	ErrUnexpectedEOF = errors.New("there is no enough data")
	// ErrUnsupportedLength occurs when unsupported lengthFieldLength is from input data.
	ErrUnsupportedLength = errors.New("unsupported lengthFieldLength. (expected: 1, 2, 3, 4, or 8)")
	// ErrTooLessLength occurs when adjusted frame length is less than zero.
	ErrTooLessLength = errors.New("adjusted frame length is less than zero")
	// ErrReadLengthInvalid 读取长度异常
	ErrReadLengthInvalid = errors.New("read length is invalid")

	// ================================================= connect errors ===============================================.

	// ErrConnectionClosed 读取长度异常
	ErrConnectionClosed = errors.New("read length is invalid")

	// ErrAcceptSocket 连接异常
	ErrAcceptSocket = errors.New("accept error")
)
