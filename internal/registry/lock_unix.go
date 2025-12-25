//go:build darwin || linux
// +build darwin linux

package registry

import (
	"os"
	"syscall"
)

// FileLock provides file locking functionality
type FileLock struct {
	path string
	file *os.File
}

// NewFileLock creates a new file lock
func NewFileLock(path string) *FileLock {
	return &FileLock{path: path}
}

// Lock acquires an exclusive lock
func (fl *FileLock) Lock() error {
	f, err := os.OpenFile(fl.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	fl.file = f
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// Unlock releases the lock
func (fl *FileLock) Unlock() error {
	if fl.file == nil {
		return nil
	}
	syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
	return fl.file.Close()
}
