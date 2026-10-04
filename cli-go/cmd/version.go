package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/update"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Long: `Print the CubeAPM CLI version, commit hash, and build date.

When an earlier command already checked GitHub for releases, version also
prints the latest known release and whether an update is available. It reads
that from the local cache and never uses the network.

Examples:
  cubeapm version`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "cubeapm version %s\n", Version)
			fmt.Fprintf(out, "  commit:     %s\n", Commit)
			fmt.Fprintf(out, "  built:      %s\n", BuildDate)
			if info := update.CachedUpdateInfo(Version, updateRepo, config.ConfigDir()); info != nil {
				fmt.Fprintf(out, "  latest:     %s\n", info.LatestVersion)
				fmt.Fprintf(out, "  update_available: %t\n", info.Available)
			}
		},
	}
}
