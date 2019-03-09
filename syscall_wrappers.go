package gonet

import (
	"syscall"
)

type (
	syscallWrapperFuncs struct {
		EpollCreate1  func(flag int) (fd int, err error)
		Socket        func(domain, typ, proto int) (fd int, err error)
		SetNonblock   func(fd int, nonblocking bool) (err error)
		SetsockoptInt func(fd, level, opt int, value int) (err error)
		Bind          func(fd int, sa syscall.Sockaddr) (err error)
		Listen        func(s int, n int) (err error)
		Syscall6      func(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

		EpollCreate1Skips int
		//SocketSkips        int
		//SetNonblockSkips   int
		//SetsockoptIntSkips int
		//BindSkips          int
		//ListenSkips        int
		Syscall6Skips int

		Syscall6Trap uintptr
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
		Syscall6:      syscall.Syscall6,
	}

	// Сбойные варианты функций (read-only)
	errorableSyscallWrappers = syscallWrapperFuncs{
		EpollCreate1: func(flag int) (fd int, err error) {
			if SyscallWrappers.EpollCreate1Skips > 0 {
				SyscallWrappers.EpollCreate1Skips--
				return defaultSyscallWrappers.EpollCreate1(flag)
			}
			return 0, syscall.EINVAL
		},

		Socket: func(domain, typ, proto int) (fd int, err error) {
			return 0, syscall.EINVAL
		},

		SetNonblock: func(fd int, nonblocking bool) (err error) {
			return syscall.EINVAL
		},

		SetsockoptInt: func(fd, level, opt int, value int) (err error) {
			return syscall.EINVAL
		},

		Bind: func(fd int, sa syscall.Sockaddr) (err error) {
			return syscall.EINVAL
		},

		Listen: func(s int, n int) (err error) {
			return syscall.EINVAL
		},

		Syscall6: func(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
			if (SyscallWrappers.Syscall6Trap != 0) && (trap != SyscallWrappers.Syscall6Trap) {
			} else if SyscallWrappers.Syscall6Skips > 0 {
				SyscallWrappers.Syscall6Skips--
			} else {
				return 0, 0, syscall.EINVAL
			}
			return defaultSyscallWrappers.Syscall6(trap, a1, a2, a3, a4, a5, a6)
		},
	}

	// Рабочие варианты функций
	SyscallWrappers = defaultSyscallWrappers
)

func (sw *syscallWrapperFuncs) setWrongEpollCreate1(skips int) {
	sw.EpollCreate1 = errorableSyscallWrappers.EpollCreate1
	sw.EpollCreate1Skips = skips
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

func (sw *syscallWrapperFuncs) setWrongSetsockoptInt() {
	sw.SetsockoptInt = errorableSyscallWrappers.SetsockoptInt
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

func (sw *syscallWrapperFuncs) setWrongSyscall6(trap uintptr, skip int) {
	sw.Syscall6 = errorableSyscallWrappers.Syscall6
	sw.Syscall6Trap = trap
	sw.Syscall6Skips = skip
}

func (sw *syscallWrapperFuncs) setRealSyscall6() {
	sw.Syscall6 = defaultSyscallWrappers.Syscall6
}
