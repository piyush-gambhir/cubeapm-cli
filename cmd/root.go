package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	cmdconfig "github.com/piyush-gambhir/cubeapm-cli/cmd/config"
	cmdingest "github.com/piyush-gambhir/cubeapm-cli/cmd/ingest"
	cmdlogs "github.com/piyush-gambhir/cubeapm-cli/cmd/logs"
	cmdmetrics "github.com/piyush-gambhir/cubeapm-cli/cmd/metrics"
	cmdtraces "github.com/piyush-gambhir/cubeapm-cli/cmd/traces"
	"github.com/piyush-gambhir/cubeapm-cli/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/update"
)

// Build-time variables set via ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Global flag values.
var (
	flagOutput     string
	flagProfile    string
	flagServer     string
	flagToken      string
	flagQueryPort  int
	flagIngestPort int
	flagAdminPort  int
	flagNoColor    bool
	flagVerbose    bool
	flagReadOnly   bool
	flagNoInput    bool
	flagQuiet      bool
)

// Background update check state.
var (
	updateInfo     *update.UpdateInfo
	updateInfoOnce sync.Once
	updateInfoDone = make(chan struct{})
)

var rootCmd = &cobra.Command{
	Use:   "cubeapm",
	Short: "CubeAPM CLI - Interact with CubeAPM observability platform",
	Long: `CubeAPM CLI provides a command-line interface for querying traces, metrics,
and logs from CubeAPM. It supports Jaeger-compatible traces, Prometheus-compatible
metrics (PromQL), and VictoriaLogs-compatible logs (LogsQL) APIs.

Command groups:
  traces   Search, view, and analyze distributed traces (Jaeger API)
  metrics  Query and explore Prometheus-compatible metrics (PromQL)
  logs     Query and manage logs (VictoriaLogs / LogsQL)
  ingest   Push metrics and log data to CubeAPM
  config   Manage CLI configuration and connection profiles
  login    Interactively set up a connection profile
  version  Print CLI version information
  update   Check for and install CLI updates

Global flags (apply to all commands):
  -o, --output <format>   Output format: table (default), json, yaml
  --server <addr>         Override server address
  --token <token>         Override authentication token
  --profile <name>        Use a specific connection profile
  --query-port <port>     Override query port (default: 3140)
  --ingest-port <port>    Override ingest port (default: 3130)
  --admin-port <port>     Override admin port (default: 3199)
  --no-color              Disable colored output
  --verbose               Enable verbose HTTP request logging
  --no-input              Disable all interactive prompts (for CI/agent use)
  -q, --quiet             Suppress informational output

Quick start:
  cubeapm login                                          # Configure connection
  cubeapm traces services                                # List services
  cubeapm traces search --service api-gateway --last 1h  # Search traces
  cubeapm traces get <trace-id>                          # View a trace
  cubeapm metrics query 'up'                             # Query metrics
  cubeapm logs query 'error' --last 30m                  # Query logs`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Resolve --no-input: flag > env var
		if !cmd.Flags().Changed("no-input") {
			if v := os.Getenv("CUBEAPM_NO_INPUT"); v == "1" || v == "true" {
				flagNoInput = true
			}
		}
		cmdutil.NoInput = flagNoInput

		// Resolve --quiet: flag > env var
		if !cmd.Flags().Changed("quiet") {
			if v := os.Getenv("CUBEAPM_QUIET"); v == "1" || v == "true" {
				flagQuiet = true
			}
		}
		cmdutil.Quiet = flagQuiet

		// Start a background update check for commands that should show
		// the update notice. Skip for "update" and "version" commands.
		cmdName := cmd.Name()
		parentName := ""
		if cmd.Parent() != nil {
			parentName = cmd.Parent().Name()
		}
		if cmdName != "update" && cmdName != "version" {
			startBackgroundUpdateCheck()
		}

		// Skip client setup for commands that don't need it
		if cmdName == "version" || cmdName == "help" || cmdName == "update" {
			return nil
		}
		// Config commands don't need a client
		if parentName == "config" {
			return loadConfigOnly()
		}
		if parentName == "profiles" {
			return loadConfigOnly()
		}

		if err := setupClient(cmd); err != nil {
			return err
		}

		return checkPermissions(cmd)
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Wait for the background update check and print a notice if available.
		// Skip for "update" and "version" commands, and when --quiet is set.
		cmdName := cmd.Name()
		if cmdName == "update" || cmdName == "version" {
			return nil
		}
		<-updateInfoDone
		if updateInfo != nil && !cmdutil.Quiet {
			update.PrintUpdateNotice(os.Stderr, updateInfo)
		}
		return nil
	},
}

