package gonetz

import (
	"syscall"
	"testing"
)

func Test_SyscallWrappers_CheckFuncSkipN(t *testing.T) {
	cb := checkFuncSkipN(0, nil)
	for i := 0; i < 10; i++ {
		if cb(nil) != false {
			t.Errorf(`checkFuncSkipN(0) returns true`)
			break
		}
	}

	for skip := 1; skip <= 3; skip++ {
		cb := checkFuncSkipN(skip, nil)
		for i := 0; i < skip; i++ {
			if cb(nil) != false {
				t.Errorf(`checkFuncSkipN(%d) return true for %dth call`, skip, i)
				break
			}
		}
		for i := 0; i < 10; i++ {
			if cb(nil) != true {
				t.Errorf(`checkFuncSkipN(%d) returns false for non %dth calls`, skip, i)
				break
			}
		}
	}

	calls := 0
	nextCbData := int64(100500)
	nextCb := func(data interface{}) bool {
		if n, ok := data.(int64); !ok {
			t.Errorf(`Wrong nextCheck data type`)
			return false
		} else if n != nextCbData {
			t.Errorf(`Wrong nextCheck data value`)
			return false
		}

		calls++
		return true
	}

	cb1 := checkFuncSkipN(1, nextCb)
	cb1(nextCbData) // skip first call
	for i := 1; i <= 10; i++ {
		if cb1(nextCbData) != true {
			t.Errorf(`checkFuncSkipN() returns true`)
			break
		} else if calls != i {
			t.Errorf(`checkFuncSkipN() nextCheck call count differs from expected. Exp %d got %d`, i, calls)
			break
		}
	}
}

func Test_SyscallWrappers_CheckFuncSyscallTrapSkipN(t *testing.T) {
	dataCb := (interface{})([]uintptr{syscall.SYS_EPOLL_WAIT})

	if cb := checkFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 0, nil); true {
		for i := 0; i < 10; i++ {
			if cb(dataCb) != false {
				t.Errorf(`checkFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 0) returns true`)
				break
			}
		}
	}

	if cb := checkFuncSyscallTrapSkipN(1, 0, nil); true {
		for i := 0; i < 10; i++ {
			if cb(dataCb) != true {
				t.Errorf(`checkFuncSyscallTrapSkipN(1, 0) returns false`)
				break
			}
		}
	}

	for skip := 1; skip <= 3; skip++ {
		cb := checkFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, skip, nil)
		for i := 0; i < skip; i++ {
			if cb(dataCb) != false {
				t.Errorf(`checkFuncSyscallTrapSkipN(%d) return true for %dth call`, skip, i)
				break
			}
		}
		for i := 0; i < 10; i++ {
			if cb(dataCb) != true {
				t.Errorf(`checkFuncSyscallTrapSkipN(%d) returns false for non %dth calls`, skip, i)
				break
			}
		}
	}

	calls := 0
	nextCb := func(data interface{}) bool {
		if d, ok := data.([]uintptr); !ok {
			t.Errorf(`Wrong nextCheck data type`)
			return false
		} else if d[0] != dataCb.([]uintptr)[0] {
			t.Errorf(`Wrong nextCheck data value`)
			return false
		}

		calls++
		return true
	}

	cb1 := checkFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 1, nextCb)
	cb1(dataCb) // skip first call
	for i := 1; i <= 10; i++ {
		if cb1(dataCb) != true {
			t.Errorf(`checkFuncSyscallTrapSkipN() returns true`)
			break
		} else if calls != i {
			t.Errorf(`checkFuncSyscallTrapSkipN() nextCheck call count differs from expected. Exp %d got %d`, i, calls)
			break
		}
	}
}

