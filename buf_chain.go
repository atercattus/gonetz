package gonet

import "sync"

type (
	BufChain struct {
		chain    [][]byte
		totalLen int
	}
)

var (
	bufPool4K = sync.Pool{
		New: func() interface{} {
			return make([]byte, 4096)
		},
	}
)

func (bc *BufChain) Write(buf []byte) {
	bc.totalLen += len(buf)

	for len(buf) > 0 {
		if w := bc.appendToLast(buf); w > 0 {
			buf = buf[w:]
		}
		if len(buf) == 0 {
			break
		}

		bc.growChain()
	}
}

func (bc *BufChain) growChain() {
	chunk := bufPool4K.Get().([]byte)[:0]
	bc.chain = append(bc.chain, chunk)
}

func (bc *BufChain) appendToLast(buf []byte) int {
	lastIdx := len(bc.chain) - 1
	if lastIdx < 0 {
		return 0
	}
	last := bc.chain[lastIdx]

	avail := cap(last) - len(last)
	if avail == 0 {
		return 0
	}

	if len(buf) > avail {
		buf = buf[:avail]
	}
	bc.chain[lastIdx] = append(last, buf...)

	return len(buf)
}

//func (bc *BufChain) Read(buf []byte) (dst []byte) {
//	//return buf[:0]
//}

func (bc *BufChain) Clean() {
	for _, chunk := range bc.chain {
		bufPool4K.Put(chunk)
	}
	bc.chain = [][]byte{}
	bc.totalLen = 0
}
