package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	cmdconfig "github.com/piyush-gambhir/cubeapm-cli/cli-go/cmd/config"
	cmdingest "github.com/piyush-gambhir/cubeapm-cli/cli-go/cmd/ingest"
	cmdlogs "github.com/piyush-gambhir/cubeapm-cli/cli-go/cmd/logs"
	cmdmetrics "github.com/piyush-gambhir/cubeapm-cli/cli-go/cmd/metrics"
	cmdtraces "github.com/piyush-gambhir/cubeapm-cli/cli-go/cmd/traces"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/update"
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
	flagEmail      string
	flagPassword   string
	flagQueryPort  int
	flagIngestPort int
	flagAdminPort  int
	flagNoColor    bool
	flagVerbose    bool
	flagReadOnly   bool
	flagNoInput    bool
	flagQuiet      bool
)

// updateCheck is the release check started by PersistentPreRunE. done is
// closed once info is final; info stays nil when nothing is known.
type updateCheck struct {
	done chan struct{}
	info *update.UpdateInfo
}

var (
	pendingUpdateCheck *updateCheck

	// updateNoticeWait bounds how long a command waits, at most once a day,
	// for the release check it started.
	updateNoticeWait = time.Second

	// Test seams for the update notifier.
	stderrIsTerminal           = func() bool { return term.IsTerminal(int(os.Stderr.Fd())) }
	noticeOutput     io.Writer = os.Stderr
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
  --email <email>         Override login email
  --password <password>   Override login password
  --profile <name>        Use a specific connection profile
  --query-port <port>     Override query port (default: 3140)
  --ingest-port <port>    Override ingest port (default: 3130)
  --admin-port <port>     Override admin port (default: 3199)
  --no-color              Disable colored output
  --verbose               Enable verbose HTTP request logging
  --read-only             Block write commands (cannot be turned off once set)
  --no-input              Disable all interactive prompts (for CI/agent use)
  -q, --quiet             Suppress informational output

Quick start:
  cubeapm login                                          # Configure connection
  cubeapm traces services                                # List services
  cubeapm traces search --service api-gateway --last 1h  # Search traces
  cubeapm traces get <trace-id>                          # View a trace
  cubeapm metrics query 'up'                             # Query metrics
  cubeapm logs query 'error' --last 30m                  # Query logs

Full command reference (for agents/LLMs): https://github.com/piyush-gambhir/cubeapm-cli/blob/main/docs/llms.txt
Claude Code skill: https://github.com/piyush-gambhir/cubeapm-cli/blob/main/cubeapm/SKILL.md`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Agent-safety environment settings are sticky: an explicit false flag
		// must not silently re-enable prompts in automation.
		if v := os.Getenv("CUBEAPM_NO_INPUT"); v == "1" || v == "true" {
			flagNoInput = true
		}
		cmdutil.NoInput = flagNoInput

		// Resolve --quiet: flag > env var
		if !cmd.Flags().Changed("quiet") {
			if v := os.Getenv("CUBEAPM_QUIET"); v == "1" || v == "true" {
				flagQuiet = true
			}
		}
		cmdutil.Quiet = flagQuiet

		if updateNotifierEnabled(cmd) {
			startUpdateCheck()
		}

		cmdName := cmd.Name()
		parentName := ""
		if cmd.Parent() != nil {
			parentName = cmd.Parent().Name()
		}

		// Skip client setup for commands that don't need it
		if cmdName == "version" || cmdName == "help" {
			return nil
		}
		// Config commands and update don't need a client, but they still load
		// the profile so its read_only setting applies to them.
		if cmdName == "update" || parentName == "config" || parentName == "profiles" {
			if err := loadConfigOnly(); err != nil {
				return err
			}
			// Config writes change the active profile whatever --profile
			// names, and update has no profile target, so --profile must not
			// sidestep the active profile's read_only.
			if active, ok := cmdutil.AppConfig.Profiles[cmdutil.AppConfig.CurrentProfile]; ok && active.ReadOnly {
				cmdutil.Resolved.ReadOnly = true
			}
		} else if err := setupClient(cmd); err != nil {
			return err
		}

		return checkPermissions(cmd)
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		showUpdateNotice()
		return nil
	},
}

// checkPermissions enforces read-only mode and no-input constraints on the
// current command based on resolved configuration and flag overrides.
func checkPermissions(cmd *cobra.Command) error {
	// A read-only profile or environment setting is sticky; the flag can only
	// add protection, never silently remove it for one command.
	effectiveReadOnly := cmdutil.Resolved.ReadOnly
	if flagReadOnly {
		effectiveReadOnly = true
	}
	if effectiveReadOnly && isWrite(cmd) {
		return fmt.Errorf("command '%s' is blocked in read-only mode; to permit writes, drop --read-only, unset CUBEAPM_READ_ONLY, and remove read_only from the profile", cmd.CommandPath())
	}

	// Enforce no-input mode for commands that require interactive input
	if cmdutil.NoInput && cmd.Annotations != nil && cmd.Annotations["interactive"] == "true" {
		return fmt.Errorf("command '%s' requires interactive input but --no-input is set", cmd.CommandPath())
	}

	return nil
}

