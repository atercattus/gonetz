package gonet

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"
)

func getSocketPort(fd int) int {
	if sa, err := syscall.Getsockname(fd); err != nil {
		return 0
	} else if sa4, ok := sa.(*syscall.SockaddrInet4); ok {
		return sa4.Port
	} else if sa6, ok := sa.(*syscall.SockaddrInet6); ok {
		return sa6.Port
	} else {
		return 0
	}
}

func Test_TCPServer_setupAcceptAddr(t *testing.T) {
	var srv TCPServer

	srv.setupAcceptAddr()
	if srv.acceptAddrPtr == 0 {
		t.Fatalf(`srv.acceptAddrPtr == 0`)
	}

	if srv.acceptAddrLen == 0 {
		t.Fatalf(`srv.acceptAddrLen == 0`)
	}

	if srv.acceptAddrLenPtr == 0 {
		t.Fatalf(`srv.acceptAddrLenPtr == 0`)
	}
}

func Test_TCPServer_makeListener_1(t *testing.T) {
	var srv TCPServer

	if err := srv.makeListener(`lol.kek`, 0); err == nil {
		t.Fatalf(`makeListener for wrong host was successfull`)
	} else if err != ErrWrongHost {
		t.Fatalf(`makeListener for wrong host returned wrong error`)
	}

	if err := srv.makeListener(``, 0); err != nil {
		t.Fatalf(`makeListener with empty host failed: %s`, err)
	}

	if err := srv.makeListener(`127.0.0.1`, 0); err != nil {
		t.Fatalf(`makeListener with 127.0.0.1 host failed: %s`, err)
	}
}

func Test_TCPServer_makeListener_2(t *testing.T) {
	var srv TCPServer

	SyscallWrappers.setWrongSocket()
	err := srv.makeListener(``, 0)
	SyscallWrappers.setRealSocket()
	if err == nil {
		t.Fatalf(`makeListener with wrong syscall.Socket was successfull`)
	}

	SyscallWrappers.setWrongSetNonblock()
	err = srv.makeListener(``, 0)
	SyscallWrappers.setRealSetNonblock()
	if err == nil {
		t.Fatalf(`makeListener with wrong syscall.SetNonblock was successfull`)
	}

	SyscallWrappers.setWrongSetsockoptInt(nil)
	err = srv.makeListener(``, 0)
	SyscallWrappers.setRealSetsockoptInt()
	if err == nil {
		t.Fatalf(`makeListener with wrong syscall.SetsockoptInt was successfull`)
	}

	SyscallWrappers.setWrongBind()
	err = srv.makeListener(``, 0)
	SyscallWrappers.setRealBind()
	if err == nil {
		t.Fatalf(`makeListener with wrong syscall.Bind was successfull`)
	}

	SyscallWrappers.setWrongListen()
	err = srv.makeListener(``, 0)
	SyscallWrappers.setRealListen()
	if err == nil {
		t.Fatalf(`makeListener with wrong syscall.Listen was successfull`)
	}
}

func Test_TCPServer_setupServerWorkers_1(t *testing.T) {
	var srv TCPServer

	if err := srv.setupServerWorkers(0); err == nil {
		t.Errorf(`setupServerWorkers successed with 0 pool size`)
		return
	}

	const (
		poolSize         = 1
		epollWaitTimeout = 10
	)
	DefaultEPollWaitTimeout = epollWaitTimeout // для проверки srv.workerPool ниже

	if err := srv.setupServerWorkers(poolSize); err != nil {
		t.Errorf(`setupServerWorkers failed: %s`, err)
		return
	}

	if l := len(srv.workerPool.epolls); l != poolSize {
		t.Errorf(`pool size after setupServerWorkers is wrong: expect %d got %d`, poolSize, l)
		return
	}

	if l := len(srv.workerPool.fds); l != poolSize {
		t.Errorf(`fds size after setupServerWorkers is wrong: expect %d got %d`, poolSize, l)
		return
	}

	fds := map[int]struct{}{}

	for _, fd := range srv.workerPool.fds {
		if fd == 0 {
			t.Errorf(`fd == 0 in pool`)
			return
		} else if _, ok := fds[fd]; ok {
			t.Errorf(`same fd for different worker pools`)
			return
		} else {
			fds[fd] = struct{}{}
		}
	}

	for _, epoll := range srv.workerPool.epolls {
		if epoll.fd == 0 {
			t.Errorf(`worker epoll is not initialized`)
			return
		}
	}

	// test event loop (primitive)
	clientFd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Errorf(`cannot create test socket: %s`, err)
		return
	}

	srv.workerPool.epolls[0].AddClient(clientFd)
	time.Sleep(10 * epollWaitTimeout * time.Millisecond)
	syscall.Syscall(syscall.SYS_CLOSE, uintptr(clientFd), 0, 0)
	time.Sleep(10 * epollWaitTimeout * time.Millisecond)

	//srv.Close()
}

