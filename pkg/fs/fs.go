package fs

import (
	"errors"
	"fmt"
	"os"
	"strings"
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
