package errors

import (
	"errors"
)

var (
	// ================================================= codec errors =================================================.

	// ErrUnsupportedLength occurs when unsupported lengthFieldLength is from input data.
	ErrUnsupportedLength = errors.New("unsupported lengthFieldLength. (expected: 1, 2, 3, 4, or 8)")
	// ErrTooLessLength occurs when adjusted frame length is less than zero.
	ErrTooLessLength = errors.New("adjusted frame length is less than zero")

	// ================================================= connect errors ===============================================.

	// ErrAcceptSocket 连接异常
	ErrAcceptSocket = errors.New("accept error")

	// ErrUnsupportedOp occurs when calling some methods that has not been implemented yet.
	ErrUnsupportedOp = errors.New("unsupported operation")

	// ErrConnClosed occurs when calling some methods that has not been implemented yet.
	ErrConnClosed = errors.New("connection closed")
)
