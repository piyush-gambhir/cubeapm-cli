package output

import (
	"fmt"
	"io"
	"strings"
)

// TableFormatter outputs data as an aligned text table.
type TableFormatter struct {
	NoColor bool
}

// Format writes a TableDef as an aligned text table.
func (f *TableFormatter) Format(w io.Writer, data interface{}) error {
	td, ok := data.(TableDef)
	if !ok {
		// Fall back to JSON for non-table data
		jf := &JSONFormatter{}
		return jf.Format(w, data)
	}

	if len(td.Headers) == 0 {
		return nil
	}

	// Calculate column widths
	widths := make([]int, len(td.Headers))
	for i, h := range td.Headers {
		widths[i] = len(h)
	}
	for _, row := range td.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	header := make([]string, len(td.Headers))
	for i, h := range td.Headers {
		header[i] = padRight(h, widths[i])
	}
	headerLine := strings.Join(header, "  ")
	if !f.NoColor {
		fmt.Fprintf(w, "\033[1m%s\033[0m\n", headerLine)
	} else {
		fmt.Fprintln(w, headerLine)
	}

	// Print rows
	for _, row := range td.Rows {
		cells := make([]string, len(td.Headers))
		for i := range td.Headers {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			cells[i] = padRight(val, widths[i])
		}
		fmt.Fprintln(w, strings.Join(cells, "  "))
	}

	return nil
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
