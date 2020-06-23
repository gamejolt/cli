// +build !prod

package io

import (
	"io"
	// "github.com/juju/ratelimit"
)

// NewReader creates a new reader to consume
// This will limit the reader to 1MB/s
func NewReader(src io.Reader) io.Reader {
	// var MB int64 = 1024 * 1024
	// rate := 1 * MB
	// return ratelimit.Reader(src, ratelimit.NewBucketWithRate(float64(rate), rate))
	return src
}
