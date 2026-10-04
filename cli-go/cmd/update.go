package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/update"
)

const updateRepo = "piyush-gambhir/cubeapm-cli"

// stdinIsTerminal decides whether update may prompt. Tests replace it.
var stdinIsTerminal = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }

// updateCheckResult is the `update --check -o json|yaml` document.
type updateCheckResult struct {
	CurrentVersion  string `json:"current_version" yaml:"current_version"`
	LatestVersion   string `json:"latest_version" yaml:"latest_version"`
	UpdateAvailable bool   `json:"update_available" yaml:"update_available"`
	ReleaseURL      string `json:"release_url" yaml:"release_url"`
	InstallMethod   string `json:"install_method" yaml:"install_method"`
}

func newUpdateCmd() *cobra.Command {
	var checkOnly, yes bool

	cmd := &cobra.Command{
		Use:         "update",
		Annotations: map[string]string{"mutates": "true"},
		Short:       "Update cubeapm to the latest version",
		Long: `Check for and install the latest release of the cubeapm CLI on macOS, Linux,
and Windows.

update downloads this platform's release archive from GitHub, verifies its
SHA-256 checksum against the release's checksums.txt, and replaces the running
binary. On Windows the old binary is renamed to cubeapm.exe.old and deleted on a
later run. If the binary's directory is not writable, update fails and leaves
the current binary in place: re-run with sudo, or reinstall with the install
script into a writable directory. A binary that make install put in a Go bin
directory ($GOBIN, $GOPATH/bin, or ~/go/bin) is not replaced; update prints the
source-build command instead.

update asks "Update now? [Y/n]" in a terminal. --yes skips the question; with
--no-input or without a terminal, update fails unless --yes is given.

--check only reports the current and latest versions, always asking GitHub
(bypassing the daily cache) and storing the answer in that cache. Use -o json for the fields current_version,
latest_version, update_available, release_url, and install_method (self or go).
Read-only mode blocks update but allows update --check.

In an interactive terminal, other commands check GitHub for a new release at
most once a day and print a notice on stderr after their output, at most once a
day per release. On the run that checks, the command waits at most 1 second for
GitHub, so a fast command still shows the notice. There is no check when stderr is not a terminal, when CI is
set, with --quiet or CUBEAPM_QUIET, or when CUBEAPM_NO_UPDATE_NOTIFIER or
NO_UPDATE_NOTIFIER is set.

Examples:
  cubeapm update --check
  cubeapm update --check -o json
  cubeapm update
  cubeapm update --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, checkOnly, yes)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, do not install")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Install without asking for confirmation")
	return cmd
}

func runUpdate(cmd *cobra.Command, checkOnly, yes bool) error {
	out := cmd.OutOrStdout()
	if !update.IsReleaseVersion(Version) {
		return fmt.Errorf("cannot check for updates: %q is a development build; install a release with the install script or from https://github.com/%s/releases", Version, updateRepo)
	}

	// FetchLatest stores the answer in the notifier's cache, so a later notice
	// agrees with this check (and, after an install, stays quiet).
	info, err := update.FetchLatest(cmd.Context(), Version, updateRepo, config.ConfigDir(), update.CommandTimeout)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	execPath, execErr := update.ExecutablePath()
	method := update.InstallSelf
	if execErr == nil {
		method = update.DetectInstallMethod(execPath)
	}

	if checkOnly {
		return printUpdateCheck(out, info, method)
	}

	if !info.Available {
		if !cmdutil.Quiet {
			fmt.Fprintf(out, "cubeapm v%s is already the latest version.\n", info.CurrentVersion)
		}
		return nil
	}
	if method == update.InstallGo {
		fmt.Fprintf(out, "cubeapm v%s -> v%s is available. This binary was built from source into a Go bin directory, so update it with:\n  %s\nRelease notes: %s\n",
			info.CurrentVersion, info.LatestVersion, update.SourceUpdateCommand, info.ReleaseURL)
		return nil
	}
	if execErr != nil {
		return execErr
	}

	if !cmdutil.Quiet {
		fmt.Fprintf(out, "Update available: v%s -> v%s\n", info.CurrentVersion, info.LatestVersion)
	}
	if !yes {
		if cmdutil.NoInput || !stdinIsTerminal() {
			return fmt.Errorf("installing v%s needs confirmation: pass --yes to update without a prompt", info.LatestVersion)
		}
		if !confirmUpdate(cmd.InOrStdin(), out) {
			fmt.Fprintln(out, "Update cancelled.")
			return nil
		}
	}

	progress := out
	if cmdutil.Quiet {
		progress = io.Discard
	}
	if err := update.Install(cmd.Context(), updateRepo, info.LatestVersion, execPath, progress); err != nil {
		return fmt.Errorf("update failed, cubeapm v%s is still installed: %w", info.CurrentVersion, err)
	}
	fmt.Fprintf(out, "Updated cubeapm v%s -> v%s\nRelease notes: %s\n", info.CurrentVersion, info.LatestVersion, info.ReleaseURL)
	return nil
}

func printUpdateCheck(out io.Writer, info *update.UpdateInfo, method string) error {
	result := updateCheckResult{
		CurrentVersion:  info.CurrentVersion,
		LatestVersion:   info.LatestVersion,
		UpdateAvailable: info.Available,
		ReleaseURL:      info.ReleaseURL,
		InstallMethod:   method,
	}
	if cmdutil.OutputFormat == output.FormatJSON || cmdutil.OutputFormat == output.FormatYAML {
		return output.NewFormatter(cmdutil.OutputFormat, true).Format(out, result)
	}

	available := "no"
	if info.Available {
		available = "yes"
	}
	fmt.Fprintf(out, "Current version:  v%s\n", info.CurrentVersion)
	fmt.Fprintf(out, "Latest version:   v%s\n", info.LatestVersion)
	fmt.Fprintf(out, "Update available: %s\n", available)
	fmt.Fprintf(out, "Release notes:    %s\n", info.ReleaseURL)
	if info.Available {
		fmt.Fprintf(out, "Update with:      %s\n", update.UpdateCommand(method))
	}
	return nil
}

// confirmUpdate asks "Update now? [Y/n]"; an empty answer means yes.
func confirmUpdate(in io.Reader, out io.Writer) bool {
	fmt.Fprint(out, "Update now? [Y/n] ")
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		fmt.Fprintln(out)
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "y", "yes":
		return true
	}
	return false
}
