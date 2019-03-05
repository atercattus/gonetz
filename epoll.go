package gonet

import (
	"syscall"
	"unsafe"
)

const (
	EPOLLET = 1 << 31 // в stdlib syscall.EPOLLET идет не того типа, как хотелось бы
)

type (
	EPoll struct {
		fd    int
		event syscall.EpollEvent

		eventsCap      int
		events         []syscall.EpollEvent
		eventsFirstPtr uintptr

		WaitTimeout int
	}
)

func InitClientEpoll(eventsCap int, epoll *EPoll) (err error) {
	epoll.fd, err = syscall.EpollCreate1(0)
	if err != nil {
		return err
	}

	epoll.WaitTimeout = -1

	epoll.eventsCap = eventsCap
	epoll.events = make([]syscall.EpollEvent, eventsCap)
	epoll.eventsFirstPtr = uintptr(unsafe.Pointer(&epoll.events[0]))

	return nil
}

func InitServerEpoll(serverFd int, eventsCap int, epoll *EPoll) (err error) {
	if err = InitClientEpoll(eventsCap, epoll); err != nil {
		return err
	}

	epoll.event.Events = syscall.EPOLLIN | EPOLLET
	epoll.event.Fd = int32(serverFd)

	if err = syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, serverFd, &epoll.event); err != nil {
		syscall.Close(epoll.fd)
		return err
	}

	return nil
}

func (epoll *EPoll) DeleteFd(fd int) (err error) {
	return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (epoll *EPoll) AddClient(clientFd int) (err error) {
	epoll.event.Events = syscall.EPOLLIN | EPOLLET // | syscall.EPOLLOUT
	epoll.event.Fd = int32(clientFd)

	if err = syscall.SetNonblock(clientFd, true); err != nil {
	} else if err = syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, clientFd, &epoll.event); err != nil {
	} else if err = syscall.SetsockoptInt(clientFd, syscall.SOL_TCP, syscall.TCP_NODELAY, 1); err != nil {
	} else if err = syscall.SetsockoptInt(clientFd, syscall.SOL_TCP, syscall.TCP_QUICKACK, 1); err != nil {
	} else {
		// all ok
		return nil
	}

	//} else if err = syscall.SetsockoptInt(serverFd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
	//} else if err = syscall.SetsockoptInt(serverFd, syscall.SOL_SOCKET, SO_REUSEPORT, 1); err != nil {

	//syscall.Close(clientFd)
	return err
}

func (epoll *EPoll) Wait() (nEvents int, errno syscall.Errno) {
	r1, _, errno := syscall.Syscall6(
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
