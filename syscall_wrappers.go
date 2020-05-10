package gonet

import (
	"syscall"
)

type (
	syscallCheckFunc func(data interface{}) bool

	syscallWrapperFuncs struct {
		EpollCreate1  func(flag int) (fd int, err error)
		Socket        func(domain, typ, proto int) (fd int, err error)
		SetNonblock   func(fd int, nonblocking bool) (err error)
		SetsockoptInt func(fd, level, opt int, value int) (err error)
		Bind          func(fd int, sa syscall.Sockaddr) (err error)
		Listen        func(s int, n int) (err error)
		Syscall       func(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
		Syscall6      func(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

		EpollCreate1Check  syscallCheckFunc
		SetsockoptIntCheck syscallCheckFunc
		SyscallCheck       syscallCheckFunc
		Syscall6Check      syscallCheckFunc
	}
)

var (
	// Врапперы над функциями для возможности потестировать сбои в работе этих вызовов (read-only)
	defaultSyscallWrappers = syscallWrapperFuncs{
		EpollCreate1:  syscall.EpollCreate1,
		Socket:        syscall.Socket,
		SetNonblock:   syscall.SetNonblock,
		SetsockoptInt: syscall.SetsockoptInt,
		Bind:          syscall.Bind,
		Listen:        syscall.Listen,
		Syscall:       syscall.Syscall,
		Syscall6:      syscall.Syscall6,
	}

	// Сбойные варианты функций (read-only)
	errorableSyscallWrappers = syscallWrapperFuncs{
		EpollCreate1: func(flag int) (fd int, err error) {
			if cb := syscallWrappers.EpollCreate1Check; cb != nil {
				if !cb(flag) {
					return 0, syscall.EINVAL
				}
			} else {
				return 0, syscall.EINVAL
			}
			return defaultSyscallWrappers.EpollCreate1(flag)
		},

		Socket: func(domain, typ, proto int) (fd int, err error) {
			return 0, syscall.EINVAL
		},

		SetNonblock: func(fd int, nonblocking bool) (err error) {
			return syscall.EINVAL
		},

		SetsockoptInt: func(fd, level, opt int, value int) (err error) {
			if cb := syscallWrappers.SetsockoptIntCheck; cb != nil {
				args := []int{fd, level, opt, value}
				if !cb(args) {
					return syscall.EINVAL
				}
			} else {
				return syscall.EINVAL
			}
			return defaultSyscallWrappers.SetsockoptInt(fd, level, opt, value)
		},

		Bind: func(fd int, sa syscall.Sockaddr) (err error) {
			return syscall.EINVAL
		},

		Listen: func(s int, n int) (err error) {
			return syscall.EINVAL
		},

		Syscall: func(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
			if cb := syscallWrappers.SyscallCheck; cb != nil {
				args := []uintptr{trap, a1, a2, a3}
				if !cb(args) {
					return 0, 0, syscall.EINVAL
				}
			} else {
				return 0, 0, syscall.EINVAL
			}
			return defaultSyscallWrappers.Syscall(trap, a1, a2, a3)
		},

		Syscall6: func(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
			if cb := syscallWrappers.Syscall6Check; cb != nil {
				args := []uintptr{trap, a1, a2, a3, a4, a5, a6}
				if !cb(args) {
					return 0, 0, syscall.EINVAL
				}
			} else {
				return 0, 0, syscall.EINVAL
			}
			return defaultSyscallWrappers.Syscall6(trap, a1, a2, a3, a4, a5, a6)
		},
	}

	// Рабочие варианты функций
	syscallWrappers = defaultSyscallWrappers
)

var (
	// Запрещает первые calls вызовов.
	// Если calls == 0, то запрещает все вызовы
	checkFuncSkipN = func(calls int, nextCheck syscallCheckFunc) syscallCheckFunc {
		call := 0
		return func(data interface{}) bool {
			if calls == 0 {
				return false
			} else if call < calls {
				call++
				return false
			}

			if nextCheck != nil {
				return nextCheck(data)
			}
			return true
		}
	}

	// Запрещает первые calls вызовов trap для Syscall + Syscall6.
	// Если calls == 0, то запрещает все вызовы
	checkFuncSyscallTrapSkipN = func(trap uintptr, calls int, nextCheck syscallCheckFunc) syscallCheckFunc {
		call := 0
		return func(data interface{}) bool {
			if args, ok := data.([]uintptr); ok {
				if len(args) > 0 {
					if args[0] != trap {
					} else if calls == 0 {
						return false
					} else if call < calls {
						call++
						return false
					}
				}
			}

			if nextCheck != nil {
				return nextCheck(data)
			}
			return true
		}
	}
)

func (sw *syscallWrapperFuncs) setWrongEpollCreate1(checkFunc syscallCheckFunc) {
	sw.EpollCreate1 = errorableSyscallWrappers.EpollCreate1
	sw.EpollCreate1Check = checkFunc
}

func (sw *syscallWrapperFuncs) setRealEpollCreate1() {
	sw.EpollCreate1 = defaultSyscallWrappers.EpollCreate1
}

func (sw *syscallWrapperFuncs) setWrongSocket() {
	sw.Socket = errorableSyscallWrappers.Socket
}

func (sw *syscallWrapperFuncs) setRealSocket() {
	sw.Socket = defaultSyscallWrappers.Socket
}

func (sw *syscallWrapperFuncs) setWrongSetNonblock() {
	sw.SetNonblock = errorableSyscallWrappers.SetNonblock
}

func (sw *syscallWrapperFuncs) setRealSetNonblock() {
	sw.SetNonblock = defaultSyscallWrappers.SetNonblock
}

func (sw *syscallWrapperFuncs) setWrongSetsockoptInt(checkFunc syscallCheckFunc) {
	sw.SetsockoptInt = errorableSyscallWrappers.SetsockoptInt
	sw.SetsockoptIntCheck = checkFunc
}

func (sw *syscallWrapperFuncs) setRealSetsockoptInt() {
	sw.SetsockoptInt = defaultSyscallWrappers.SetsockoptInt
}

func (sw *syscallWrapperFuncs) setWrongBind() {
	sw.Bind = errorableSyscallWrappers.Bind
}

func (sw *syscallWrapperFuncs) setRealBind() {
	sw.Bind = defaultSyscallWrappers.Bind
}

func (sw *syscallWrapperFuncs) setWrongListen() {
	sw.Listen = errorableSyscallWrappers.Listen
}

func (sw *syscallWrapperFuncs) setRealListen() {
	sw.Listen = defaultSyscallWrappers.Listen
}

func (sw *syscallWrapperFuncs) setWrongSyscall(checkFunc syscallCheckFunc) {
	sw.Syscall = errorableSyscallWrappers.Syscall
	sw.SyscallCheck = checkFunc
}

func (sw *syscallWrapperFuncs) setRealSyscall() {
	sw.Syscall = defaultSyscallWrappers.Syscall
	sw.SyscallCheck = nil
}

func (sw *syscallWrapperFuncs) setWrongSyscall6(checkFunc syscallCheckFunc) {
	sw.Syscall6 = errorableSyscallWrappers.Syscall6
	sw.Syscall6Check = checkFunc
}

func (sw *syscallWrapperFuncs) setRealSyscall6() {
	sw.Syscall6 = defaultSyscallWrappers.Syscall6
	sw.Syscall6Check = nil
}
