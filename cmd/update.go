package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/internal/update"
)

const updateRepo = "piyush-gambhir/cubeapm-cli"

func newUpdateCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update cubeapm to the latest version",
		Long: `Check for and install the latest version of the cubeapm CLI.

Use --check to only check whether an update is available without installing it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(checkOnly)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, do not install")
	return cmd
}

func runUpdate(checkOnly bool) error {
	currentVersion := Version
	if currentVersion == "" || currentVersion == "dev" {
		return fmt.Errorf("cannot check for updates: running a development build")
	}

	if !cmdutil.Quiet {
		fmt.Printf("Current version: v%s\n", currentVersion)
		fmt.Println("Checking for updates...")
	}

	info, err := update.CheckForUpdateFresh(currentVersion, updateRepo, config.ConfigDir())
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	if !info.Available {
		if !cmdutil.Quiet {
			fmt.Printf("Already up to date (v%s)\n", currentVersion)
		}
		return nil
	}

	if !cmdutil.Quiet {
		fmt.Printf("\nNew version available: v%s -> v%s\n", info.CurrentVersion, info.LatestVersion)
		if info.PublishedAt != "" {
			fmt.Printf("Published: %s\n", info.PublishedAt)
		}
		fmt.Printf("Release:   %s\n", info.ReleaseURL)
	}

	if checkOnly {
		if !cmdutil.Quiet {
			fmt.Printf("\nRun `cubeapm update` to install the update.\n")
		}
		return nil
	}

	if cmdutil.NoInput {
		return fmt.Errorf("update requires confirmation; cannot run with --no-input (use --check to check only)")
	}

	fmt.Println()
	if !update.ConfirmPrompt("Do you want to update?") {
		fmt.Println("Update cancelled.")
		return nil
	}

	fmt.Println()
	if err := update.SelfUpdate(info.LatestVersion, updateRepo); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

