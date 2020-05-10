package gonet

import (
	"syscall"
	"testing"
)

func Test_EPoll_InitClientEpoll(t *testing.T) {
	var epoll EPoll

	if err := InitClientEpoll(&epoll); err != nil {
		t.Fatalf(`InitClientEpoll failed: %s`, err)
	}

	if epoll.fd == 0 {
		t.Fatalf(`epoll.fd == 0`)
	}

	if epoll.eventsCap != len(epoll.events) {
		t.Fatalf(`epoll.eventsCap != len(epoll.events): %d vs %d`, epoll.eventsCap, len(epoll.events))
	}

	if epoll.eventsFirstPtr == 0 {
		t.Fatalf(`epoll.eventsFirstPtr == 0`)
	}

	// Проверка на ошибку
	syscallWrappers.setWrongEpollCreate1(nil)
	err := InitClientEpoll(&epoll)
	syscallWrappers.setRealEpollCreate1()
	if err == nil {
		t.Fatalf(`InitClientEpoll didnt failed with wrong EpollCreate1`)
	}
}

func Test_EPoll_InitServerEpoll(t *testing.T) {
	var (
		epoll    EPoll
		serverFd int
		err      error
	)

	serverFd, err = syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf(`cannot create test socket: %s`, err)
	}
	defer syscall.Syscall(syscall.SYS_CLOSE, uintptr(serverFd), 0, 0)

	if err := InitServerEpoll(serverFd, &epoll); err != nil {
		t.Fatalf(`InitServerEpoll failed: %s`, err)
	}

	if epoll.fd == 0 {
		t.Fatalf(`epoll.fd == 0`)
	}

	// Проверка на ошибку
	syscallWrappers.setWrongEpollCreate1(nil)
	err = InitServerEpoll(serverFd, &epoll)
	syscallWrappers.setRealEpollCreate1()
	if err == nil {
		t.Fatalf(`InitServerEpoll didnt failed with wrong EpollCreate1`)
	}

	// Проверка на закрытие дескриптора
	syscall.Syscall(syscall.SYS_CLOSE, uintptr(serverFd), 0, 0)

	if err := InitServerEpoll(serverFd, &epoll); err == nil {
		t.Fatalf(`InitServerEpoll didnt failed after socket closing`)
	}

	if epoll.fd != 0 {
		t.Fatalf(`epoll.fd != 0 after InitServerEpoll failure`)
	}
}

func Test_EPoll_AddClient1(t *testing.T) {
	var (
		epoll    EPoll
		clientFd int
		err      error
	)

	clientFd, err = syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf(`cannot create test socket: %s`, err)
	}
	defer syscall.Syscall(syscall.SYS_CLOSE, uintptr(clientFd), 0, 0)

	if err := InitClientEpoll(&epoll); err != nil {
		t.Fatalf(`InitClientEpoll failed: %s`, err)
	}

	if err := epoll.AddClient(clientFd); err != nil {
		t.Fatalf(`AddClient failed: %s`, err)
	}

	if err := epoll.AddClient(clientFd); err == nil {
		t.Fatalf(`double call of AddClient with same clientFd succeeded`)
	}

	if err := epoll.AddClient(0); err == nil {
		t.Fatalf(`wrong fd for AddClient succeeded`)
	}

	syscall.Syscall(syscall.SYS_CLOSE, uintptr(clientFd), 0, 0)
	if err := epoll.AddClient(clientFd); err == nil {
		t.Fatalf(`AddClient didnt failed after socket closing`)
	}
}

func Test_EPoll_AddClient2(t *testing.T) {
	var (
		epoll    EPoll
		clientFd int
		err      error
	)

	clientFd, err = syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf(`cannot create test socket: %s`, err)
	}
	defer syscall.Syscall(syscall.SYS_CLOSE, uintptr(clientFd), 0, 0)

	syscallWrappers.setWrongSetsockoptInt(nil)
	err = epoll.AddClient(clientFd)
	syscallWrappers.setRealSetsockoptInt()
	if err == nil {
		t.Errorf(`Successful AddClient with wrong SetsockoptInt`)
		return
	}
}

func Test_EPoll_AddClient3(t *testing.T) {
	var (
		epoll    EPoll
		clientFd int
		err      error
	)

	clientFd, err = syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf(`cannot create test socket: %s`, err)
	}
	defer syscall.Syscall(syscall.SYS_CLOSE, uintptr(clientFd), 0, 0)

	syscallWrappers.setWrongSetsockoptInt(func(data interface{}) bool {
		if ints, ok := data.([]int); ok && len(ints) > 3 {
			return ints[2] != syscall.TCP_QUICKACK
		}
		return true
	})
	err = epoll.AddClient(clientFd)
	syscallWrappers.setRealSetsockoptInt()
	if err == nil {
		t.Errorf(`Successful AddClient with wrong SetsockoptInt`)
		return
	}
}

func Test_EPoll_DeleteFd(t *testing.T) {
	var (
		epoll    EPoll
		clientFd int
		err      error
	)

	clientFd, err = syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf(`cannot create test socket: %s`, err)
	}
	defer syscall.Syscall(syscall.SYS_CLOSE, uintptr(clientFd), 0, 0)

	if err := InitClientEpoll(&epoll); err != nil {
		t.Fatalf(`InitClientEpoll failed: %s`, err)
	}

	if err := epoll.AddClient(clientFd); err != nil {
		t.Fatalf(`AddClient failed: %s`, err)
	}

	if err := epoll.DeleteFd(clientFd); err != nil {
		t.Fatalf(`DeleteFd failed: %s`, err)
	}

	if err := epoll.DeleteFd(clientFd); err == nil {
		t.Fatalf(`double call of DeleteFd didnt failed`)
	}

	if err := epoll.AddClient(clientFd); err != nil {
		t.Fatalf(`Second AddClient failed: %s`, err)
	}

	if err := epoll.DeleteFd(clientFd); err != nil {
		t.Fatalf(`Second DeleteFd failed: %s`, err)
	}
}

func Test_EPoll_Wait(t *testing.T) {
	var (
		epoll    EPoll
		clientFd int
		err      error
	)

	clientFd, err = syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf(`cannot create test socket: %s`, err)
	}
	defer syscall.Syscall(syscall.SYS_CLOSE, uintptr(clientFd), 0, 0)

	if err := InitClientEpoll(&epoll); err != nil {
		t.Fatalf(`InitClientEpoll failed: %s`, err)
	}

	if err := epoll.AddClient(clientFd); err != nil {
		t.Fatalf(`AddClient failed: %s`, err)
	}

	epoll.WaitTimeout = Millisecond(1)
	n, errno := epoll.Wait()
	if (n != 1) || (errno != 0) {
		t.Fatalf(`Wait failed: nEvents=%d errno=%d`, n, errno)
	}

	ev := epoll.events[0]

	if int(ev.Fd) != clientFd {
		t.Fatalf(`event fd != test fd: %d vs %d`, ev.Fd, clientFd)
	}
	if ev.Events != syscall.EPOLLHUP {
		t.Fatalf(`events mask (%d) != syscall.EPOLLHUP (%d)`, ev.Events, syscall.EPOLLHUP)
	}
}
