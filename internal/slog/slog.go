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

// Package slog provides service-level logging for LGH server
package slog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents log severity level
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelFatal Level = "FATAL"
)

// Entry represents a single log entry
type Entry struct {
	Timestamp time.Time              `json:"ts"`
	Level     Level                  `json:"level"`
	Message   string                 `json:"msg"`
	Component string                 `json:"component,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger provides service logging functionality
type Logger struct {
	file      *os.File
	filePath  string
	mu        *sync.Mutex
	component string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger
func Init(dataDir string) error {
	var err error
	once.Do(func() {
		logsDir := filepath.Join(dataDir, "logs")
		if mkErr := os.MkdirAll(logsDir, 0700); mkErr != nil {
			err = mkErr
			return
		}

		logPath := filepath.Join(logsDir, "server.jsonl")
		// nolint:gosec // G304: path is internally constructed
		f, openErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if openErr != nil {
			err = openErr
			return
		}

		defaultLogger = &Logger{
			file:     f,
			filePath: logPath,
			mu:       &sync.Mutex{},
		}
	})
	return err
}

// GetLogPath returns the path to the service log file
func GetLogPath() string {
	if defaultLogger != nil {
		return defaultLogger.filePath
	}
	return ""
}

// Close closes the default logger
func Close() error {
	if defaultLogger != nil && defaultLogger.file != nil {
		return defaultLogger.file.Close()
	}
	return nil
}

// WithComponent returns a logger for a specific component
func WithComponent(name string) *Logger {
	if defaultLogger == nil {
		return nil
	}
	return &Logger{
		file:      defaultLogger.file,
		filePath:  defaultLogger.filePath,
		mu:        defaultLogger.mu,
		component: name,
	}
}

// write writes a log entry
func (l *Logger) write(level Level, msg string, fields map[string]interface{}) {
	if l == nil || l.file == nil {
		return
	}

	entry := Entry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Component: l.component,
		Fields:    fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.file.Write(data)
	_, _ = l.file.WriteString("\n")
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.write(LevelDebug, msg, f)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.write(LevelInfo, msg, f)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.write(LevelWarn, msg, f)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.write(LevelError, msg, f)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.write(LevelError, fmt.Sprintf(format, args...), nil)
}

// Package-level convenience functions

// Debug logs a debug message to default logger
func Debug(msg string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	}
}

// Info logs an info message to default logger
func Info(msg string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	}
}

// Warn logs a warning message to default logger
func Warn(msg string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	}
}

// Error logs an error message to default logger
func Error(msg string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	}
}

// Errorf logs a formatted error message to default logger
func Errorf(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Errorf(format, args...)
	}
}
