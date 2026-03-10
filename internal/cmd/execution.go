package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	execAgentID string
	execPage    int
)

func init() {
	execListCmd.Flags().StringVar(&execAgentID, "agent", "", "Filter by agent ID")
	execListCmd.Flags().IntVar(&execPage, "page", 1, "Page number")

	execCmd.AddCommand(execListCmd)
	execCmd.AddCommand(execGetCmd)
	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:     "execution",
	Aliases: []string{"exec"},
	Short:   "View executions",
}

var execListCmd = &cobra.Command{
	Use:   "list",
	Short: "List executions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Executions.List(cmdContext(), &promptrails.ListExecutionsParams{
			Page:    execPage,
			Limit:   20,
			AgentID: execAgentID,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, e := range resp.Data {
			rows = append(rows, []string{
				e.ID,
				e.AgentName,
				e.Status,
				fmt.Sprintf("%dms", e.DurationMs),
				fmt.Sprintf("%d", e.TotalTokens),
				e.CreatedAt.Format("2006-01-02 15:04"),
			})
		}
		output.Table([]string{"ID", "AGENT", "STATUS", "DURATION", "TOKENS", "CREATED"}, rows)

		if resp.Meta.TotalPages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total))
		}
		return nil
	},
}

var execGetCmd = &cobra.Command{
	Use:   "get <execution-id>",
	Short: "Get execution details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		exec, err := client.Executions.Get(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(exec)
		}

		fmt.Println()
		output.KeyValue("ID", exec.ID)
		output.KeyValue("Agent", exec.AgentName)
		output.KeyValue("Status", exec.Status)
		output.KeyValue("Duration", fmt.Sprintf("%dms", exec.DurationMs))
		output.KeyValue("Tokens", fmt.Sprintf("%d", exec.TotalTokens))
		output.KeyValue("Created", exec.CreatedAt.Format("2006-01-02 15:04:05"))
		if exec.Error != nil && *exec.Error != "" {
			output.KeyValue("Error", *exec.Error)
		}

		if exec.Input != nil {
			inputJSON, _ := json.MarshalIndent(exec.Input, "  ", "  ")
			output.KeyValue("Input", string(inputJSON))
		}
		if exec.Output != nil {
			outputJSON, _ := json.MarshalIndent(exec.Output, "  ", "  ")
			output.KeyValue("Output", string(outputJSON))
		}
		fmt.Println()
		return nil
	},
}
