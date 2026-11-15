package cmdutil

import (
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/output"
)

var (
	AppConfig    *config.Config
	Resolved     config.ResolvedConfig
	APIClient    *client.Client
	OutputFormat output.Format
	NoInput      bool
	Quiet        bool
)
