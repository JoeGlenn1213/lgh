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

// Package ui provides terminal UI components and colorized output for LGH
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	// Green prints text in green
	Green = color.New(color.FgGreen).SprintFunc()
	// Red prints text in red
	Red = color.New(color.FgRed).SprintFunc()
	// Yellow prints text in yellow
	Yellow = color.New(color.FgYellow).SprintFunc()
	// Blue prints text in blue
	Blue = color.New(color.FgBlue).SprintFunc()
	// Cyan prints text in cyan
	Cyan = color.New(color.FgCyan).SprintFunc()
	// Gray prints text in high-intensity black (gray)
	Gray = color.New(color.FgHiBlack).SprintFunc()
	// Bold prints text in bold
	Bold = color.New(color.Bold).SprintFunc()

	// Styled text
	greenBold  = color.New(color.FgGreen, color.Bold).SprintFunc()
	redBold    = color.New(color.FgRed, color.Bold).SprintFunc()
	yellowBold = color.New(color.FgYellow, color.Bold).SprintFunc()
	cyanBold   = color.New(color.FgCyan, color.Bold).SprintFunc()
)

// Success prints a success message
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", greenBold("✓"), msg)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", redBold("✗"), msg)
}

// Warning prints a warning message
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", yellowBold("⚠"), msg)
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", cyanBold("ℹ"), msg)
}

// Title prints a title
func Title(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("\n%s\n", Bold(msg))
	fmt.Println(strings.Repeat("─", len(msg)+2))
}

// Command prints a command that can be executed
func Command(cmd string) {
	fmt.Printf("  %s %s\n", Gray("$"), Cyan(cmd))
}

// URL prints a URL
func URL(url string) string {
	return Blue(url)
}

// Table creates and displays a table
type Table struct {
	headers []string
	rows    [][]string
	writer  io.Writer
}

// NewTable creates a new table with headers
func NewTable(headers []string) *Table {
	return &Table{
		headers: headers,
		rows:    [][]string{},
		writer:  os.Stdout,
	}
}

// NewTableWithWriter creates a table with a custom writer
func NewTableWithWriter(w io.Writer, headers []string) *Table {
	return &Table{
		headers: headers,
		rows:    [][]string{},
		writer:  w,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

// stripAnsi removes ANSI escape codes for length calculation
func stripAnsi(s string) string {
	// Simple ANSI escape sequence removal
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

// Render displays the table
func (t *Table) Render() {
	if len(t.headers) == 0 {
		return
	}

	// Calculate column widths (accounting for ANSI codes)
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(stripAnsi(h))
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) {
				cellLen := len(stripAnsi(cell))
				if cellLen > widths[i] {
					widths[i] = cellLen
				}
			}
		}
	}

	// Print headers
	for i, h := range t.headers {
		padding := widths[i] - len(stripAnsi(h))
		fmt.Fprintf(t.writer, "%s%s  ", Bold(h), strings.Repeat(" ", padding))
	}
	fmt.Fprintln(t.writer)

	// Print separator
	for i := range t.headers {
		fmt.Fprintf(t.writer, "%s  ", strings.Repeat("─", widths[i]))
	}
	fmt.Fprintln(t.writer)

	// Print rows
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) {
				padding := widths[i] - len(stripAnsi(cell))
				if padding < 0 {
					padding = 0
				}
				fmt.Fprintf(t.writer, "%s%s  ", cell, strings.Repeat(" ", padding))
			}
		}
		fmt.Fprintln(t.writer)
	}
}

// ProgressBar displays a simple progress indicator
func ProgressBar(current, total int, label string) {
	percent := float64(current) / float64(total) * 100
	barWidth := 30
	filled := int(percent / 100 * float64(barWidth))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("\r%s [%s] %.0f%% ", label, bar, percent)

	if current == total {
		fmt.Println()
	}
}

// Spinner displays a simple spinner (call in a goroutine)
type Spinner struct {
	chars   []string
	message string
	stop    chan bool
	running bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		chars:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		stop:    make(chan bool),
	}
}

// Start starts the spinner
func (s *Spinner) Start() {
	s.running = true
	go func() {
		i := 0
		for {
			select {
			case <-s.stop:
				return
			default:
				fmt.Printf("\r%s %s", Cyan(s.chars[i]), s.message)
				i = (i + 1) % len(s.chars)
				// Small sleep handled externally
			}
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop(success bool) {
	if !s.running {
		return
	}
	s.stop <- true
	s.running = false

	if success {
		fmt.Printf("\r%s %s\n", Green("✓"), s.message)
	} else {
		fmt.Printf("\r%s %s\n", Red("✗"), s.message)
	}
}

// Box prints text in a box
func Box(title string, lines []string) {
	maxLen := len(title)
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	border := "─"
	fmt.Printf("┌%s┐\n", strings.Repeat(border, maxLen+2))
	fmt.Printf("│ %s%s │\n", Bold(title), strings.Repeat(" ", maxLen-len(title)))
	fmt.Printf("├%s┤\n", strings.Repeat(border, maxLen+2))

	for _, line := range lines {
		fmt.Printf("│ %s%s │\n", line, strings.Repeat(" ", maxLen-len(line)))
	}

	fmt.Printf("└%s┘\n", strings.Repeat(border, maxLen+2))
}

// DisableColors disables color output
func DisableColors() {
	color.NoColor = true
}

// EnableColors enables color output
func EnableColors() {
	color.NoColor = false
}
