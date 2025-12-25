//go:build windows
// +build windows

package registry

import (
	"os"
	"sync"
)

// Windows uses a global mutex since Windows file locking is different
var globalMutex sync.Mutex

// FileLock provides file locking functionality
// On Windows, we use a simple mutex-based approach
type FileLock struct {
	path string
	file *os.File
}

// NewFileLock creates a new file lock
func NewFileLock(path string) *FileLock {
	return &FileLock{path: path}
}

// Lock acquires an exclusive lock
// On Windows, this uses a global mutex
func (fl *FileLock) Lock() error {
	globalMutex.Lock()
	f, err := os.OpenFile(fl.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		globalMutex.Unlock()
		return err
	}
	fl.file = f
	return nil
}

// Unlock releases the lock
func (fl *FileLock) Unlock() error {
	defer globalMutex.Unlock()
	if fl.file == nil {
		return nil
	}
	return fl.file.Close()
}
