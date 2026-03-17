package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cubeapm version %s\n", Version)
			fmt.Printf("  commit:     %s\n", Commit)
			fmt.Printf("  built:      %s\n", BuildDate)
		},
	}
}
