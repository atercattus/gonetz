package gonetz

import (
	"syscall"
	"unsafe"
)

const (
	maxEpollEvents = 2048

	// EPOLLET в syscall имеет неудобный тип, так что завожу свою константу
	EPOLLET = 1 << 31
)

type (
	// Millisecond - тип для хранения времени в миллисекундах
	Millisecond int

	// EPoll реализует фукционал работы с одним epoll (и клиент и сервер)
	EPoll struct {
		fd    int
		event syscall.EpollEvent

		eventsCap      int
		events         []syscall.EpollEvent
		eventsFirstPtr uintptr

		WaitTimeout Millisecond
	}
)

var (
	// DefaultEPollWaitTimeout можно менять для изменения максимальной паузы при вызовах EPoll.Wait
	DefaultEPollWaitTimeout = Millisecond(10)
)

// InitClientEpoll настраивает новый клиентский epoll
func InitClientEpoll(epoll *EPoll) (err error) {
	epoll.fd, err = syscallWrappers.EpollCreate1(0)
	if err != nil {
		return err
	}

	epoll.WaitTimeout = DefaultEPollWaitTimeout

	epoll.eventsCap = maxEpollEvents
	epoll.events = make([]syscall.EpollEvent, maxEpollEvents)
	epoll.eventsFirstPtr = uintptr(unsafe.Pointer(&epoll.events[0]))

	return nil
}

// InitServerEpoll настраивает новый серверный epoll поверх слушающего сокета serverFd
func InitServerEpoll(serverFd int, epoll *EPoll) (err error) {
	if err = InitClientEpoll(epoll); err != nil {
		return err
	}

	epoll.event.Events = syscall.EPOLLIN | EPOLLET
	epoll.event.Fd = int32(serverFd)

	if err = syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, serverFd, &epoll.event); err != nil {
		_ = syscall.Close(epoll.fd)
		epoll.fd = 0
		return err
	}

	return nil
}

// DeleteFd удаляет дескриптор fd из пула
func (epoll *EPoll) DeleteFd(fd int) (err error) {
	return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

// AddClient добавляет нового клиента в серверный пул
func (epoll *EPoll) AddClient(clientFd int) (err error) {
	epoll.event.Events = syscall.EPOLLIN | EPOLLET // | syscall.EPOLLOUT
	epoll.event.Fd = int32(clientFd)

	if err = syscallWrappers.SetNonblock(clientFd, true); err != nil {
	} else if err = syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, clientFd, &epoll.event); err != nil {
	} else if err = syscallWrappers.SetsockoptInt(clientFd, syscall.SOL_TCP, syscall.TCP_NODELAY, 1); err != nil {
	} else if err = syscallWrappers.SetsockoptInt(clientFd, syscall.SOL_TCP, syscall.TCP_QUICKACK, 1); err != nil {
	} else {
		return nil
	}

	return err
}

// Wait блокируется до наступления события на любом из сокетов в пузе, либо на время epoll.WaitTimeout
func (epoll *EPoll) Wait() (nEvents int, errno syscall.Errno) {
	r1, _, errno := syscallWrappers.Syscall6(
		syscall.SYS_EPOLL_WAIT,
		uintptr(epoll.fd),
		epoll.eventsFirstPtr,
		uintptr(epoll.eventsCap),
		uintptr(epoll.WaitTimeout),
		0,
		0,
	)
	return int(r1), errno
}
