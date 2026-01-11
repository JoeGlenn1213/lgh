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

package slog

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// ReadLastLines reads the last n lines from a log file, optionally filtering by level.
// If levelFilter is empty, all levels are returned.
func ReadLastLines(path string, n int, levelFilter string) ([]string, error) {
	// nolint:gosec // G304: path is trusted
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read last chunk for performance
	const maxReadSize = 2 * 1024 * 1024
	var offset int64

	fi, err := file.Stat()
	if err == nil && fi.Size() > maxReadSize {
		offset = fi.Size() - maxReadSize
		if _, err := file.Seek(offset, 0); err != nil {
			return nil, err
		}
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	// Skip partial line if seeked
	if offset > 0 {
		scanner.Scan()
	}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Filter and select last N
	var output []string
	count := 0

	filter := strings.ToUpper(levelFilter)

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// Level filter
		if filter != "" {
			var entry Entry
			if err := json.Unmarshal([]byte(line), &entry); err == nil {
				if string(entry.Level) != filter {
					continue
				}
			}
		}

		output = append([]string{line}, output...)
		count++
		if count >= n {
			break
		}
	}

	return output, nil
}
