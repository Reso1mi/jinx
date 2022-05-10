package internal

import (
	"fmt"
	"testing"
	"time"
)

func TestEpoll_Polling(t *testing.T) {
	epoll, err := CreateEpoll()
	if err != nil {
		t.Fatal("create epoll fail")
	}

	go func() {
		epoll.Polling(func(fd int, eventType EventType) error {
			t.Fatal("callback shouldn't exec")
			return nil
		})
	}()

	time.Sleep(time.Second * 3)

	if err := epoll.WakeUp(); err != nil {
		fmt.Println(err)
		t.Fatal("wakeUp fail")
	}

	time.Sleep(time.Second * 3)

	if err := epoll.Close(); err != nil {
		t.Fatal("epoll close fail!")
	}
}
