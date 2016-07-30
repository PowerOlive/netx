package netx

import (
	"io"
	"net"
	"sync/atomic"
	"time"
)

const (
	ioTimeout       = "i/o timeout"
	ioTimeoutLength = 11
)

var (
	copyTimeout = 1 * time.Second
)

// BidiCopy copies between in and out in both directions using the specified
// buffers, returning the errors from copying to out and copying to in. BidiCopy
// continues trying to write out to the respective connections for up to
// writeTimeout to flush any buffered data before giving up and returning
// io.ErrShortWrite.
func BidiCopy(out net.Conn, in net.Conn, bufOut []byte, bufIn []byte, writeTimeout time.Duration) (outErr error, inErr error) {
	stop := uint32(0)
	outErrCh := make(chan error, 1)
	inErrCh := make(chan error, 1)
	go doCopy(out, in, bufIn, writeTimeout, outErrCh, &stop)
	go doCopy(in, out, bufOut, writeTimeout, inErrCh, &stop)
	return <-outErrCh, <-inErrCh
}

// doCopy is based on io.copyBuffer
func doCopy(dst net.Conn, src net.Conn, buf []byte, writeTimeout time.Duration, errCh chan error, stop *uint32) {
	var err error
	defer func() {
		atomic.StoreUint32(stop, 1)
		dst.SetReadDeadline(time.Now().Add(copyTimeout))
		errCh <- err
	}()

	for {
		stopping := atomic.LoadUint32(stop) == 1
		if stopping {
			src.SetReadDeadline(time.Now().Add(copyTimeout))
		}
		nr, er := src.Read(buf)
		if nr > 0 {
			dst.SetWriteDeadline(time.Now().Add(writeTimeout))
			nw, ew := dst.Write(buf[0:nr])
			if err != nil {
				err = ew
			}
			if nw != nr {
				err = io.ErrShortWrite
				return
			}
		}
		if er == io.EOF {
			return
		}
		if er != nil {
			if isTimeout(er) {
				if stopping {
					return
				}
			} else {
				err = er
				return
			}
		}
	}
}

func isTimeout(err error) bool {
	es := err.Error()
	esl := len(es)
	return esl >= ioTimeoutLength && es[esl-ioTimeoutLength:] == ioTimeout
}
