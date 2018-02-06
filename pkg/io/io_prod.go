// +build prod

package io

import "io"

// NewReader creates a new reader to consume.
// In a prod config we want to return the reader as is.
// Development will return a throttled one.
func NewReader(src io.Reader) io.Reader {
	return src
}
