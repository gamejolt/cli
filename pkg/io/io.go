package io

import (
	"context"
	"io"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
)

// CopyCallback is a callback you can use to hook into a copy
type CopyCallback func(written int64) bool

// Copy is a modification of io.copyBuffer.
// It runs the given callback after every chunk. If it returns true the copy will be resumed, on false it'll abort the copy
func Copy(dst io.Writer, src io.Reader, buf []byte, cb CopyCallback) (written int64, err error) {
	if buf == nil {
		buf = make([]byte, 32*1024)
	}

	for {
		if cb != nil && !cb(written) {
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

	if cb != nil && !cb(written) {
		return written, context.Canceled
	}

	return written, err
}

type copyResult struct {
	written int64
	err     error
}

// BarMaker is a function that allows you to create a bar on demand for functions that might need it during their run time.
// For example, CopyWithSlowBar runs it if the copy takes too long by running this function to allow bar customization.
type BarMaker func() *pb.ProgressBar

// CopyWithSlowBar does an io copy that displays a progress bar if the copy takes too long.
func CopyWithSlowBar(dest io.Writer, src io.Reader, tooLong time.Duration, makeBar BarMaker) (written int64, err error) {
	var bar *pb.ProgressBar
	ch := make(chan copyResult)

	go func() {
		defer close(ch)

		showBarOn := time.Now().Add(tooLong)
		wantsBar := true
		nextBarRefresh := time.Now()
		cb := func(written int64) bool {
			if time.Now().After(showBarOn) && wantsBar {
				bar = makeBar()

				if bar != nil {
					bar.ManualUpdate = true
					bar.Add64(written)
					bar.Start()
				}

				wantsBar = false
			}

			if bar != nil && (written == bar.Total || time.Now().After(nextBarRefresh)) {
				bar.Set64(written)
				bar.Update()

				nextBarRefresh = time.Now().Add(bar.RefreshRate)
			}

			return true
		}

		written, err := Copy(dest, src, nil, cb)
		ch <- copyResult{written, err}
	}()
	res := <-ch

	if bar != nil {
		bar.Finish()
	}

	return res.written, res.err
}
