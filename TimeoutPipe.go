package TimeoutPipe

import (
	"errors"
	"io"
	"time"
)

var TimeoutErr error = errors.New("Pipe has timed out!")

//TimeoutPipe is a io.pipe that times out.
//Purpose of this is to stop goroutine r/w deadlocks by intentionally causing error
//less code to implement on the potentially-deadlocking-r/w goroutines, no need to worry about channels or timers.
type TimeoutPipe struct {
	BufferSize int

	Reader *io.PipeReader
	Writer *io.PipeWriter

	internalReader *io.PipeReader
	internalWriter *io.PipeWriter

	timer   *time.Timer
	timeout time.Duration
}

//NewTimeoutPipe retusn a new instance of TimeoutPipe interface.
func NewTimeoutPipe() *TimeoutPipe {

	r, Wr := io.Pipe()
	Re, w := io.Pipe()
	t := TimeoutPipe{
		BufferSize:     1500,
		internalReader: r,
		internalWriter: w,
		Reader:         Re,
		Writer:         Wr,
	}
	return &t
}

//Pipe returns instance of io.pipereader & io.pipewriter
func (t *TimeoutPipe) Pipe(timeout time.Duration) (*io.PipeReader, *io.PipeWriter) {
	t.timeout = timeout
	t.rwProxy()
	t.start()
	return t.Reader, t.Writer
}

func (t *TimeoutPipe) start() {
	t.timer = time.AfterFunc(t.timeout, func() {
		t.closePipes(TimeoutErr)
	})
}

func (t *TimeoutPipe) rwProxy() {

	go func() {
		var buffer []byte = make([]byte, t.BufferSize)
		for {
			var written int = 0

			readLength, err := t.internalReader.Read(buffer)
			if err != nil || readLength == 0 {
				switch err {
				case io.EOF:
					t.internalWriter.Write(buffer[:readLength])
					t.closePipes(err)
					return
				case io.ErrClosedPipe:
					t.closePipes(err)
					return
				default:
					t.closePipes(err)
					return
				}
			}
			t.ResetTimer()
			for written < readLength {
				w, err := t.internalWriter.Write(buffer[written:readLength])
				written += w
				if err != nil {
					t.closePipes(err)
				}
			}
		}
	}()
}

func (t *TimeoutPipe) closePipes(e error) {
	if e != nil {
		t.internalReader.Close()
		t.internalWriter.Close()

		t.Reader.CloseWithError(e)
		t.Writer.Close()
	} else {
		t.internalReader.Close()
		t.internalWriter.Close()

		t.Reader.Close()
		t.Writer.Close()
	}
}

//
func (t *TimeoutPipe) ResetTimer() {
	t.timer.Stop()
	t.timer.Reset(t.timeout)
}