func Test_TCPServer_setupServerWorkers_2(t *testing.T) {
	var srv TCPServer

	SyscallWrappers.setWrongEpollCreate1(nil)
	err := srv.setupServerWorkers(1)
	SyscallWrappers.setRealEpollCreate1()
	if err == nil {
		t.Errorf(`setupServerWorkers with wrong syscall.EpollCreate1 was successfull`)
		return
	}

	SyscallWrappers.setWrongEpollCreate1(
		CheckFuncSkipN(1, nil),
	)
	err = srv.setupServerWorkers(1)
	SyscallWrappers.setRealEpollCreate1()
	if err == nil {
		t.Errorf(`setupServerWorkers with wrong syscall.EpollCreate1 (skip 1) was successfull`)
		return
	}
}

func Test_TCPServer_accept(t *testing.T) {
	var srv TCPServer

	if clientFd, errno := srv.accept(); clientFd != -1 || errno == 0 {
		t.Errorf(`unexpected accept response for wrong call. clientFd:%d errno:%d`, clientFd, errno)
		return
	}

	// ToDo:
}

func Test_TCPServer_MakeServer_1(t *testing.T) {
	if _, err := MakeServer(`lol.kek`, 0); err == nil {
		t.Fatalf(`MakeServer didnt failed with wrong listen addr`)
	}
}

func Test_TCPServer_MakeServer_2(t *testing.T) {
	SyscallWrappers.setWrongEpollCreate1(
		CheckFuncSkipN(1, nil),
	)
	_, err := MakeServer(``, 0)
	SyscallWrappers.setRealEpollCreate1()
	if err == nil {
		t.Errorf(`MakeServer didnt failed with wrong syscall.EpollCreate1`)
	}
}

func Test_TCPServer_MakeServer_3(t *testing.T) {
	srv, err := MakeServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`MakeServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	port := getSocketPort(srv.fd)
	if port == 0 {
		t.Errorf(`Cannot determine test socket port`)
	}
}

func Test_TCPServer_MakeServer_4(t *testing.T) {
	call := 0
	SyscallWrappers.setWrongEpollCreate1(func(data interface{}) bool {
		call++
		return call < 2
	})
	_, err := MakeServer(`127.0.0.1`, 0)
	SyscallWrappers.setRealEpollCreate1()
	if err == nil {
		//srv.Close()
		t.Errorf(`Successfull MakeServer with wrong EpollCreate1`)
	}
}

func Test_TCPServer_Start_1(t *testing.T) {
	DefaultEPollWaitTimeout = 10

	SyscallWrappers.setWrongSyscall6(
		CheckFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 1, nil),
	)
	defer SyscallWrappers.setRealSyscall6()

	srv, err := MakeServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`MakeServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	timeLimiter := time.After(1 * time.Second)
	success := make(chan bool, 1)

	go func() {
		success <- srv.Start() == nil
	}()

	select {
	case <-timeLimiter:
		success <- false
	case succ := <-success:
		if succ {
			t.Errorf(`Successfull server start with wrong syscall.Syscall6`)
		}
	}
}