// isWrite reports whether cmd is annotated as mutating. `update --check` only
// reports whether a release exists, so read-only mode still allows it.
func isWrite(cmd *cobra.Command) bool {
	if cmd.Annotations["mutates"] != "true" {
		return false
	}
	if cmd.CommandPath() == "cubeapm update" {
		if check, err := cmd.Flags().GetBool("check"); err == nil && check {
			return false
		}
	}
	return true
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
	if cmd.Name() != "login" {
		if err := config.ValidateSelectedProfile(cmdutil.AppConfig, flagProfile); err != nil {
			return err
		}
	}

	flags := config.FlagOverrides{
		Server:     flagServer,
		Email:      flagEmail,
		Password:   flagPassword,
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

	// Wire up session refresh callback so re-auth persists the new cookie
	cmdutil.APIClient.SetOnSessionRefresh(func(cookie, expiry string) error {
		profileName := cmdutil.AppConfig.CurrentProfile
		if flagProfile != "" {
			profileName = flagProfile
		}
		if profileName == "" {
			return nil
		}
		return config.Update(func(cfg *config.Config) error {
			p, ok := cfg.Profiles[profileName]
			if !ok {
				return nil
			}
			p.SessionCookie = cookie
			p.SessionExpiry = expiry
			cfg.Profiles[profileName] = p
			return nil
		})
	})

	return nil
}

// updateNotifierEnabled reports whether this run may look for a new release.
// Scripts, CI, agents, quiet runs, and development builds never do, and
// neither do commands that report versions or complete shell input.
func updateNotifierEnabled(cmd *cobra.Command) bool {
	if cmdutil.Quiet || !update.IsReleaseVersion(Version) || !stderrIsTerminal() {
		return false
	}
	for _, name := range []string{"CI", "CUBEAPM_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		if os.Getenv(name) != "" {
			return false
		}
	}
	for c := cmd; c != nil; c = c.Parent() {
		switch name := c.Name(); {
		case name == "update", name == "version", name == "completion", name == "help",
			strings.HasPrefix(name, "__complete"):
			return false
		}
	}
	return true
}

// startUpdateCheck answers from a fresh cache right away; otherwise it asks
// GitHub in the background so the command never waits on the network.
func startUpdateCheck() {
	check := &updateCheck{done: make(chan struct{})}
	pendingUpdateCheck = check
	configDir := config.ConfigDir()
	if info, ok := update.CachedCheck(Version, updateRepo, configDir, time.Now()); ok {
		check.info = info
		close(check.done)
		return
	}
	// Record the attempt first, so a command that exits before the answer
	// arrives does not leave the next command to ask GitHub again. Without a
	// record there is no once-a-day limit, so skip the check.
	if err := update.RecordCheckAttempt(configDir, time.Now()); err != nil {
		close(check.done)
		return
	}
	go func() {
		defer close(check.done)
		info, err := update.FetchLatest(context.Background(), Version, updateRepo, configDir, update.BackgroundTimeout)
		if err == nil {
			check.info = info
		}
	}()
}

// showUpdateNotice prints the notice after the command's output. A cached
// answer is ready at once. A request this run started gets at most
// updateNoticeWait to finish: the attempt is already recorded, so a fast
// command that exits first would otherwise lose the day's only check.
func showUpdateNotice() {
	check := pendingUpdateCheck
	if check == nil {
		return
	}
	select {
	case <-check.done:
	case <-time.After(updateNoticeWait):
		return
	}
	if check.info == nil || !check.info.Available {
		return
	}
	method := update.InstallSelf
	if exe, err := update.ExecutablePath(); err == nil {
		method = update.DetectInstallMethod(exe)
	}
	update.Notify(noticeOutput, config.ConfigDir(), check.info, update.UpdateCommand(method), time.Now())
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVar(&flagProfile, "profile", "", "Config profile to use")
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "CubeAPM server address")
	rootCmd.PersistentFlags().StringVar(&flagEmail, "email", "", "Login email")
	rootCmd.PersistentFlags().StringVar(&flagPassword, "password", "", "Login password")
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
	update.RemoveOldExecutable()
	if err := rootCmd.Execute(); err != nil {
		output.WriteError(os.Stderr, string(cmdutil.OutputFormat), err, 0)
		return err
	}
	return nil
}
