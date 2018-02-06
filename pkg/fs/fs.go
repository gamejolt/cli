package fs

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gamejolt/cli/pkg/concurrency"
	"github.com/gamejolt/cli/pkg/io"
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
	if _, err := io.ResumableCopy(r, hash, file, nil); err != nil {
		return "", err
	}

	var result []byte
	result = hash.Sum(result)
	return hex.EncodeToString(result), nil
}
