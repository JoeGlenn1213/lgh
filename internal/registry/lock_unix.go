// Copyright (c) 2025 JoeGlenn1213
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:build darwin || linux
// +build darwin linux

// Package registry provide platform-specific file locking and project registration
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
	f, err := os.OpenFile(fl.path, os.O_CREATE|os.O_RDWR, 0600)
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
	_ = syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
	return fl.file.Close()
}
