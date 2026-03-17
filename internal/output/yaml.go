package output

import (
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter outputs data as YAML.
type YAMLFormatter struct{}

// Format writes data as YAML.
func (f *YAMLFormatter) Format(w io.Writer, data interface{}) error {
	// If it's a TableDef, convert to a list of maps for YAML output
	if td, ok := data.(TableDef); ok {
		data = tableDefToMaps(td)
	}

	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(data)
}
