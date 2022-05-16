package jinx

import (
	"github.com/imlgw/jinx/internal"
)

type Reactor interface {
	handleEvent(fd int, eventType internal.EventType) error
}
