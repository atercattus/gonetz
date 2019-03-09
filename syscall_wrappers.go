package gonet

import "syscall"

type (
	syscallWrapperFuncs struct {
		EpollCreate1 func(flag int) (fd int, err error)
	}
)

var (
	// Врапперы над функциями для возможности потестировать сбои в работе этих вызовов (read-only)
	defaultSyscallWrappers = syscallWrapperFuncs{
		EpollCreate1: syscall.EpollCreate1,
	}

	// Рабочие варианты функций
	SyscallWrappers = defaultSyscallWrappers

	// Сбойные варианты функций (read-only)
	errorableSyscallWrappers = syscallWrapperFuncs{
		EpollCreate1: func(flag int) (fd int, err error) {
			return 0, syscall.EINVAL
		},
	}
)

func (sw *syscallWrapperFuncs) setWrongEpollCreate1() {
	sw.EpollCreate1 = errorableSyscallWrappers.EpollCreate1
}

func (sw *syscallWrapperFuncs) setRealEpollCreate1() {
	sw.EpollCreate1 = defaultSyscallWrappers.EpollCreate1
}
