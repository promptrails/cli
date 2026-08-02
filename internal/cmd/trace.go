package cmd

import (
	"fmt"

	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	traceKind     string
	tracePage     int
	traceDateFrom string
	traceDateTo   string
	traceStatus   string
	traceAgentID  string
	traceModel    string
)

func init() {
	traceListCmd.Flags().StringVar(&traceKind, "kind", "", "Filter by span kind (e.g. llm, prompt, agent, tool)")
	traceListCmd.Flags().IntVar(&tracePage, "page", 1, "Page number")

	traceSummaryCmd.Flags().StringVar(&traceDateFrom, "from", "", "Start date (RFC3339 or YYYY-MM-DD)")
	traceSummaryCmd.Flags().StringVar(&traceDateTo, "to", "", "End date (RFC3339 or YYYY-MM-DD)")
	traceSummaryCmd.Flags().StringVar(&traceStatus, "status", "", "Filter by status")
	traceSummaryCmd.Flags().StringVar(&traceKind, "kind", "", "Filter by span kind")
	traceSummaryCmd.Flags().StringVar(&traceAgentID, "agent", "", "Filter by agent ID")
	traceSummaryCmd.Flags().StringVar(&traceModel, "model", "", "Filter by model name")

	traceCmd.AddCommand(traceSummaryCmd)
	traceCmd.AddCommand(traceListCmd)
	traceCmd.AddCommand(traceGetCmd)
	rootCmd.AddCommand(traceCmd)
}

var traceCmd = &cobra.Command{
	Use:     "trace",
	Aliases: []string{"traces"},
	Short:   "Inspect observability traces and usage/cost summaries",
}

var traceSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show aggregate token/cost/latency statistics over filtered traces",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		summary, err := client.Traces.GetSummary(cmdContext(), &promptrails.TraceFilterParams{
			DateFrom:  traceDateFrom,
			DateTo:    traceDateTo,
			Status:    traceStatus,
			Kind:      traceKind,
			AgentID:   traceAgentID,
			ModelName: traceModel,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(summary)
		}

		fmt.Println()
		output.KeyValue("Total Traces", fmt.Sprintf("%d", summary.TotalTraces))
		output.KeyValue("Total Tokens", fmt.Sprintf("%d", summary.TotalTokens))
		output.KeyValue("Total Cost", fmt.Sprintf("$%.4f", summary.TotalCost))
		output.KeyValue("Avg Duration", fmt.Sprintf("%.0fms", summary.AvgDurationMs))
		output.KeyValue("Errors", fmt.Sprintf("%d", summary.ErrorCount))
		output.KeyValue("Unique Models", fmt.Sprintf("%d", summary.UniqueModels))
		output.KeyValue("Unique Sessions", fmt.Sprintf("%d", summary.UniqueSessions))
		fmt.Println()
		return nil
	},
}

var traceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List trace spans",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Traces.List(cmdContext(), &promptrails.ListTracesParams{
			Page:  tracePage,
			Limit: 20,
			Kind:  traceKind,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		printTraceTable(resp.Data)

		if resp.Meta.Pages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.Pages, resp.Meta.Total))
		}
		return nil
	},
}

var traceGetCmd = &cobra.Command{
	Use:   "get <trace-id>",
	Short: "Get all spans for a trace ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		spans, err := client.Traces.GetByTraceID(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(spans)
		}

		printTraceTable(spans)
		return nil
	},
}

// printTraceTable renders trace spans as a table.
func printTraceTable(spans []promptrails.Trace) {
	var rows [][]string
	for _, t := range spans {
		rows = append(rows, []string{
			t.ID,
			t.Name,
			t.Kind,
			t.Status,
			t.ModelName,
			fmt.Sprintf("%dms", t.DurationMs),
			t.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	output.Table([]string{"ID", "NAME", "KIND", "STATUS", "MODEL", "DURATION", "CREATED"}, rows)
}
