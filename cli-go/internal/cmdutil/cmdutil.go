package cmdutil

import (
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/output"
)

// Shared state set in PersistentPreRunE and used by subcommands.
var (
	AppConfig    *config.Config
	Resolved     config.ResolvedConfig
	APIClient    *client.Client
	OutputFormat output.Format
	NoInput      bool // Disable interactive prompts (for CI/agent use)
	Quiet        bool // Suppress informational output
)
