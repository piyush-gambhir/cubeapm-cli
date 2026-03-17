package traces

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

func newGetCmd() *cobra.Command {
	var (
		from string
		to   string
		last string
	)

	cmd := &cobra.Command{
		Use:   "get <trace-id>",
		Short: "Get a trace by ID",
		Long:  "Retrieve and display a trace. In table mode, renders a waterfall view of spans.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			traceID := args[0]

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			trace, err := cmdutil.APIClient.GetTrace(traceID, start, end)
			if err != nil {
				return err
			}

			// For non-table output, return the raw trace data
			if cmdutil.OutputFormat != output.FormatTable {
				return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, trace)
			}

			// Table mode: render waterfall
			return renderWaterfall(os.Stdout, trace, cmdutil.Resolved.NoColor)
		},
	}

	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}

// spanNode is used to build the span tree.
type spanNode struct {
	span     types.Span
	service  string
	children []*spanNode
}

func renderWaterfall(w io.Writer, trace *types.Trace, noColor bool) error {
	if len(trace.Spans) == 0 {
		fmt.Fprintln(w, "Trace has no spans.")
		return nil
	}

	// Calculate total duration from root span
	var totalDuration int64
	for _, s := range trace.Spans {
		if s.Duration > totalDuration {
			totalDuration = s.Duration
		}
	}

	// Print trace header
	if !noColor {
		fmt.Fprintf(w, "\033[1mTRACE: %s (%d spans, %s)\033[0m\n\n", trace.TraceID, len(trace.Spans), formatDuration(totalDuration))
	} else {
		fmt.Fprintf(w, "TRACE: %s (%d spans, %s)\n\n", trace.TraceID, len(trace.Spans), formatDuration(totalDuration))
	}

	// Build span lookup and tree
	nodeMap := make(map[string]*spanNode, len(trace.Spans))
	for _, s := range trace.Spans {
		service := ""
		if proc, ok := trace.Processes[s.ProcessID]; ok {
			service = proc.ServiceName
		}
		nodeMap[s.SpanID] = &spanNode{
			span:    s,
			service: service,
		}
	}

	// Find roots and build parent-child relationships
	var roots []*spanNode
	for _, s := range trace.Spans {
		node := nodeMap[s.SpanID]
		parentFound := false
		for _, ref := range s.References {
			if ref.RefType == "CHILD_OF" {
				if parent, ok := nodeMap[ref.SpanID]; ok {
					parent.children = append(parent.children, node)
					parentFound = true
					break
				}
			}
		}
		if !parentFound {
			roots = append(roots, node)
		}
	}

	// Sort roots and children by start time
	sortNodes(roots)
	for _, node := range nodeMap {
		sortNodes(node.children)
	}

	// Flatten tree into rows
	type flatRow struct {
		prefix    string
		service   string
		operation string
		duration  string
		status    string
	}

	var rows []flatRow
	var flatten func(node *spanNode, prefix string, isLast bool, depth int)
	flatten = func(node *spanNode, prefix string, isLast bool, depth int) {
		var linePrefix string
		if depth == 0 {
			linePrefix = ""
		} else {
			if isLast {
				linePrefix = prefix + "\u2514\u2500 "
			} else {
				linePrefix = prefix + "\u251C\u2500 "
			}
		}

		status := "OK"
		for _, tag := range node.span.Tags {
			if tag.Key == "otel.status_code" || tag.Key == "status.code" {
				status = fmt.Sprintf("%v", tag.Value)
			}
			if tag.Key == "error" && fmt.Sprintf("%v", tag.Value) == "true" {
				status = "ERROR"
			}
		}

		rows = append(rows, flatRow{
			prefix:    linePrefix,
			service:   node.service,
			operation: node.span.OperationName,
			duration:  formatDuration(node.span.Duration),
			status:    status,
		})

		var childPrefix string
		if depth == 0 {
			childPrefix = ""
		} else {
			if isLast {
				childPrefix = prefix + "   "
			} else {
				childPrefix = prefix + "\u2502  "
			}
		}

		for i, child := range node.children {
			isChildLast := i == len(node.children)-1
			flatten(child, childPrefix, isChildLast, depth+1)
		}
	}

	for i, root := range roots {
		flatten(root, "", i == len(roots)-1, 0)
	}

	// Calculate max widths
	maxService := len("SERVICE")
	maxOp := len("OPERATION")
	maxDur := len("DURATION")

	for _, r := range rows {
		combined := displayWidth(r.prefix) + len(r.service)
		if combined > maxService {
			maxService = combined
		}
		if len(r.operation) > maxOp {
			maxOp = len(r.operation)
		}
		if len(r.duration) > maxDur {
			maxDur = len(r.duration)
		}
	}

	// Print header
	hdr := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
		maxService, "SERVICE",
		maxOp, "OPERATION",
		maxDur, "DURATION",
		"STATUS")
	if !noColor {
		fmt.Fprintf(w, "\033[1m%s\033[0m\n", hdr)
	} else {
		fmt.Fprintln(w, hdr)
	}

	// Print rows
	for _, r := range rows {
		svc := r.prefix + r.service
		// Pad using display width
		padding := maxService - displayWidth(svc)
		if padding < 0 {
			padding = 0
		}
		paddedSvc := svc + strings.Repeat(" ", padding)

		statusStr := r.status
		if !noColor && r.status == "ERROR" {
			statusStr = "\033[31m" + r.status + "\033[0m"
		}

		fmt.Fprintf(w, "%s  %-*s  %-*s  %s\n",
			paddedSvc,
			maxOp, r.operation,
			maxDur, r.duration,
			statusStr)
	}

	return nil
}

// displayWidth returns the visible width of a string (accounting for multi-byte
// Unicode characters used in tree drawing).
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		// Box drawing characters are single width
		w++
		_ = r
	}
	return w
}

func sortNodes(nodes []*spanNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].span.StartTime < nodes[j].span.StartTime
	})
}
