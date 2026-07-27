package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	execAgentID  string
	execStatus   string
	execPage     int
	decideReason string
)

func init() {
	execListCmd.Flags().StringVar(&execAgentID, "agent", "", "Filter by agent ID")
	execListCmd.Flags().StringVar(&execStatus, "status", "", "Filter by status (e.g. completed, failed, running, waiting_approval)")
	execListCmd.Flags().IntVar(&execPage, "page", 1, "Page number")

	inboxCmd.Flags().IntVar(&execPage, "page", 1, "Page number")

	execApproveCmd.Flags().StringVar(&decideReason, "reason", "", "Optional reason recorded with the decision")
	execDenyCmd.Flags().StringVar(&decideReason, "reason", "", "Optional reason recorded with the decision")

	execCmd.AddCommand(execListCmd)
	execCmd.AddCommand(execGetCmd)
	execCmd.AddCommand(execTreeCmd)
	execCmd.AddCommand(execCancelCmd)
	execCmd.AddCommand(inboxCmd)
	execCmd.AddCommand(execApproveCmd)
	execCmd.AddCommand(execDenyCmd)
	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:     "execution",
	Aliases: []string{"exec"},
	Short:   "View executions and resolve approval-gated runs",
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
			Status:  execStatus,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		printExecutionTable(resp.Data)

		if resp.Meta.Pages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.Pages, resp.Meta.Total))
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

		printExecutionDetails(exec)
		return nil
	},
}

var execTreeCmd = &cobra.Command{
	Use:   "tree <execution-id>",
	Short: "Show an execution with its full sub-execution tree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		exec, err := client.Executions.Tree(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(exec)
		}

		fmt.Println()
		printExecutionTree(exec, 0)
		fmt.Println()
		return nil
	},
}

var execCancelCmd = &cobra.Command{
	Use:   "cancel <execution-id>",
	Short: "Request cooperative cancellation of a running execution",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		exec, err := client.Executions.Cancel(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(exec)
		}

		output.Success(fmt.Sprintf("Cancellation requested — status: %s", exec.Status))
		return nil
	},
}

var inboxCmd = &cobra.Command{
	Use:     "approval-inbox",
	Aliases: []string{"inbox"},
	Short:   "List executions parked at waiting_approval",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Executions.ApprovalInbox(cmdContext(), &promptrails.ListParams{
			Page:  execPage,
			Limit: 20,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, e := range resp.Data {
			expires := ""
			if e.ApprovalExpiresAt != nil {
				expires = e.ApprovalExpiresAt.Format("2006-01-02 15:04")
			}
			rows = append(rows, []string{e.ID, e.AgentName, e.Status, expires, e.CreatedAt.Format("2006-01-02 15:04")})
		}
		output.Table([]string{"ID", "AGENT", "STATUS", "APPROVAL EXPIRES", "CREATED"}, rows)

		if resp.Meta.Pages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.Pages, resp.Meta.Total))
		}
		return nil
	},
}

var execApproveCmd = &cobra.Command{
	Use:   "approve <execution-id>",
	Short: "Approve a run parked at waiting_approval and resume it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		exec, err := client.Executions.Approve(cmdContext(), args[0], &promptrails.DecideParams{Reason: decideReason})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(exec)
		}

		output.Success(fmt.Sprintf("Approved — status: %s", exec.Status))
		return nil
	},
}

var execDenyCmd = &cobra.Command{
	Use:   "deny <execution-id>",
	Short: "Deny a run parked at waiting_approval and resume with a denial",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		exec, err := client.Executions.Deny(cmdContext(), args[0], &promptrails.DecideParams{Reason: decideReason})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(exec)
		}

		output.Success(fmt.Sprintf("Denied — status: %s", exec.Status))
		return nil
	},
}

// printExecutionTable renders a list of executions as a table.
func printExecutionTable(execs []promptrails.Execution) {
	var rows [][]string
	for _, e := range execs {
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
}

// printExecutionDetails renders a single execution's key fields.
func printExecutionDetails(exec *promptrails.Execution) {
	fmt.Println()
	output.KeyValue("ID", exec.ID)
	output.KeyValue("Agent", exec.AgentName)
	output.KeyValue("Status", exec.Status)
	if exec.ParentExecutionID != nil && *exec.ParentExecutionID != "" {
		output.KeyValue("Parent", *exec.ParentExecutionID)
	}
	output.KeyValue("Duration", fmt.Sprintf("%dms", exec.DurationMs))
	output.KeyValue("Tokens", fmt.Sprintf("%d", exec.TotalTokens))
	output.KeyValue("Created", exec.CreatedAt.Format("2006-01-02 15:04:05"))
	if exec.Status == "waiting_approval" && exec.ApprovalExpiresAt != nil {
		output.KeyValue("Approval Expires", exec.ApprovalExpiresAt.Format("2006-01-02 15:04:05"))
		output.Info("Resolve with 'promptrails execution approve|deny " + exec.ID + "'")
	}
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
}

// printExecutionTree renders an execution and its children recursively.
func printExecutionTree(exec *promptrails.Execution, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%s- %s  %s  [%s]  %dms  %d tok\n",
		indent, exec.ID, exec.AgentName, exec.Status, exec.DurationMs, exec.TotalTokens)
	for i := range exec.Children {
		printExecutionTree(&exec.Children[i], depth+1)
	}
}
