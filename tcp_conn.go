package gonet

import (
	"io"
)

type (
	// TCPConn реализует двунаправленый буфер полученных и готовых к отправке данных на соединении
	TCPConn struct {
		fd    int
		RdBuf BufChain
		WrBuf BufChain
	}
)

// Read реализует io.Reader
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
