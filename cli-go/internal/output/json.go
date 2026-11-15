package output

import (
	"encoding/json"
	"io"
)

type JSONFormatter struct{}

func (f *JSONFormatter) Format(w io.Writer, data interface{}) error {
	if td, ok := data.(TableDef); ok {
		data = tableDefToMaps(td)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}

func tableDefToMaps(td TableDef) []map[string]string {
	result := make([]map[string]string, 0, len(td.Rows))
	for _, row := range td.Rows {
		m := make(map[string]string, len(td.Headers))
		for i, h := range td.Headers {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		result = append(result, m)
	}
	return result
}
