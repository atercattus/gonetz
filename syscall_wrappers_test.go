package gonet

import (
	"syscall"
	"testing"
)

func Test_SyscallWrappers_CheckFuncSkipN(t *testing.T) {
	cb := CheckFuncSkipN(0, nil)
	for i := 0; i < 10; i++ {
		if cb(nil) != false {
			t.Errorf(`CheckFuncSkipN(0) returns true`)
			break
		}
	}

	for skip := 1; skip <= 3; skip++ {
		cb := CheckFuncSkipN(skip, nil)
		for i := 0; i < skip; i++ {
			if cb(nil) != false {
				t.Errorf(`CheckFuncSkipN(%d) return true for %dth call`, skip, i)
				break
			}
		}
		for i := 0; i < 10; i++ {
			if cb(nil) != true {
				t.Errorf(`CheckFuncSkipN(%d) returns false for non %dth calls`, skip, i)
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

	cb1 := CheckFuncSkipN(1, nextCb)
	cb1(nextCbData) // skip first call
	for i := 1; i <= 10; i++ {
		if cb1(nextCbData) != true {
			t.Errorf(`CheckFuncSkipN() returns true`)
			break
		} else if calls != i {
			t.Errorf(`CheckFuncSkipN() nextCheck call count differs from expected. Exp %d got %d`, i, calls)
			break
		}
	}
}

func Test_SyscallWrappers_CheckFuncSyscallTrapSkipN(t *testing.T) {
	dataCb := (interface{})([]uintptr{syscall.SYS_EPOLL_WAIT})

	if cb := CheckFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 0, nil); true {
		for i := 0; i < 10; i++ {
			if cb(dataCb) != false {
				t.Errorf(`CheckFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 0) returns true`)
				break
			}
		}
	}

	if cb := CheckFuncSyscallTrapSkipN(1, 0, nil); true {
		for i := 0; i < 10; i++ {
			if cb(dataCb) != true {
				t.Errorf(`CheckFuncSyscallTrapSkipN(1, 0) returns false`)
				break
			}
		}
	}

	for skip := 1; skip <= 3; skip++ {
		cb := CheckFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, skip, nil)
		for i := 0; i < skip; i++ {
			if cb(dataCb) != false {
				t.Errorf(`CheckFuncSyscallTrapSkipN(%d) return true for %dth call`, skip, i)
				break
			}
		}
		for i := 0; i < 10; i++ {
			if cb(dataCb) != true {
				t.Errorf(`CheckFuncSyscallTrapSkipN(%d) returns false for non %dth calls`, skip, i)
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

	cb1 := CheckFuncSyscallTrapSkipN(syscall.SYS_EPOLL_WAIT, 1, nextCb)
	cb1(dataCb) // skip first call
	for i := 1; i <= 10; i++ {
		if cb1(dataCb) != true {
			t.Errorf(`CheckFuncSyscallTrapSkipN() returns true`)
			break
		} else if calls != i {
			t.Errorf(`CheckFuncSyscallTrapSkipN() nextCheck call count differs from expected. Exp %d got %d`, i, calls)
			break
		}
	}
}

func Test_SyscallWrappers_EpollCreate1(t *testing.T) {
	SyscallWrappers.setWrongEpollCreate1(nil)
	_, errno := SyscallWrappers.EpollCreate1(0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	SyscallWrappers.setRealEpollCreate1()

	SyscallWrappers.setWrongEpollCreate1(CheckFuncSkipN(0, nil))
	_, errno = SyscallWrappers.EpollCreate1(0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno (with checkFunc)`)
	}
	SyscallWrappers.setRealEpollCreate1()

	SyscallWrappers.setWrongEpollCreate1(func(data interface{}) bool {
		return true
	})
	_, errno = SyscallWrappers.EpollCreate1(0)
	if errno != nil {
		t.Errorf(`Wrong errno (custom checkFunc)`)
	}
	SyscallWrappers.setRealEpollCreate1()
}

func Test_SyscallWrappers_Socket(t *testing.T) {
	SyscallWrappers.setWrongSocket()
	_, errno := errorableSyscallWrappers.Socket(0, 0, 0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	SyscallWrappers.setRealSocket()
}

func Test_SyscallWrappers_SetNonblock(t *testing.T) {
	SyscallWrappers.setWrongSetNonblock()
	errno := errorableSyscallWrappers.SetNonblock(0, true)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	SyscallWrappers.setRealSetNonblock()
}

func Test_SyscallWrappers_SetsockoptInt(t *testing.T) {
	SyscallWrappers.setWrongSetsockoptInt(nil)
	errno := errorableSyscallWrappers.SetsockoptInt(0, 0, 0, 0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	SyscallWrappers.setRealSetsockoptInt()

	SyscallWrappers.setWrongSetsockoptInt(CheckFuncSkipN(0, nil))
	errno = errorableSyscallWrappers.SetsockoptInt(0, 0, 0, 0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno (with checkFunc)`)
	}
	SyscallWrappers.setRealSetsockoptInt()

	SyscallWrappers.setWrongSetsockoptInt(func(data interface{}) bool {
		return true
	})
	errno = errorableSyscallWrappers.SetsockoptInt(0, 0, 0, 0)
	if errno != syscall.ENOTSOCK {
		t.Errorf(`Wrong errno (custom checkFunc)`)
	}
	SyscallWrappers.setRealSetsockoptInt()
}

func Test_SyscallWrappers_Bind(t *testing.T) {
	SyscallWrappers.setWrongBind()
	errno := errorableSyscallWrappers.Bind(0, nil)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	SyscallWrappers.setRealBind()
}

func Test_SyscallWrappers_Listen(t *testing.T) {
	SyscallWrappers.setWrongListen()
	errno := errorableSyscallWrappers.Listen(0, 0)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	SyscallWrappers.setRealListen()
}

func Test_SyscallWrappers_Syscall(t *testing.T) {
	SyscallWrappers.setWrongSyscall(nil)
	_, _, errno := errorableSyscallWrappers.Syscall(1, 2, 3, 4)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	SyscallWrappers.setRealSyscall()

	SyscallWrappers.setWrongSyscall(CheckFuncSkipN(0, nil))
	_, _, errno = errorableSyscallWrappers.Syscall(1, 2, 3, 4)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno (with checkFunc)`)
	}
	SyscallWrappers.setRealSyscall()

	SyscallWrappers.setWrongSyscall(func(data interface{}) bool {
		return true
	})
	_, _, errno = errorableSyscallWrappers.Syscall(1, 2, 3, 4)
	if errno != syscall.EFAULT {
		t.Errorf(`Wrong errno (custom checkFunc)`)
	}
	SyscallWrappers.setRealSyscall()
}

func Test_SyscallWrappers_Syscall6(t *testing.T) {
	SyscallWrappers.setWrongSyscall6(nil)
	_, _, errno := errorableSyscallWrappers.Syscall6(1, 2, 3, 4, 5, 6, 7)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno`)
	}
	SyscallWrappers.setRealSyscall6()

	SyscallWrappers.setWrongSyscall6(CheckFuncSkipN(0, nil))
	_, _, errno = errorableSyscallWrappers.Syscall6(1, 2, 3, 4, 5, 6, 7)
	if errno != syscall.EINVAL {
		t.Errorf(`Wrong errno (with checkFunc)`)
	}
	SyscallWrappers.setRealSyscall6()

	SyscallWrappers.setWrongSyscall6(func(data interface{}) bool {
		return true
	})
	_, _, errno = errorableSyscallWrappers.Syscall6(1, 2, 3, 4, 5, 6, 7)
	if errno != syscall.EFAULT {
		t.Errorf(`Wrong errno (custom checkFunc)`)
	}
	SyscallWrappers.setRealSyscall6()
}
