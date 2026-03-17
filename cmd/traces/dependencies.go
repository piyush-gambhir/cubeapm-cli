package traces

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

func newDependenciesCmd() *cobra.Command {
	var (
		from      string
		to        string
		last      string
		outputDot bool
	)

	cmd := &cobra.Command{
		Use:     "dependencies",
		Short:   "Show service dependencies",
		Long:    "Display dependency graph between services. Supports --dot for Graphviz DOT format.",
		Aliases: []string{"deps"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			deps, err := cmdutil.APIClient.GetDependencies(start, end)
			if err != nil {
				return err
			}

			if len(deps) == 0 {
				fmt.Println("No dependencies found.")
				return nil
			}

			// DOT format output
			if outputDot {
				return renderDOT(deps)
			}

			table := output.TableDef{
				Headers: []string{"PARENT", "CHILD", "CALL_COUNT"},
			}
			for _, d := range deps {
				table.Rows = append(table.Rows, []string{
					d.Parent,
					d.Child,
					strconv.FormatInt(d.CallCount, 10),
				})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	timeflag.AddTimeFlags(cmd, &from, &to, &last)
	cmd.Flags().BoolVar(&outputDot, "dot", false, "Output in Graphviz DOT format")

	return cmd
}

func renderDOT(deps []types.Dependency) error {
	w := os.Stdout
	fmt.Fprintln(w, "digraph dependencies {")
	fmt.Fprintln(w, "  rankdir=LR;")
	fmt.Fprintln(w, "  node [shape=box, style=rounded];")
	fmt.Fprintln(w)

	// Collect unique services
	services := make(map[string]bool)
	for _, d := range deps {
		services[d.Parent] = true
		services[d.Child] = true
	}

	// Write node definitions
	for svc := range services {
		fmt.Fprintf(w, "  %q;\n", svc)
	}
	fmt.Fprintln(w)

	// Write edges
	for _, d := range deps {
		fmt.Fprintf(w, "  %q -> %q [label=%q];\n", d.Parent, d.Child, strconv.FormatInt(d.CallCount, 10))
	}

	fmt.Fprintln(w, "}")
	return nil
}
