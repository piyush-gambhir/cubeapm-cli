package output

import (
	"io"

	"go.yaml.in/yaml/v3"
)

type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(w io.Writer, data interface{}) error {
	if td, ok := data.(TableDef); ok {
		data = tableDefToMaps(td)
	}

	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(data)
}
