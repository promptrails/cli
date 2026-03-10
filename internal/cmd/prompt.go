package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	promptName        string
	promptDescription string
	promptStatus      string
	promptInput       string
	promptInputFile   string
	promptUserPrompt  string
	promptPage        int
)

func init() {
	promptListCmd.Flags().StringVar(&promptStatus, "status", "", "Filter by status")
	promptListCmd.Flags().IntVar(&promptPage, "page", 1, "Page number")

	promptCreateCmd.Flags().StringVar(&promptName, "name", "", "Prompt name (required)")
	promptCreateCmd.Flags().StringVar(&promptDescription, "description", "", "Prompt description")
	_ = promptCreateCmd.MarkFlagRequired("name")

	promptUpdateCmd.Flags().StringVar(&promptName, "name", "", "New name")
	promptUpdateCmd.Flags().StringVar(&promptDescription, "description", "", "New description")

	promptRunCmd.Flags().StringVar(&promptInput, "input", "", "Input JSON string")
	promptRunCmd.Flags().StringVar(&promptInputFile, "input-file", "", "Path to input JSON file")
	promptRunCmd.Flags().StringVar(&promptUserPrompt, "user-prompt", "", "User prompt text")

	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptGetCmd)
	promptCmd.AddCommand(promptCreateCmd)
	promptCmd.AddCommand(promptUpdateCmd)
	promptCmd.AddCommand(promptDeleteCmd)
	promptCmd.AddCommand(promptRunCmd)
	promptCmd.AddCommand(promptVersionsCmd)
	promptCmd.AddCommand(promptPromoteCmd)
	rootCmd.AddCommand(promptCmd)
}

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage prompts",
}

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List prompts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Prompts.List(cmdContext(), &promptrails.ListPromptsParams{
			Page:  promptPage,
			Limit: 20,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, p := range resp.Data {
			rows = append(rows, []string{p.ID, p.Name, p.Status})
		}
		output.Table([]string{"ID", "NAME", "STATUS"}, rows)

		if resp.Meta.TotalPages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total))
		}
		return nil
	},
}

var promptGetCmd = &cobra.Command{
	Use:   "get <prompt-id>",
	Short: "Get prompt details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		p, err := client.Prompts.Get(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(p)
		}

		fmt.Println()
		output.KeyValue("ID", p.ID)
		output.KeyValue("Name", p.Name)
		output.KeyValue("Status", p.Status)
		output.KeyValue("Description", p.Description)
		output.KeyValue("Created", p.CreatedAt.Format("2006-01-02 15:04:05"))
		output.KeyValue("Updated", p.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
		return nil
	},
}

var promptCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new prompt",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		p, err := client.Prompts.Create(cmdContext(), &promptrails.CreatePromptParams{
			Name:        promptName,
			Description: promptDescription,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(p)
		}

		output.Success(fmt.Sprintf("Prompt created: %s (%s)", p.Name, p.ID))
		return nil
	},
}

var promptUpdateCmd = &cobra.Command{
	Use:   "update <prompt-id>",
	Short: "Update a prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		params := &promptrails.UpdatePromptParams{}
		changed := false
		if cmd.Flags().Changed("name") {
			params.Name = &promptName
			changed = true
		}
		if cmd.Flags().Changed("description") {
			params.Description = &promptDescription
			changed = true
		}

		if !changed {
			return fmt.Errorf("no fields to update — use --name or --description")
		}

		p, err := client.Prompts.Update(cmdContext(), args[0], params)
		if err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Prompt updated: %s", p.Name))
		return nil
	},
}

var promptDeleteCmd = &cobra.Command{
	Use:   "delete <prompt-id>",
	Short: "Delete a prompt",
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
					Title(fmt.Sprintf("Delete prompt %s?", args[0])).
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

		if err := client.Prompts.Delete(cmdContext(), args[0]); err != nil {
			return err
		}

		output.Success("Prompt deleted.")
		return nil
	},
}

var promptRunCmd = &cobra.Command{
	Use:   "run <prompt-id>",
	Short: "Run a prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		inputData, err := parseInputJSON(promptInput, promptInputFile)
		if err != nil {
			return err
		}

		resp, err := client.Prompts.Run(cmdContext(), args[0], &promptrails.RunPromptParams{
			UserPrompt: promptUserPrompt,
			Input:      inputData,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp)
		}

		fmt.Println(resp.Output)
		return nil
	},
}

var promptVersionsCmd = &cobra.Command{
	Use:   "versions <prompt-id>",
	Short: "List prompt versions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Prompts.ListVersions(cmdContext(), args[0], nil)
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
			rows = append(rows, []string{active, v.ID, v.Version, v.CreatedAt.Format("2006-01-02 15:04")})
		}
		output.Table([]string{"", "ID", "VERSION", "CREATED"}, rows)
		return nil
	},
}

var promptPromoteCmd = &cobra.Command{
	Use:   "promote <prompt-id> <version-id>",
	Short: "Promote a prompt version to active",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		if err := client.Prompts.PromoteVersion(cmdContext(), args[0], args[1]); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Version %s promoted to active.", args[1]))
		return nil
	},
}
