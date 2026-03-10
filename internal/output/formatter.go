package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

// Format is the output format type.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Table prints data as an aligned table.
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println(dimStyle.Render("  No results found."))
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Cap max column width
	for i := range widths {
		if widths[i] > 50 {
			widths[i] = 50
		}
	}

	// Print header
	var headerParts []string
	for i, h := range headers {
		headerParts = append(headerParts, pad(h, widths[i]))
	}
	fmt.Println(headerStyle.Render(strings.Join(headerParts, "  ")))
	// Print separator
	var sepParts []string
	for _, w := range widths {
		sepParts = append(sepParts, strings.Repeat("─", w))
	}
	fmt.Println(dimStyle.Render(strings.Join(sepParts, "──")))

	// Print rows
	for _, row := range rows {
		var parts []string
		for i, cell := range row {
			if i < len(widths) {
				if len(cell) > widths[i] {
					cell = cell[:widths[i]-1] + "…"
				}
				parts = append(parts, pad(cell, widths[i]))
			}
		}
		fmt.Println(strings.Join(parts, "  "))
	}
}

// JSON prints data as formatted JSON.
func JSON(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// Success prints a success message.
func Success(msg string) {
	fmt.Println(successStyle.Render("✓ " + msg))
}

// Error prints an error message.
func Error(msg string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render("✗ "+msg))
}

// Warn prints a warning message.
func Warn(msg string) {
	fmt.Println(warnStyle.Render("! " + msg))
}

// Info prints an info message.
func Info(msg string) {
	fmt.Println(dimStyle.Render("  " + msg))
}

// KeyValue prints a key-value pair.
func KeyValue(key, value string) {
	fmt.Printf("  %s  %s\n", dimStyle.Render(pad(key+":", 20)), value)
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
