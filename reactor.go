package jinx

import (
	"github.com/imlgw/jinx/internal"
)

type reactor interface {
	handleEvent(fd int, eventType internal.EventType) error
	Close() error
}
