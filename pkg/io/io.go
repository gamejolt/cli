package io

import (
	"context"
	"io"

	"github.com/gamejolt/cli/pkg/concurrency"
)

// ResumableCopy is a modification of io.copyBuffer.
// It works in chunks of 32kb and checks the resumable state before continuing or abandoning the copy.
func ResumableCopy(r concurrency.Resumer, dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if buf == nil {
		buf = make([]byte, 32*1024)
	}

	for {
		if resume := <-r.Wait(); resume == concurrency.OpCancel {
			return written, context.Canceled
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}
