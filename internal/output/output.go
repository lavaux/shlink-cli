// Package output provides helpers for rendering CLI output in different formats.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Format is the output format type.
type Format string

const (
	// FormatTable renders output as a formatted table.
	FormatTable Format = "table"
	// FormatJSON renders output as a JSON object.
	FormatJSON Format = "json"
	// FormatPlain renders output as plain text.
	FormatPlain Format = "plain"
)

// Printer writes formatted output.
type Printer struct {
	Format Format
	Out    io.Writer
}

// safeFprintln writes a line and prints error to stderr if it fails.
func safeFprintln(w io.Writer, a ...interface{}) {
	if _, err := fmt.Fprintln(w, a...); err != nil {
		fmt.Fprintf(os.Stderr, "output error: %v\n", err)
	}
}

// safeFprintln writes a line and prints error to stderr if it fails.
func safeFprintf(w io.Writer, format string, a ...interface{}) {
	if _, err := fmt.Fprintf(w, format, a...); err != nil {
		fmt.Fprintf(os.Stderr, "output error: %v\n", err)
	}
}

func safeFlush(w *tabwriter.Writer) {
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "flush error: %v\n", err)
	}
}

// New creates a Printer for the given format string, defaulting to table.
func New(format string) *Printer {
	f := Format(format)
	if f != FormatTable && f != FormatJSON && f != FormatPlain {
		f = FormatTable
	}
	return &Printer{Format: f, Out: os.Stdout}
}

// JSON pretty-prints raw JSON bytes.
func (p *Printer) JSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		_, err2 := fmt.Fprintln(p.Out, string(data))
		return err2
	}
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Table renders a table with the given headers and rows.
func (p *Printer) Table(headers []string, rows [][]string) {
	switch p.Format {
	case FormatJSON:
		records := make([]map[string]string, len(rows))
		for i, row := range rows {
			m := make(map[string]string, len(headers))
			for j, h := range headers {
				if j < len(row) {
					m[h] = row[j]
				}
			}
			records[i] = m
		}
		enc := json.NewEncoder(p.Out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(records)
	case FormatPlain:
		for _, row := range rows {
			safeFprintln(p.Out, strings.Join(row, "\t"))
		}
	default: // table
		w := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
		safeFprintln(w, strings.Join(headers, "\t"))
		safeFprintln(w, strings.Repeat("-", 80))
		for _, row := range rows {
			safeFprintln(w, strings.Join(row, "\t"))
		}
		safeFlush(w)
	}
}

// Line prints a single formatted line.
func (p *Printer) Line(format string, args ...interface{}) {
	safeFprintf(p.Out, format+"\n", args...)
}

// Success prints a success message (always to stdout regardless of format).
func (p *Printer) Success(format string, args ...interface{}) {
	safeFprintf(p.Out, "✓ "+format+"\n", args...)
}

// Truncate shortens a string to maxLen, adding "…" if needed.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// Bool renders a boolean as a short string.
func Bool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
