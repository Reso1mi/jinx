package jinx

import (
	"github.com/imlgw/jinx/internal"
)

type Reactor interface {
	HandleEvent(fd int, eventType internal.EventType) error
	Run()
}
