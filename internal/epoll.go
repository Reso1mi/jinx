package internal

import "golang.org/x/sys/unix"

// Epoll epoll 封装
type Epoll struct {
	// epfd，epoll_create()产生的唯一标识epoll对象的文件描述符
	epfd int

	// 参考：https://zhuanlan.zhihu.com/p/393748176
	// eventfd，通过eventfd()调用产生用于事件通知的fd，只监听读事件，这里用来内部手动唤醒eventloop处理任务
	eventfd int
}

type EventType uint32

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
	// 创建eventfd，用于程序内部唤醒eventloop
	epoll.eventfd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		_ = epoll.Close()
		return nil, err
	}

	// 给 eventfd 注册读事件到epfd上  参考：https://zhuanlan.zhihu.com/p/393748176
	// eventfd 底层实现是一个8字节的计数器，计数不为0就代表可读，往eventfd写入的时候累加计数，read后清零，所以eventfd是一直可写的
	// 我们这里是为了在程序内主动唤醒eventloop，所以我们只需要监听读事件即可
	if err := epoll.RegRead(epoll.eventfd); err != nil {
		_ = epoll.Close()
		return nil, err
	}

	return epoll, nil
}

// Polling 阻塞在EpollWait，等待事件就绪后调用callback
func (ep *Epoll) Polling(callback func(fd int, eventType EventType)) {

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

func (ep *Epoll) Close() error {
	if err := unix.Close(ep.epfd); err != nil {
		return err
	}
	return unix.Close(ep.eventfd)
}
