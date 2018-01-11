package TimeoutPipe

import (
	"sync"
	"testing"
	"time"
)

var dummyData []byte = []byte("testing 12345")
var smallBuffer []byte = make([]byte, 2)
var largeBuffer []byte = make([]byte, 100)

func TestNoTimeout(t *testing.T) {

	x := NewTimeoutPipe()
	read, write := x.Pipe(time.Second * 5)

	_, e := write.Write(dummyData)
	if e != nil {
		t.Error("Writer died unexpectedly!")
	}
	r, re := read.Read(largeBuffer)
	if re != nil {
		t.Error("Reader died unexpectedly!")
	}

	if r < len(dummyData) || r > len(dummyData) {
		t.Error("unexpected write or read length")
	}
}

//if this function gets stuck -- something is wrong.
func TestTimeout(t *testing.T) {
	x := NewTimeoutPipe()
	read, _ := x.Pipe(time.Second * 1)

	_, re := read.Read(largeBuffer)
	if re != nil {
		//this is what we want
	}

}

func TestBufferedData(t *testing.T) {
	wg := sync.WaitGroup{}
	x := NewTimeoutPipe()
	read, write := x.Pipe(time.Second * 1)

	wg.Add(1)
	go func() {
		_, e := write.Write(dummyData)
		if e != nil {
			return
			t.Error("Writer died unexpectedly!")
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for {
			_, re := read.Read(smallBuffer)
			if re != nil {
				break
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestWriteLargeData(t *testing.T) {
	//we dont actually need to write large data, just set buffer size to very low ^^
	x := NewTimeoutPipe()
	x.BufferSize = 2
	read, write := x.Pipe(time.Second * 5)
	written := 0

	go func() {
		for {
			_, re := read.Read(largeBuffer)
			if re != nil {
				break
				t.Error("Reader died unexpectedly!")
			}

		}
	}()
	for written < len(dummyData) {
		w, e := write.Write(dummyData)
		if e != nil {
			t.Error("Writer died unexpectedly!")
		}
		written += w

	}

}

func TestReset(t *testing.T) {
	ti := time.Now()
	x := NewTimeoutPipe()
	x.Pipe(time.Second * 1)
	for i := 0; i < 3; i++ {
		x.ResetTimer()
		time.Sleep(time.Second * 1)
	}

	if time.Since(ti) < 1*time.Second {
		t.Error("Timer was not reset properly!")
	}

}
