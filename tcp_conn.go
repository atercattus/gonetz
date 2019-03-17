package gonet

import (
	"io"
)

type (
	TCPConn struct {
		fd    int
		RdBuf BufChain
		WrBuf BufChain
	}
)

func (conn *TCPConn) Read(b []byte) (n int, err error) {
	n, err = conn.RdBuf.Read(b)
	if n == 0 {
		err = io.EOF
	}
	return
}

//func (conn *TCPConn) Write(b []byte) (n int, err error) {
//	return conn.RdBuf.Write(b)
//}
