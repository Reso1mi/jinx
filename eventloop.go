package jinx

import (
	"github.com/imlgw/jinx/internal"
	"log"
)

type eventloop struct {
	idx int // idx

	epoll *internal.Epoll

	reactor map[int]Reactor // fd 对应的 Reactor

	conncnt uint64

	buffer []byte // read buffer 默认64KB
}

// NewLoop 创建一个事件循环，idx 为循环序号
func NewLoop(idx int) (*eventloop, error) {
	epoll, err := internal.CreateEpoll()
	if err != nil {
		return nil, err
	}

	return &eventloop{
		epoll:   epoll,
		idx:     idx,
		reactor: make(map[int]Reactor),
		conncnt: 0,
		buffer:  make([]byte, 0xffff),
	}, nil
}

// Loop 开始事件循环
func (loop *eventloop) Loop() error {
	if err := loop.epoll.Polling(
		func(fd int, eventType internal.EventType) error {
			r, ok := loop.reactor[fd]
			if ok {
				if err := r.HandleEvent(fd, eventType); err != nil {
					return err
				}
			} else {
				// 可能是不再使用的fd，netty 中的处理是直接在 epoll 事件上删除了这个 fd
				if err := loop.epoll.Delete(fd); err != nil {
					log.Printf("remove unused fd err, %v \n", err)
				}
			}
			return nil
		}); err != nil {
		return err
	}
	return nil
}

// Stop 停止事件循环
func (loop *eventloop) Stop() error {
	if err := loop.epoll.Close(); err != nil {
		return err
	}
	return nil
}
