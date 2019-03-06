package gonet

import (
	"bytes"
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

func Test_growChain(t *testing.T) {
	var (
		bc BufChain
	)

	if l, exp := len(bc.chain), 0; l != exp {
		t.Fatalf(`precheck expect %d got %d`, exp, l)
	}

	bc.growChain()

	if l, exp := len(bc.chain), 1; l != exp {
		t.Fatalf(`growChain expect %d got %d`, exp, l)
	}

	iters := 10
	for iter := 1; iter <= iters; iter++ {
		bc.growChain()
	}

	if l, exp := len(bc.chain), 1+iters; l != exp {
		t.Fatalf(`growChain expect %d got %d`, exp, l)
	}
}

func Test_appendToLast(t *testing.T) {
	var (
		bc  BufChain
		buf = []byte(`test`)
	)

	if w, exp := bc.appendToLast(buf), 0; w != exp {
		t.Fatalf(`appendToLast expect %d got %d`, exp, w)
	}

	bc.growChain()

	if w, exp := bc.appendToLast(buf), len(buf); w != exp {
		t.Fatalf(`appendToLast expect %d got %d`, exp, w)
	}

	if w, exp := bc.appendToLast(buf), len(buf); w != exp {
		t.Fatalf(`appendToLast expect %d got %d`, exp, w)
	}
}

func Test_Clean(t *testing.T) {
	var (
		bc BufChain
	)

	bc.Clean() // all must be ok

	for i := 0; i < 100; i++ {
		bc.growChain()
	}
	bc.totalLen = 100500 // for test

	bc.Clean()

	if w, exp := bc.totalLen, 0; w != exp {
		t.Fatalf(`Clean bc.totalLen expect %d got %d`, exp, w)
	}

	if w, exp := len(bc.chain), 0; w != exp {
		t.Fatalf(`Clean len(bc.chain) expect %d got %d`, exp, w)
	}
}

func Test_Write_fixed(t *testing.T) {
	var (
		bc   BufChain
		data [][]byte

		totalLen int
		plainBuf []byte

		tmpBuf []byte
	)

	for i := 1; i <= 9; i++ {
		b := byte(i - 1 + '1')
		data = append(data, bytes.Repeat([]byte{b}, i*1024))
	}

	for _, chunk := range data {
		totalLen += len(chunk)
		bc.Write(chunk)
		plainBuf = append(plainBuf, chunk...)
	}

	if totalLen != bc.totalLen {
		t.Errorf(`bc.totalLen mismatch: expect %d got %d`, totalLen, bc.totalLen)
	}

	for _, chunk := range bc.chain {
		tmpBuf = append(tmpBuf, chunk...)
	}

	if len(plainBuf) != len(tmpBuf) {
		t.Errorf(`plain buf len mismatch: expect %d got %d`, len(plainBuf), len(tmpBuf))
	}

	if !bytes.Equal(plainBuf, tmpBuf) {
		t.Errorf(`buf content differs: expect %s got %s`, plainBuf, tmpBuf)
	}
}

func Test_Write_randSize(t *testing.T) {
	if testing.Short() {
		t.Skip(`skipping test in short mode`)
	}

	var (
		bc   BufChain
		data [][]byte

		totalLen int
		plainBuf []byte

		tmpBuf []byte
	)

	type stage struct {
		name    string
		beginCb func()
	}

	stages := [...]stage{
		{`plain`, func() {}},
		{`clean1`, func() { bc.Clean() }},
		{`clean2`, func() { bc.Clean() }},
	}

	for _, stageInfo := range stages {
		data = [][]byte{}
		totalLen = 0
		plainBuf = plainBuf[:0]
		tmpBuf = tmpBuf[:0]

		stageInfo.beginCb()

		runtime.GC()

		for i := 1; i <= 200; i++ {
			b := byte(i - 1 + '1')

			var l int
			if rand.Intn(2) < 1 {
				l = rand.Intn(256)
			} else {
				l = (16+rand.Intn(64))*1024 + rand.Intn(1024)
			}

			data = append(data, bytes.Repeat([]byte{b}, i*(l+1)))
		}

		for _, chunk := range data {
			totalLen += len(chunk)
			bc.Write(chunk)
			plainBuf = append(plainBuf, chunk...)
		}

		for _, chunk := range bc.chain {
			tmpBuf = append(tmpBuf, chunk...)
		}

		if totalLen != bc.totalLen {
			t.Fatalf(`%s: bc.totalLen mismatch: expect %d got %d`, stageInfo.name, totalLen, bc.totalLen)
		}

		if len(plainBuf) != len(tmpBuf) {
			t.Fatalf(`%s: plain buf len mismatch: expect %d got %d`, stageInfo.name, len(plainBuf), len(tmpBuf))
		}

		if !bytes.Equal(plainBuf, tmpBuf) {
			t.Fatalf(`%s: buf content differs`, stageInfo.name)
		}
	}
}

func Test_Write_goro(t *testing.T) {
	if testing.Short() {
		t.Skip(`skipping test in short mode`)
	}

	var wg sync.WaitGroup

	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				bc   BufChain
				data [][]byte

				totalLen int
				plainBuf []byte

				tmpBuf []byte
			)

			for iter := 0; iter < 10; iter++ {
				bc.Clean()
				data = data[:0]
				totalLen = 0
				plainBuf = plainBuf[:0]
				tmpBuf = tmpBuf[:0]
				runtime.GC()

				for i := 1; i <= 1000; i++ {
					b := byte(i - 1 + '1')

					l := 16 + rand.Intn(256)
					data = append(data, bytes.Repeat([]byte{b}, i*l))
				}

				for _, chunk := range data {
					totalLen += len(chunk)
					bc.Write(chunk)
					plainBuf = append(plainBuf, chunk...)

					if rand.Intn(500) < 1 {
						bc.Clean()
						totalLen = 0
						plainBuf = plainBuf[:0]
					}
				}

				for _, chunk := range bc.chain {
					tmpBuf = append(tmpBuf, chunk...)
				}

				if totalLen != bc.totalLen {
					t.Fatalf(`bc.totalLen mismatch: expect %d got %d`, totalLen, bc.totalLen)
				}

				if len(plainBuf) != len(tmpBuf) {
					t.Fatalf(`plain buf len mismatch: expect %d got %d`, len(plainBuf), len(tmpBuf))
				}

				if !bytes.Equal(plainBuf, tmpBuf) {
					t.Fatal(`buf content differs`)
				}
			}
		}()
	}
	wg.Wait()
}