func Test_SyscallWrappers_EpollCreate1(t *testing.T) {
	syscallWrappers.setWrongEpollCreate1(nil)
	_, errno := syscallWrappers.EpollCreate1(0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	syscallWrappers.setRealEpollCreate1()

	syscallWrappers.setWrongEpollCreate1(checkFuncSkipN(0, nil))
	_, errno = syscallWrappers.EpollCreate1(0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno (with checkFunc)`)
	}
	syscallWrappers.setRealEpollCreate1()

	syscallWrappers.setWrongEpollCreate1(func(data interface{}) bool {
		return true
	})
	_, errno = syscallWrappers.EpollCreate1(0)
	if errno != nil {
		t.Errorf(`Wrong errno (custom checkFunc)`)
	}
	syscallWrappers.setRealEpollCreate1()
}

func Test_SyscallWrappers_Socket(t *testing.T) {
	syscallWrappers.setWrongSocket()
	_, errno := errorableSyscallWrappers.Socket(0, 0, 0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	syscallWrappers.setRealSocket()
}

func Test_SyscallWrappers_SetNonblock(t *testing.T) {
	syscallWrappers.setWrongSetNonblock()
	errno := errorableSyscallWrappers.SetNonblock(0, true)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	syscallWrappers.setRealSetNonblock()
}

func Test_SyscallWrappers_SetsockoptInt(t *testing.T) {
	syscallWrappers.setWrongSetsockoptInt(nil)
	errno := errorableSyscallWrappers.SetsockoptInt(0, 0, 0, 0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	syscallWrappers.setRealSetsockoptInt()

	syscallWrappers.setWrongSetsockoptInt(checkFuncSkipN(0, nil))
	errno = errorableSyscallWrappers.SetsockoptInt(0, 0, 0, 0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno (with checkFunc)`)
	}
	syscallWrappers.setRealSetsockoptInt()

	syscallWrappers.setWrongSetsockoptInt(func(data interface{}) bool {
		return true
	})
	errno = errorableSyscallWrappers.SetsockoptInt(0, 0, 0, 0)
	if errno != syscall.ENOTSOCK {
		t.Errorf(`Wrong errno (custom checkFunc)`)
	}
	syscallWrappers.setRealSetsockoptInt()
}

func Test_SyscallWrappers_Bind(t *testing.T) {
	syscallWrappers.setWrongBind()
	errno := errorableSyscallWrappers.Bind(0, nil)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	syscallWrappers.setRealBind()
}

func Test_SyscallWrappers_Listen(t *testing.T) {
	syscallWrappers.setWrongListen()
	errno := errorableSyscallWrappers.Listen(0, 0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	syscallWrappers.setRealListen()
}

func Test_SyscallWrappers_Syscall(t *testing.T) {
	syscallWrappers.setWrongSyscall(nil)
	_, _, errno := errorableSyscallWrappers.Syscall(1, 2, 3, 4)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	syscallWrappers.setRealSyscall()

	syscallWrappers.setWrongSyscall(checkFuncSkipN(0, nil))
	_, _, errno = errorableSyscallWrappers.Syscall(1, 2, 3, 4)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno (with checkFunc)`)
	}
	syscallWrappers.setRealSyscall()

	syscallWrappers.setWrongSyscall(func(data interface{}) bool {
		return true
	})
	_, _, errno = errorableSyscallWrappers.Syscall(1, 2, 3, 4)
	if errno != syscall.EFAULT {
		t.Errorf(`Wrong errno (custom checkFunc)`)
	}
	syscallWrappers.setRealSyscall()
}

func Test_SyscallWrappers_Syscall6(t *testing.T) {
	syscallWrappers.setWrongSyscall6(nil)
	_, _, errno := errorableSyscallWrappers.Syscall6(1, 2, 3, 4, 5, 6, 7)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	syscallWrappers.setRealSyscall6()

	syscallWrappers.setWrongSyscall6(checkFuncSkipN(0, nil))
	_, _, errno = errorableSyscallWrappers.Syscall6(1, 2, 3, 4, 5, 6, 7)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno (with checkFunc)`)
	}
	syscallWrappers.setRealSyscall6()

	syscallWrappers.setWrongSyscall6(func(data interface{}) bool {
		return true
	})
	_, _, errno = errorableSyscallWrappers.Syscall6(1, 2, 3, 4, 5, 6, 7)
	if errno != syscall.EFAULT {
		t.Errorf(`Wrong errno (custom checkFunc)`)
	}
	syscallWrappers.setRealSyscall6()
}
