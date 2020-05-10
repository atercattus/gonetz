package gonet

import (
	"bytes"
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
	if srv.acceptAddr.Ptr == 0 {
		t.Fatalf(`srv.acceptAddrPtr == 0`)
	}
	if srv.acceptAddr.Len == 0 {
		t.Fatalf(`srv.acceptAddrLen == 0`)
	}

	if srv.acceptAddr.LenPtr == 0 {
		t.Fatalf(`srv.acceptAddrLenPtr == 0`)
	}
}

func Test_TCPServer_newListenerIPv4_1(t *testing.T) {
	var srv TCPServer

	if err := srv.newListenerIPv4(`lol.kek`, 0); err == nil {
		t.Fatalf(`newListenerIPv4 for wrong host was successfull`)
	} else if err != ErrWrongHost {
		t.Fatalf(`newListenerIPv4 for wrong host returned wrong error`)
	}

	if err := srv.newListenerIPv4(``, 0); err != nil {
		t.Fatalf(`newListenerIPv4 with empty host failed: %s`, err)
	}

	if err := srv.newListenerIPv4(`127.0.0.1`, 0); err != nil {
		t.Fatalf(`newListenerIPv4 with 127.0.0.1 host failed: %s`, err)
	}
}

func Test_TCPServer_newListenerIPv4_2(t *testing.T) {
	var srv TCPServer

	syscallWrappers.setWrongSocket()
	err := srv.newListenerIPv4(``, 0)
	syscallWrappers.setRealSocket()
	if err == nil {
		t.Fatalf(`newListenerIPv4 with wrong syscall.Socket was successfull`)
	}

	syscallWrappers.setWrongSetNonblock()
	err = srv.newListenerIPv4(``, 0)
	syscallWrappers.setRealSetNonblock()
	if err == nil {
		t.Fatalf(`newListenerIPv4 with wrong syscall.SetNonblock was successfull`)
	}

	syscallWrappers.setWrongSetsockoptInt(nil)
	err = srv.newListenerIPv4(``, 0)
	syscallWrappers.setRealSetsockoptInt()
	if err == nil {
		t.Fatalf(`newListenerIPv4 with wrong syscall.SetsockoptInt was successfull`)
	}

	syscallWrappers.setWrongBind()
	err = srv.newListenerIPv4(``, 0)
	syscallWrappers.setRealBind()
	if err == nil {
		t.Fatalf(`newListenerIPv4 with wrong syscall.Bind was successfull`)
	}

	syscallWrappers.setWrongListen()
	err = srv.newListenerIPv4(``, 0)
	syscallWrappers.setRealListen()
	if err == nil {
		t.Fatalf(`newListenerIPv4 with wrong syscall.Listen was successfull`)
	}
}

func Test_TCPServer_setupServerWorkers_1(t *testing.T) {
	var srv TCPServer
	//defer srv.Close()

	if err := srv.setupServerWorkers(0); err == nil {
		t.Errorf(`setupServerWorkers succeeded with 0 pool size`)
		return
	}

	const (
		poolSize         = 1
		epollWaitTimeout = 10
	)

	var bakDefaultEPollWaitTimeout = DefaultEPollWaitTimeout
	DefaultEPollWaitTimeout = epollWaitTimeout // для проверки srv.workerPool ниже
	defer func() {
		DefaultEPollWaitTimeout = bakDefaultEPollWaitTimeout
	}()

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

	if err := srv.workerPool.epolls[0].AddClient(clientFd); err != nil {
		t.Errorf(`cannot add client to pool: %s`, err)
		return
	}
	time.Sleep(10 * (epollWaitTimeout * time.Millisecond))
	_, _, _ = syscall.Syscall(syscall.SYS_CLOSE, uintptr(clientFd), 0, 0)
	time.Sleep(10 * (epollWaitTimeout * time.Millisecond))
}

func Test_TCPServer_setupServerWorkers_2(t *testing.T) {
	var srv TCPServer

	syscallWrappers.setWrongEpollCreate1(nil)
	err := srv.setupServerWorkers(1)
	syscallWrappers.setRealEpollCreate1()
	if err == nil {
		t.Errorf(`setupServerWorkers with wrong syscall.EpollCreate1 was successfull`)
		return
	}

	syscallWrappers.setWrongEpollCreate1(
		CheckFuncSkipN(1, nil),
	)
	err = srv.setupServerWorkers(1)
	syscallWrappers.setRealEpollCreate1()
	if err == nil {
		t.Errorf(`setupServerWorkers with wrong syscall.EpollCreate1 (skip 1) was successfull`)
		return
	}
}

