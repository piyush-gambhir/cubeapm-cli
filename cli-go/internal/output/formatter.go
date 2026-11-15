package output

import (
	"fmt"
	"io"
	"os"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

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

type Formatter interface {
	Format(w io.Writer, data interface{}) error
}

type TableDef struct {
	Headers []string
	Rows    [][]string
}

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

func Print(format Format, noColor bool, data interface{}) error {
	f := NewFormatter(format, noColor)
	return f.Format(os.Stdout, data)
}

func PrintTable(noColor bool, table TableDef) error {
	f := &TableFormatter{NoColor: noColor}
	return f.Format(os.Stdout, table)
}
