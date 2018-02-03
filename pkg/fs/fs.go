package fs

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gamejolt/cli/pkg/concurrency"
)

// Filesize returns the filesize of a given file
// Just a utility function to make sure with get the actual filesize and not 0 for symlinks
func Filesize(path string) (int64, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	if finfo.IsDir() {
		return 0, fmt.Errorf("Not computing filesize for directory %s", path)
	}

	return finfo.Size(), nil
}

// IsInDir checks if a path is in a given directory.
func IsInDir(path, dir string) bool {
	return path == dir || strings.HasPrefix(path, dir+string(os.PathSeparator))
}

// EnsureFolder checks if a folder exists, and if not attempts to create it.
func EnsureFolder(dir string) error {
	stat, err := os.Lstat(dir)
	if err == nil {
		if !stat.IsDir() {
			return errors.New("Directory is not a valid directory")
		}
		return nil
	}

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(dir, 0755)
}

// ResumableMD5File calculates the md5 of a file in a resumable way
func ResumableMD5File(r concurrency.Resumer, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := ResumableCopy(r, hash, file, nil); err != nil {
		return "", err
	}

	var result []byte
	result = hash.Sum(result)
	return hex.EncodeToString(result), nil
}

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
