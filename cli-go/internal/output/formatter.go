package output

import (
	"fmt"
	"io"
	"os"
)

// Format represents an output format type.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// ParseFormat parses a string into a Format, returning an error for unknown formats.
func ParseFormat(s string) (Format, error) {
	switch s {
	case "table", "":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unknown output format %q: must be table, json, or yaml", s)
	}
}

// Formatter defines the interface for outputting data.
type Formatter interface {
	// Format writes the data to the writer in the appropriate format.
	Format(w io.Writer, data interface{}) error
}

// TableDef defines the columns for table output.
type TableDef struct {
	Headers []string
	Rows    [][]string
}

// NewFormatter creates a new Formatter for the given format.
func NewFormatter(format Format, noColor bool) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	case FormatYAML:
		return &YAMLFormatter{}
	default:
		return &TableFormatter{NoColor: noColor}
	}
}

// Print is a convenience function to format and print data to stdout.
func Print(format Format, noColor bool, data interface{}) error {
	f := NewFormatter(format, noColor)
	return f.Format(os.Stdout, data)
}

// PrintTable is a convenience function to print a table to stdout.
func PrintTable(noColor bool, table TableDef) error {
	f := &TableFormatter{NoColor: noColor}
	return f.Format(os.Stdout, table)
}
