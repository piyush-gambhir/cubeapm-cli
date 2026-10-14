package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/output"
)

// resolution describes arguments using the real parser without running a command.
type resolution struct {
	Command     string            `json:"command"`
	Effect      string            `json:"effect"`
	Conditional bool              `json:"conditional"`
	Interactive bool              `json:"interactive"`
	Secret      bool              `json:"secret_output"`
	Flags       map[string]string `json:"flags"`
	Args        []string          `json:"args"`
}

func newCommandsCmd() *cobra.Command {
	commands := &cobra.Command{Use: "commands", Short: "Offline command discovery for agents"}
	commands.AddCommand(&cobra.Command{
		Use:                "resolve -- <arguments>",
		Short:              "Show the command, effect and flags that arguments would run, without running them",
		Example:            "  cubeapm commands resolve -- traces services\n  cubeapm commands resolve -- -o json ingest logs --file logs.jsonl",
		DisableFlagParsing: true,
		// Override root hooks to avoid config, credentials, and update checks.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.OutputFormat = output.FormatJSON
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error { return nil },
		RunE: func(cmd *cobra.Command, args []string) error {
			for i, arg := range args {
				if arg == "--" {
					args = args[i+1:]
					break
				}
			}
			probe := cmd.Root() // nothing executes and the process exits after answering
			target, rest, err := probe.Find(args)
			if err != nil || target == probe || !target.Runnable() {
				return fmt.Errorf("arguments do not name a runnable cubeapm command")
			}
			if err := target.ParseFlags(rest); err != nil {
				return err
			}
			res := resolveCommand(target)
			return output.NewFormatter(cmdutil.OutputFormat, false).Format(cmd.OutOrStdout(), res)
		},
	})
	return commands
}

func resolveCommand(target *cobra.Command) resolution {
	res := resolution{Command: target.CommandPath(), Effect: "read", Flags: map[string]string{}, Args: target.Flags().Args()}
	if res.Args == nil {
		res.Args = []string{}
	}
	ann := target.Annotations
	flagSet := func(name string) bool { return name != "" && target.Flags().Changed(name) }
	switch {
	case ann["mutates-when-flag"] != "":
		if flagSet(ann["mutates-when-flag"]) {
			res.Effect = "remote_write"
		}
	case ann["mutates"] == "true" || ann["mutates"] == "conditional":
		res.Effect, res.Conditional = "remote_write", ann["mutates"] == "conditional"
	case ann["writes-local"] == "true" || ann["writes-local"] == "conditional" || ann["self-update"] != "" || flagSet(ann["writes-local-when-flag"]):
		res.Effect, res.Conditional = "local_write", ann["writes-local"] == "conditional"
	}
	res.Interactive = ann["interactive"] == "true"
	res.Secret = ann["secret-output"] == "true"
	target.Flags().Visit(func(f *pflag.Flag) { res.Flags[f.Name] = f.Value.String() })
	return res
}