func Test_TCPServer_Start_2(t *testing.T) {
	srv, err := MakeServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`MakeServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	call := 0
	SyscallWrappers.Syscall6 = func(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		if call++; call < 2 {
			return 0, 0, syscall.EINTR
		}
		return 0, 0, syscall.EINVAL
	}
	err = srv.Start()
	SyscallWrappers.setRealSyscall6()
	if err == nil {
		t.Errorf(`Successfull Start with wrong Syscall6`)
		return
	}
}

func Test_TCPServer_Start_3(t *testing.T) {
	srv, err := MakeServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`MakeServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	call := 0
	SyscallWrappers.Syscall = func(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		if call++; call < 2 {
			return 0, 0, syscall.EAGAIN
		}
		return 0, 0, syscall.EINVAL
	}
	defer SyscallWrappers.setRealSyscall()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Errorf(`Timeout when server started`)
	}
}

func Test_TCPServer_Start_4(t *testing.T) {
	srv, err := MakeServer(`127.0.0.1`, 0)
	if err != nil {
		t.Fatalf(`MakeServer failed: %s`, err)
	}
	//defer srv.Close()

	port := getSocketPort(srv.fd)
	if port == 0 {
		t.Errorf(`Cannot determine test socket port`)
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)
	srv.OnClientRead(func(conn *TCPConn) bool {
		wg.Done()
		return true
	})

	SyscallWrappers.setWrongSetNonblock()
	defer SyscallWrappers.setRealSetNonblock()

	call := 0
	SyscallWrappers.Syscall = func(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		if trap != syscall.SYS_ACCEPT {
			return syscall.Syscall(trap, a1, a2, a3)
		} else if call++; call < 2 {
			return 0, 0, syscall.EAGAIN
		}
		return 0, 0, syscall.EBADF
	}
	defer SyscallWrappers.setRealSyscall()

	go func() {
		srv.Start()
		wg.Done()
	}()

	time.Sleep(100 * time.Millisecond) // для srv.Start()

	if client, err := net.DialTimeout(`tcp`, `127.0.0.1:`+strconv.Itoa(port), 1*time.Second); err != nil {
	} else {
		client.Write([]byte(`test data`))
		client.Close()
	}
	wg.Wait()
}

func Test_TCPServer_Start_5(t *testing.T) {
	srv, err := MakeServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`MakeServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	port := getSocketPort(srv.fd)
	if port == 0 {
		t.Errorf(`Cannot determine test socket port`)
		return
	}

	var wg sync.WaitGroup

	var readed bytes.Buffer
	var testData = make([]byte, 100)
	rand.Read(testData)

	wg.Add(1)
	srv.OnClientRead(func(conn *TCPConn) bool {
		readed.ReadFrom(conn)
		wg.Done()
		return true
	})

	timeLimiter := time.After(2 * time.Second)
	rdrWaiter := make(chan bool, 2)

	go func() {
		if err := srv.Start(); err != nil {
			rdrWaiter <- false
			t.Errorf(`Server start failed: %s (%s)`, err, time.Now())
		}
	}()

	go func() {
		time.Sleep(200 * time.Millisecond) // для srv.Start()

		if client, err := net.DialTimeout(`tcp`, `127.0.0.1:`+strconv.Itoa(port), 1*time.Second); err != nil {
			fmt.Println(err)
			rdrWaiter <- false
		} else {
			client.Write(testData)
			client.Close()
		}

		wg.Wait()

		rdrWaiter <- true
	}()

	select {
	case <-timeLimiter:
		t.Errorf(`Timelimit reached`)
	case succ := <-rdrWaiter:
		if !succ {
			// ok
		} else if got, exp := readed.Len(), len(testData); got != exp {
			t.Errorf(`Readed data len differs from sended. Expect:%d got:%d`, exp, got)
		} else if !bytes.Equal(readed.Bytes(), testData) {
			t.Errorf(`Readed data differs from sended. Expect: %q got: %q`, testData, readed.Bytes())
		}
	}
}

func Test_TCPServer_startWorkerLoop(t *testing.T) {
	var srv TCPServer

	SyscallWrappers.setWrongSyscall6(
		CheckFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 0, nil),
	)
	srv.setupServerWorkers(1)
	err := srv.startWorkerLoop(&srv.workerPool.epolls[0])
	SyscallWrappers.setRealSyscall6()
	if err == nil {
		t.Errorf(`setupServerWorkers with wrong syscall.Syscall6(syscall.SYS_EPOLL_WAIT) was successfull`)
	}
}