// checkPermissions enforces read-only mode and no-input constraints on the
// current command based on resolved configuration and flag overrides.
func checkPermissions(cmd *cobra.Command) error {
	// Enforce read-only mode: flag > resolved config
	effectiveReadOnly := cmdutil.Resolved.ReadOnly
	if cmd.Flags().Changed("read-only") {
		effectiveReadOnly = flagReadOnly
	}
	if effectiveReadOnly && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return fmt.Errorf("command '%s' is blocked in read-only mode.\nTo disable, use --read-only=false or remove read_only from your config profile.", cmd.CommandPath())
	}

	// Enforce no-input mode for commands that require interactive input
	if cmdutil.NoInput && cmd.Annotations != nil && cmd.Annotations["interactive"] == "true" {
		return fmt.Errorf("command '%s' requires interactive input but --no-input is set", cmd.CommandPath())
	}

	return nil
}

func loadConfigOnly() error {
	var err error
	cmdutil.AppConfig, err = config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	flags := config.FlagOverrides{
		Output:  flagOutput,
		Profile: flagProfile,
	}
	cmdutil.Resolved = config.ResolveAuth(cmdutil.AppConfig, flags)

	cmdutil.OutputFormat, err = output.ParseFormat(cmdutil.Resolved.Output)
	if err != nil {
		return err
	}

	return nil
}

func setupClient(cmd *cobra.Command) error {
	var err error
	cmdutil.AppConfig, err = config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	flags := config.FlagOverrides{
		Server:     flagServer,
		Token:      flagToken,
		QueryPort:  flagQueryPort,
		IngestPort: flagIngestPort,
		AdminPort:  flagAdminPort,
		Output:     flagOutput,
		Profile:    flagProfile,
		Verbose:    flagVerbose,
		NoColor:    flagNoColor,
	}

	cmdutil.Resolved = config.ResolveAuth(cmdutil.AppConfig, flags)

	cmdutil.OutputFormat, err = output.ParseFormat(cmdutil.Resolved.Output)
	if err != nil {
		return err
	}

	// Login command handles client creation itself
	if cmd.Name() == "login" {
		return nil
	}

	cmdutil.APIClient, err = client.NewClient(cmdutil.Resolved)
	if err != nil {
		return err
	}

	return nil
}

// startBackgroundUpdateCheck kicks off a goroutine to check for CLI updates.
// The result is stored in updateInfo and updateInfoDone is closed when finished.
func startBackgroundUpdateCheck() {
	updateInfoOnce.Do(func() {
		go func() {
			defer close(updateInfoDone)
			info, err := update.CheckForUpdate(Version, updateRepo, config.ConfigDir())
			if err != nil {
				return
			}
			updateInfo = info
		}()
	})
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVar(&flagProfile, "profile", "", "Config profile to use")
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "CubeAPM server address")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Authentication token")
	rootCmd.PersistentFlags().IntVar(&flagQueryPort, "query-port", 0, "Query port (default 3140)")
	rootCmd.PersistentFlags().IntVar(&flagIngestPort, "ingest-port", 0, "Ingest port (default 3130)")
	rootCmd.PersistentFlags().IntVar(&flagAdminPort, "admin-port", 0, "Admin port (default 3199)")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&flagReadOnly, "read-only", false, "Block write operations (safety mode for agents)")
	rootCmd.PersistentFlags().BoolVar(&flagNoInput, "no-input", false, "Disable all interactive prompts (for CI/agent use)")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress informational output")

	// Register subcommands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(cmdconfig.NewConfigCmd())
	rootCmd.AddCommand(cmdtraces.NewTracesCmd())
	rootCmd.AddCommand(cmdmetrics.NewMetricsCmd())
	rootCmd.AddCommand(cmdlogs.NewLogsCmd())
	rootCmd.AddCommand(cmdingest.NewIngestCmd())
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		output.WriteError(os.Stderr, string(cmdutil.OutputFormat), err, 0)
		return err
	}
	return nil
}
