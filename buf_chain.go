package gonet

import (
	"sync"
)

type (
	// BufChain - это цепочка буферов для хранения потоковых данных
	BufChain struct {
		chain           [][]byte
		chainIf         []interface{}
		totalLen        int
		posInFirstChunk int
	}
)

var (
	bufPool4K = sync.Pool{ // ToDo: можно сделать мой вариант канал+пул (но только с bench сравнением)
		New: func() interface{} {
			return make([]byte, 4096)
		},
	}
)

// Len возвращает количество записанных, но еще не прочитанных байт во всей цепочке буферов
//   (т.е. столько еще можно вычитать через Read)
func (bc *BufChain) Len() int {
	return bc.totalLen
}

// Write реализует io.Writer.
// Никогда не возвращает error
func (bc *BufChain) Write(buf []byte) (n int, _ error) {
	n = len(buf)
	bc.totalLen += n

	for len(buf) > 0 {
		if w := bc.appendToLast(buf); w > 0 {
			buf = buf[w:]
		}
		if len(buf) == 0 {
			break
		}

		bc.growChain()
	}

	return
}

func (bc *BufChain) growChain() {
	chunkIf := bufPool4K.Get()
	chunk := chunkIf.([]byte)[:0]

	bc.chain = append(bc.chain, chunk)
	bc.chainIf = append(bc.chainIf, chunkIf)

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

// Read реализует io.Reader.
// Никогда не возвращает error
func (bc *BufChain) Read(buf []byte) (n int, _ error) {
	var (
		bufPos    int
		bufLen    = len(buf)
		oldChunks = -1 // максимальный номер чанка, который уже не нужен
	)

	for chunkIdx, chunk := range bc.chain {
		ch := chunk[bc.posInFirstChunk:]
		rdd := copy(buf[bufPos:], ch)

		if rdd > 0 {
			n += rdd
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

	// Если последний чанк прочитал полностью, то не возвращаю его в пул, а оставляю на будущее,
	//   чтобы уменьшить число взаимодействий с sync.Pool
	if (oldChunks > -1) && (oldChunks == len(bc.chain)-1) {
		bc.chain[oldChunks] = bc.chain[oldChunks][:0]
		oldChunks--
	}

	if oldChunks > -1 {
		for i := 0; i <= oldChunks; i++ {
			bufPool4K.Put(bc.chainIf[i])
		}

		if oldChunks < len(bc.chain)-1 {
			from := oldChunks + 1
			copy(bc.chain[0:], bc.chain[from:])
			copy(bc.chainIf[0:], bc.chainIf[from:])

			to := len(bc.chain) - oldChunks - 1
			bc.chain = bc.chain[:to]
			bc.chainIf = bc.chainIf[:to]
		}
	}

	return
}

// Clean очищает внутренние буферы, чтобы перевести буфер в изначальное состояние.
func (bc *BufChain) Clean() {
	for _, chunkIf := range bc.chainIf {
		bufPool4K.Put(chunkIf)
	}
	bc.chain = [][]byte{}
	bc.chainIf = []interface{}{}
	bc.totalLen = 0
}
