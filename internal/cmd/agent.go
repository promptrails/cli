package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	agentType        string
	agentName        string
	agentDescription string
	agentInput       string
	agentInputFile   string
	agentPage        int
)

func init() {
	agentListCmd.Flags().StringVar(&agentType, "type", "", "Filter by type (simple, chain, multi_agent, workflow, composite)")
	agentListCmd.Flags().StringVar(&agentName, "name", "", "Filter by name")
	agentListCmd.Flags().IntVar(&agentPage, "page", 1, "Page number")

	agentCreateCmd.Flags().StringVar(&agentName, "name", "", "Agent name (required)")
	agentCreateCmd.Flags().StringVar(&agentType, "type", "simple", "Agent type")
	agentCreateCmd.Flags().StringVar(&agentDescription, "description", "", "Agent description")
	_ = agentCreateCmd.MarkFlagRequired("name")

	agentUpdateCmd.Flags().StringVar(&agentName, "name", "", "New name")
	agentUpdateCmd.Flags().StringVar(&agentDescription, "description", "", "New description")

	agentExecuteCmd.Flags().StringVar(&agentInput, "input", "", "Input JSON string")
	agentExecuteCmd.Flags().StringVar(&agentInputFile, "input-file", "", "Path to input JSON file")

	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentGetCmd)
	agentCmd.AddCommand(agentCreateCmd)
	agentCmd.AddCommand(agentUpdateCmd)
	agentCmd.AddCommand(agentDeleteCmd)
	agentCmd.AddCommand(agentExecuteCmd)
	agentCmd.AddCommand(agentVersionsCmd)
	agentCmd.AddCommand(agentPromoteCmd)
	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents",
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Agents.List(cmdContext(), &promptrails.ListAgentsParams{
			Page:   agentPage,
			Limit:  20,
			Type:   agentType,
			Search: agentName,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, a := range resp.Data {
			rows = append(rows, []string{a.ID, a.Name, a.Type, a.Status})
		}
		output.Table([]string{"ID", "NAME", "TYPE", "STATUS"}, rows)

		if resp.Meta.TotalPages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total))
		}
		return nil
	},
}

var agentGetCmd = &cobra.Command{
	Use:   "get <agent-id>",
	Short: "Get agent details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		agent, err := client.Agents.Get(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(agent)
		}

		fmt.Println()
		output.KeyValue("ID", agent.ID)
		output.KeyValue("Name", agent.Name)
		output.KeyValue("Type", agent.Type)
		output.KeyValue("Status", agent.Status)
		output.KeyValue("Description", agent.Description)
		output.KeyValue("Created", agent.CreatedAt.Format("2006-01-02 15:04:05"))
		output.KeyValue("Updated", agent.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
		return nil
	},
}

var agentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		agent, err := client.Agents.Create(cmdContext(), &promptrails.CreateAgentParams{
			Name:        agentName,
			Type:        agentType,
			Description: agentDescription,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(agent)
		}

		output.Success(fmt.Sprintf("Agent created: %s (%s)", agent.Name, agent.ID))
		return nil
	},
}

var agentUpdateCmd = &cobra.Command{
	Use:   "update <agent-id>",
	Short: "Update an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		params := &promptrails.UpdateAgentParams{}
		changed := false
		if cmd.Flags().Changed("name") {
			params.Name = &agentName
			changed = true
		}
		if cmd.Flags().Changed("description") {
			params.Description = &agentDescription
			changed = true
		}

		if !changed {
			return fmt.Errorf("no fields to update — use --name or --description")
		}

		agent, err := client.Agents.Update(cmdContext(), args[0], params)
		if err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Agent updated: %s", agent.Name))
		return nil
	},
}

var agentDeleteCmd = &cobra.Command{
	Use:   "delete <agent-id>",
	Short: "Delete an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Delete agent %s?", args[0])).
					Description("This action cannot be undone.").
					Value(&confirm),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		if !confirm {
			output.Info("Cancelled.")
			return nil
		}

		if err := client.Agents.Delete(cmdContext(), args[0]); err != nil {
			return err
		}

		output.Success("Agent deleted.")
		return nil
	},
}

var agentExecuteCmd = &cobra.Command{
	Use:   "execute <agent-id>",
	Short: "Execute an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		inputData, err := parseInputJSON(agentInput, agentInputFile)
		if err != nil {
			return err
		}

		resp, err := client.Agents.Execute(cmdContext(), args[0], &promptrails.ExecuteAgentParams{
			Input: inputData,
			Sync:  true,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp)
		}

		fmt.Println()
		output.KeyValue("Execution ID", resp.ExecutionID)
		output.KeyValue("Status", resp.Status)
		if resp.Output != nil {
			outputJSON, _ := json.MarshalIndent(resp.Output, "  ", "  ")
			output.KeyValue("Output", string(outputJSON))
		}
		fmt.Println()
		return nil
	},
}

var agentVersionsCmd = &cobra.Command{
	Use:   "versions <agent-id>",
	Short: "List agent versions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Agents.ListVersions(cmdContext(), args[0], nil)
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, v := range resp.Data {
			active := ""
			if v.IsActive {
				active = "●"
			}
			rows = append(rows, []string{active, v.ID, v.Version, v.Message, v.CreatedAt.Format("2006-01-02 15:04")})
		}
		output.Table([]string{"", "ID", "VERSION", "MESSAGE", "CREATED"}, rows)
		return nil
	},
}

var agentPromoteCmd = &cobra.Command{
	Use:   "promote <agent-id> <version-id>",
	Short: "Promote an agent version to active",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		if err := client.Agents.PromoteVersion(cmdContext(), args[0], args[1]); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Version %s promoted to active.", args[1]))
		return nil
	},
}

// parseInputJSON reads input from --input flag or --input-file flag.
func parseInputJSON(inputStr, inputFile string) (map[string]any, error) {
	var raw []byte

	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("read input file: %w", err)
		}
		raw = data
	} else if inputStr != "" {
		raw = []byte(inputStr)
	} else {
		return map[string]any{}, nil
	}

	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}
	return result, nil
}