func Test_TCPServer_accept(t *testing.T) {
	var srv TCPServer

	if clientFd, errno := srv.accept(); (clientFd != -1) || (errno == 0) {
		t.Errorf(`unexpected accept response for wrong call. clientFd:%d errno:%d`, clientFd, errno)
		return
	}

	// Полноценные тесты в рамках тестов Test_TCPServer_Start_*
}

func Test_TCPServer_NewServer_1(t *testing.T) {
	if _, err := NewServer(`lol.kek`, 0); err == nil {
		t.Fatalf(`NewServer didnt failed with wrong listen addr`)
	}
}

func Test_TCPServer_NewServer_2(t *testing.T) {
	syscallWrappers.setWrongEpollCreate1(
		CheckFuncSkipN(1, nil),
	)
	_, err := NewServer(``, 0)
	syscallWrappers.setRealEpollCreate1()
	if err == nil {
		t.Errorf(`NewServer didnt failed with wrong syscall.EpollCreate1`)
	}
}

func Test_TCPServer_NewServer_3(t *testing.T) {
	srv, err := NewServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`NewServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	port := getSocketPort(srv.fd)
	if port == 0 {
		t.Errorf(`Cannot determine test socket port`)
	}
}

func Test_TCPServer_NewServer_4(t *testing.T) {
	call := 0
	syscallWrappers.setWrongEpollCreate1(func(data interface{}) bool {
		call++
		return call <= 1 // "Ломаю" только первый вызов
	})
	_, err := NewServer(`127.0.0.1`, 0)
	syscallWrappers.setRealEpollCreate1()
	if err == nil {
		//srv.Close()
		t.Errorf(`Successfull NewServer with wrong EpollCreate1`)
	}
}

func Test_TCPServer_Start_1(t *testing.T) {
	var bakDefaultEPollWaitTimeout = DefaultEPollWaitTimeout
	DefaultEPollWaitTimeout = 10
	defer func() {
		DefaultEPollWaitTimeout = bakDefaultEPollWaitTimeout
	}()

	syscallWrappers.setWrongSyscall6(
		CheckFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 1, nil),
	)
	defer syscallWrappers.setRealSyscall6()

	srv, err := NewServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`NewServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	timeout := time.Duration(DefaultEPollWaitTimeout) * time.Millisecond
	timeLimiter := time.After(timeout * 2) // x2 запас на время реакции
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
	srv, err := NewServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`NewServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	call := 0
	syscallWrappers.Syscall6 = func(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		if call++; call <= 1 {
			return 0, 0, syscall.EINTR
		}
		return 0, 0, syscall.EINVAL
	}
	err = srv.Start()
	syscallWrappers.setRealSyscall6()
	if err == nil {
		t.Errorf(`Successfull Start with wrong Syscall6`)
		return
	}
}

func Test_TCPServer_Start_3(t *testing.T) {
	srv, err := NewServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`NewServer failed: %s`, err)
		return
	}
	//defer srv.Close()

	call := 0
	syscallWrappers.Syscall = func(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		if call++; call <= 1 {
			return 0, 0, syscall.EAGAIN
		}
		return 0, 0, syscall.EINVAL
	}
	defer syscallWrappers.setRealSyscall()

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

// Тест на базовый коннект и заглушку srv.rdEvent
func Test_TCPServer_Start_4(t *testing.T) {
	srv, err := NewServer(`127.0.0.1`, 0)
	if err != nil {
		t.Fatalf(`NewServer failed: %s`, err)
	}
	//defer srv.Close()

	port := getSocketPort(srv.fd)
	if port == 0 {
		t.Errorf(`Cannot determine test socket port`)
		return
	}

	go func() {
		if err := srv.Start(); err != nil {
			t.Errorf(`Could not Start: %s`, err)
		}
	}()

	time.Sleep(100 * time.Millisecond) // для srv.Start()

	if client, err := net.DialTimeout(`tcp`, `127.0.0.1:`+strconv.Itoa(port), 1*time.Second); err != nil {
		t.Errorf(`Could not dial to server: %s`, err)
	} else {
		if _, err := client.Write([]byte(`test data`)); err != nil {
			t.Errorf(`Could not write to client: %s`, err)
		}
		_ = client.Close()
	}

	time.Sleep(1 * time.Second)
	srv.Close()
}

// Тест на TCPServer.startWorkerLoop syscall.EPOLLIN => (errno != syscall.EAGAIN)
func Test_TCPServer_Start_5(t *testing.T) {
	srv, err := NewServer(`127.0.0.1`, 0)
	if err != nil {
		t.Fatalf(`NewServer failed: %s`, err)
	}

	port := getSocketPort(srv.fd)
	if port == 0 {
		t.Errorf(`Cannot determine test socket port`)
		return
	}

	go func() {
		if err := srv.Start(); err != nil {
			t.Errorf(`Could not Start: %s`, err)
		}
	}()

	time.Sleep(100 * time.Millisecond) // для srv.Start()

	call := 0
	syscallWrappers.Syscall = func(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		if trap != syscall.SYS_READ {
			return syscall.Syscall(trap, a1, a2, a3)
		} else if call++; call <= 1 {
			return 0, 0, syscall.EAGAIN
		}
		return 0, 0, syscall.EBADF
	}
	defer syscallWrappers.setRealSyscall()

	var errGot error

	if client, err := net.DialTimeout(`tcp`, `127.0.0.1:`+strconv.Itoa(port), 1*time.Second); err != nil {
		errGot = err
	} else {
		if _, err := client.Write([]byte(`test data`)); err != nil {
			errGot = err
		} else {
			time.Sleep(100 * time.Millisecond) // Чтобы сервер успел получить данные
			_ = client.Close()
		}
	}

	if errGot == nil {
		t.Errorf(`Successfull client logic with wrong Syscall(SYS_READ)`)
	}
}

// Тест на полноценную обработку запроса сервером
func Test_TCPServer_Start_6(t *testing.T) {
	srv, err := NewServer(`127.0.0.1`, 0)
	if err != nil {
		t.Errorf(`NewServer failed: %s`, err)
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
		if _, err := readed.ReadFrom(conn); err != nil {
			t.Errorf(`Could not read client data: %s`, err.Error())
		}
		wg.Done()
		return true
	})

	timeLimiter := time.After(2 * time.Second)
	rdrWaiter := make(chan bool, 2)

	go func() {
		if err := srv.Start(); err != nil {
			t.Errorf(`Server start failed: %s (%s)`, err, time.Now())
			rdrWaiter <- false
		}
	}()

	go func() {
		time.Sleep(200 * time.Millisecond) // для srv.Start()

		if client, err := net.DialTimeout(`tcp`, `127.0.0.1:`+strconv.Itoa(port), 1*time.Second); err != nil {
			t.Errorf(`Could not dial to server: %s`, err)
			rdrWaiter <- false
		} else {
			if _, err := client.Write(testData); err != nil {
				t.Errorf(`Could not write to client: %s`, err)
			}
			_ = client.Close()
		}

		wg.Wait()

		rdrWaiter <- true
	}()

	select {
	case <-timeLimiter:
		t.Errorf(`Timelimit reached`)
	case succ := <-rdrWaiter:
		if !succ {
			// ошибка уже возвращена
		} else if got, exp := readed.Len(), len(testData); got != exp {
			t.Errorf(`Readed data len differs from sended. Expect:%d got:%d`, exp, got)
		} else if !bytes.Equal(readed.Bytes(), testData) {
			t.Errorf(`Readed data differs from sended. Expect: %q got: %q`, testData, readed.Bytes())
		}
	}
}

func Test_TCPServer_startWorkerLoop(t *testing.T) {
	var srv TCPServer

	syscallWrappers.setWrongSyscall6(
		CheckFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 0, nil),
	)
	defer syscallWrappers.setRealSyscall6()

	if err := srv.setupServerWorkers(1); err != nil {
		t.Errorf(`setupServerWorkers was failed: %s`, err)
	}

	if err := srv.startWorkerLoop(&srv.workerPool.epolls[0]); err == nil {
		t.Errorf(`startWorkerLoop with wrong syscall.Syscall6(syscall.SYS_EPOLL_WAIT) was successfull`)
	}
}
