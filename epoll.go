package gonet

import (
	"syscall"
	"unsafe"
)

const (
	maxEpollEvents = 2048
	EPOLLET        = 1 << 31 // syscall.EPOLLET has wrong type
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

	syscallWrapperFuncs struct {
		EpollCreate1 func(flag int) (fd int, err error)
	}
)

var (
	DefaultEPollWaitTimeout = -1
)

var (
	// Врапперы над функциями для возможности потестировать сбои в работе этих вызовов
	defaultSyscallWrappers = syscallWrapperFuncs{
		EpollCreate1: syscall.EpollCreate1,
	}
	syscallWrappers = defaultSyscallWrappers
)

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

func InitServerEpoll(serverFd int, epoll *EPoll) (err error) {
	if err = InitClientEpoll(epoll); err != nil {
		return err
	}

	epoll.event.Events = syscall.EPOLLIN | EPOLLET
	epoll.event.Fd = int32(serverFd)

	if err = syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, serverFd, &epoll.event); err != nil {
		syscall.Close(epoll.fd)
		epoll.fd = 0
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
