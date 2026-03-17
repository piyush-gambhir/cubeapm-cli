package cmdutil

import (
	"github.com/piyush-gambhir/cubeapm-cli/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
)

// Shared state set in PersistentPreRunE and used by subcommands.
var (
	AppConfig    *config.Config
	Resolved     config.ResolvedConfig
	APIClient    *client.Client
	OutputFormat output.Format
)
