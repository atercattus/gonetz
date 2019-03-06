package gonet

import "sync"

type (
	BufChain struct {
		chain           [][]byte
		totalLen        int
		posInFirstChunk int
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
	bc.posInFirstChunk = 0
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

func (bc *BufChain) Read(buf []byte) (readed int) {
	var (
		bufPos    int
		bufLen    = len(buf)
		oldChunks = -1 // максимальный номер чанка, который уже не нужен
	)

	for chunkIdx, chunk := range bc.chain {
		ch := chunk[bc.posInFirstChunk:]
		rdd := copy(buf[bufPos:], ch)

		if rdd > 0 {
			readed += rdd
			bc.totalLen -= rdd

			if bc.posInFirstChunk += rdd; bc.posInFirstChunk >= len(chunk) {
				// текущий чанк закончился
				bc.posInFirstChunk = 0
				oldChunks = chunkIdx
			}

			if bufPos += rdd; bufPos == bufLen {
				break
			}
		}
	}

	if oldChunks > -1 {
		for i := 0; i <= oldChunks; i++ {
			bufPool4K.Put(bc.chain[i])
		}

		if oldChunks < len(bc.chain)-1 {
			copy(bc.chain[0:], bc.chain[oldChunks+1:])
			bc.chain = bc.chain[:len(bc.chain)-oldChunks-1]
		} else {
			bc.chain = bc.chain[:0] // gc?
		}
	}

	return readed
}

func (bc *BufChain) Clean() {
	for _, chunk := range bc.chain {
		bufPool4K.Put(chunk)
	}
	bc.chain = [][]byte{}
	bc.totalLen = 0
}
