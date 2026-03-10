package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	wtAgentID        string
	wtName           string
	wtGenerateSecret bool
	wtPage           int
)

func init() {
	wtListCmd.Flags().StringVar(&wtAgentID, "agent-id", "", "Filter by agent ID")
	wtListCmd.Flags().IntVar(&wtPage, "page", 1, "Page number")

	wtCreateCmd.Flags().StringVar(&wtName, "name", "", "Webhook trigger name (required)")
	wtCreateCmd.Flags().StringVar(&wtAgentID, "agent-id", "", "Agent ID (required)")
	wtCreateCmd.Flags().BoolVar(&wtGenerateSecret, "secret", false, "Generate HMAC secret for signature verification")
	_ = wtCreateCmd.MarkFlagRequired("name")
	_ = wtCreateCmd.MarkFlagRequired("agent-id")

	wtUpdateCmd.Flags().StringVar(&wtName, "name", "", "New name")
	wtUpdateCmd.Flags().BoolVar(&wtActive, "active", false, "Set active")
	wtUpdateCmd.Flags().BoolVar(&wtInactive, "inactive", false, "Set inactive")

	webhookTriggerCmd.AddCommand(wtListCmd)
	webhookTriggerCmd.AddCommand(wtGetCmd)
	webhookTriggerCmd.AddCommand(wtCreateCmd)
	webhookTriggerCmd.AddCommand(wtUpdateCmd)
	webhookTriggerCmd.AddCommand(wtDeleteCmd)
	rootCmd.AddCommand(webhookTriggerCmd)
}

var (
	wtActive   bool
	wtInactive bool
)

var webhookTriggerCmd = &cobra.Command{
	Use:     "webhook-trigger",
	Aliases: []string{"wt"},
	Short:   "Manage webhook triggers",
}

var wtListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhook triggers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.WebhookTriggers.List(cmdContext(), &promptrails.ListWebhookTriggersParams{
			Page:    wtPage,
			Limit:   20,
			AgentID: wtAgentID,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, wt := range resp.Data {
			active := "✗"
			if wt.IsActive {
				active = "✓"
			}
			secret := "No"
			if wt.HasSecret {
				secret = "Yes"
			}
			lastUsed := "Never"
			if wt.LastUsedAt != nil {
				lastUsed = wt.LastUsedAt.Format("2006-01-02 15:04")
			}
			rows = append(rows, []string{wt.ID, wt.Name, wt.TokenPrefix + "...", active, secret, lastUsed})
		}
		output.Table([]string{"ID", "NAME", "TOKEN", "ACTIVE", "SECRET", "LAST USED"}, rows)

		if resp.Meta.TotalPages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total))
		}
		return nil
	},
}

var wtGetCmd = &cobra.Command{
	Use:   "get <trigger-id>",
	Short: "Get webhook trigger details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		wt, err := client.WebhookTriggers.Get(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(wt)
		}

		fmt.Println()
		output.KeyValue("ID", wt.ID)
		output.KeyValue("Name", wt.Name)
		output.KeyValue("Agent ID", wt.AgentID)
		output.KeyValue("Token", wt.Token)
		active := "No"
		if wt.IsActive {
			active = "Yes"
		}
		output.KeyValue("Active", active)
		secret := "No"
		if wt.HasSecret {
			secret = "Yes"
		}
		output.KeyValue("Has Secret", secret)
		if wt.LastUsedAt != nil {
			output.KeyValue("Last Used", wt.LastUsedAt.Format("2006-01-02 15:04:05"))
		} else {
			output.KeyValue("Last Used", "Never")
		}
		output.KeyValue("Created", wt.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()

		output.Info("Usage example:")
		fmt.Printf("\n  curl -X POST <BASE_URL>/api/v1/hooks/%s \\\n", wt.Token)
		fmt.Printf("    -H \"Content-Type: application/json\" \\\n")
		fmt.Printf("    -d '{\"input\": {\"message\": \"Hello\"}}'\n\n")
		return nil
	},
}

var wtCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new webhook trigger",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.WebhookTriggers.Create(cmdContext(), &promptrails.CreateWebhookTriggerParams{
			Name:      wtName,
			AgentID:   wtAgentID,
			HasSecret: wtGenerateSecret,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp)
		}

		fmt.Println()
		output.Success(fmt.Sprintf("Webhook trigger created: %s (%s)", resp.Name, resp.ID))
		fmt.Printf("\n  Token: %s\n", resp.Token)
		if resp.Secret != "" {
			fmt.Printf("  Secret: %s\n", resp.Secret)
			output.Warn("Save the secret — it cannot be retrieved later.")
			output.Info("Use the secret to compute HMAC-SHA256 signatures (X-PromptRails-Signature header).")
		}
		fmt.Printf("\n  Endpoint: POST /api/v1/hooks/%s\n\n", resp.Token)

		output.Info("Usage example:")
		fmt.Printf("\n  curl -X POST <BASE_URL>/api/v1/hooks/%s \\\n", resp.Token)
		fmt.Printf("    -H \"Content-Type: application/json\" \\\n")
		fmt.Printf("    -d '{\"input\": {\"message\": \"Hello\"}}'\n\n")
		return nil
	},
}

var wtUpdateCmd = &cobra.Command{
	Use:   "update <trigger-id>",
	Short: "Update a webhook trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		params := &promptrails.UpdateWebhookTriggerParams{}
		changed := false
		if cmd.Flags().Changed("name") {
			params.Name = &wtName
			changed = true
		}
		if cmd.Flags().Changed("active") {
			v := true
			params.IsActive = &v
			changed = true
		}
		if cmd.Flags().Changed("inactive") {
			v := false
			params.IsActive = &v
			changed = true
		}

		if !changed {
			return fmt.Errorf("no fields to update — use --name, --active, or --inactive")
		}

		wt, err := client.WebhookTriggers.Update(cmdContext(), args[0], params)
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(wt)
		}

		output.Success(fmt.Sprintf("Webhook trigger updated: %s", wt.Name))
		return nil
	},
}

var wtDeleteCmd = &cobra.Command{
	Use:   "delete <trigger-id>",
	Short: "Delete a webhook trigger",
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
					Title(fmt.Sprintf("Delete webhook trigger %s?", args[0])).
					Description("External services using this webhook will no longer be able to trigger the agent.").
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

		if err := client.WebhookTriggers.Delete(cmdContext(), args[0]); err != nil {
			return err
		}

		output.Success("Webhook trigger deleted.")
		return nil
	},
}
