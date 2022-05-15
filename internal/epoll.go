package internal

import (
	"golang.org/x/sys/unix"
	"log"
)

// Epoll epoll 封装
type Epoll struct {
	// epfd，epoll_create()产生的唯一标识epoll对象的文件描述符
	epfd int

	// 参考：https://zhuanlan.zhihu.com/p/393748176
	// eventfd，通过eventfd()调用产生用于事件通知的fd，只监听读事件，这里用来内部手动唤醒eventloop处理任务
	eventfd int
	// eventfd 对应文件内容buf，避免重复开辟空间
	eventfdBuf []byte

	taskQueue []func(interface{}) error
}

type EventType = uint32

//
// const (
//
// 	// ReadEvents EPOLLPRI 代表有紧急数据需要读取
// 	// https://stackoverflow.com/questions/15422919/difference-between-pollin-and-pollpri-in-poll-syscall
// 	ReadEvents = unix.EPOLLIN | unix.EPOLLPRI
//
// 	WriteEvent = unix.EPOLLOUT
//
// 	ErrEvent = unix.EPOLLERR
// )

// CreateEpoll 创建Epoll实例
func CreateEpoll() (*Epoll, error) {
	epoll := new(Epoll)
	// 创建epoll实例，设置EPOLL_CLOEXEC标识，避免泄露
	// https://evian-zhang.github.io/introduction-to-linux-x86_64-syscall/src/filesystem/epoll_create-epoll_wait-epoll_ctl-epoll_pwait-epoll_create1.html
	fd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	epoll.epfd = fd

	// https://evian-zhang.github.io/introduction-to-linux-x86_64-syscall/src/filesystem/eventfd-eventfd2.html
	// 创建eventfd，用于程序内部唤醒eventloop。设置为非阻塞，计数器为0返回EAGAIN
	epoll.eventfd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		_ = epoll.Close()
		return nil, err
	}

	// 注册 eventfd 读事件到epfd上  参考：https://zhuanlan.zhihu.com/p/393748176
	// eventfd 底层实现是一个8字节的计数器，计数不为0就代表可读，往eventfd写入的时候累加计数，read后清零，所以eventfd是一直可写的
	// 我们这里是为了在程序内主动唤醒eventloop，所以我们只需要监听读事件即可
	if err := epoll.RegRead(epoll.eventfd); err != nil {
		_ = epoll.Close()
		return nil, err
	}
	epoll.eventfdBuf = make([]byte, 8)
	return epoll, nil
}

// Polling 阻塞在EpollWait，等待事件就绪后调用callback
func (ep *Epoll) Polling(callback func(fd int, eventType EventType) error) error {
	events := make([]unix.EpollEvent, 1024)
	for {
		// 阻塞直到有事件就绪
		numPolled, err := unix.EpollWait(ep.epfd, events, -1)
		// EINTR https://man7.org/linux/man-pages/man2/epoll_wait.2.html
		if err != nil && err != unix.EINTR {
			log.Printf("epollwait error, %v \n", err)
			continue
		}

		var runTask bool

		for i := 0; i < numPolled; i++ {
			ev := events[i]
			if pfd := int(ev.Fd); pfd != ep.eventfd { // io事件就绪，非内部任务
				// ev.Events 是一个 bitmask， 可能出现的事件： https://man7.org/linux/man-pages/man2/epoll_ctl.2.html
				if err := callback(pfd, ev.Events); err != nil {
					log.Printf("callback error, %v \n", err)
					continue
				}
			} else { // WakeUp 主动唤醒，执行内部任务，比如定时任务之类
				// 将 eventfd 中的数据读取出来清零，解除读就绪事件，避免 epoll 被重复唤醒，早期 evio 有这个bug
				_, _ = unix.Read(ep.eventfd, ep.eventfdBuf)
				// log.Printf("eventfd buf : %v \n", ep.eventfdBuf)
				runTask = true
			}
		}

		if runTask {
			log.Println("run all task !")
		}
		return nil
	}
}

// WakeUp 主动唤醒eventloop，执行任务（非IO事件任务）
func (ep *Epoll) WakeUp() error {
	// 向 eventfd 写入数据，触发可读事件，唤醒 EpollWait（写入必须是一个8字节数）
	if _, err := unix.Write(ep.eventfd, []byte{0, 0, 0, 0, 0, 0, 0, 1}); err != nil {
		return err
	}
	return nil
}

const (
	readEvent      = unix.EPOLLPRI | unix.EPOLLIN
	writeEvent     = unix.EPOLLOUT
	readWriteEvent = readEvent | writeEvent
)

// RegReadWrite 注册fd读写事件到 epoll
func (ep *Epoll) RegReadWrite(fd int) error {
	return unix.EpollCtl(ep.epfd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: readWriteEvent,
		Fd:     int32(fd),
	})
}

// RegRead 注册fd读事件到 epoll
func (ep *Epoll) RegRead(fd int) error {
	return unix.EpollCtl(ep.epfd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: readEvent,
		Fd:     int32(fd),
	})
}

// RegWrite 注册fd写事件到 epoll
func (ep *Epoll) RegWrite(fd int) error {
	return unix.EpollCtl(ep.epfd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: writeEvent,
		Fd:     int32(fd),
	})
}

// ModReadWrite 修改fd注册事件为读写事件
func (ep *Epoll) ModReadWrite(fd int) error {
	return unix.EpollCtl(ep.epfd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Events: readWriteEvent,
		Fd:     int32(fd),
	})
}

// ModRead 修改fd注册事件为读事件
func (ep *Epoll) ModRead(fd int) error {
	return unix.EpollCtl(ep.epfd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Events: readEvent,
		Fd:     int32(fd),
	})
}

// ModWrite 修改fd注册事件为写事件
func (ep *Epoll) ModWrite(fd int) error {
	return unix.EpollCtl(ep.epfd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Events: writeEvent,
		Fd:     int32(fd),
	})
}

// Delete 在 epoll 上删除 fd （不再监听）
func (ep *Epoll) Delete(fd int) error {
	return unix.EpollCtl(ep.epfd, unix.EPOLL_CTL_DEL, fd, nil)
}

func (ep *Epoll) Close() error {
	if err := unix.Close(ep.epfd); err != nil {
		return err
	}
	return unix.Close(ep.eventfd)
}
