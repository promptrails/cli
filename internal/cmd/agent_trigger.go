package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	atAgentID        string
	atName           string
	atSource         string
	atGenerateSecret bool
	atPage           int
)

func init() {
	atListCmd.Flags().StringVar(&atAgentID, "agent-id", "", "Filter by agent ID")
	atListCmd.Flags().IntVar(&atPage, "page", 1, "Page number")

	atCreateCmd.Flags().StringVar(&atName, "name", "", "Agent trigger name (required)")
	atCreateCmd.Flags().StringVar(&atAgentID, "agent-id", "", "Agent ID (required)")
	atCreateCmd.Flags().StringVar(&atSource, "source", "generic", "Trigger source: generic | slack | telegram | whatsapp | teams | schedule")
	atCreateCmd.Flags().BoolVar(&atGenerateSecret, "secret", false, "Generate HMAC secret for signature verification")
	_ = atCreateCmd.MarkFlagRequired("name")
	_ = atCreateCmd.MarkFlagRequired("agent-id")

	atUpdateCmd.Flags().StringVar(&atName, "name", "", "New name")
	atUpdateCmd.Flags().BoolVar(&atActive, "active", false, "Set active")
	atUpdateCmd.Flags().BoolVar(&atInactive, "inactive", false, "Set inactive")

	agentTriggerCmd.AddCommand(atListCmd)
	agentTriggerCmd.AddCommand(atGetCmd)
	agentTriggerCmd.AddCommand(atCreateCmd)
	agentTriggerCmd.AddCommand(atUpdateCmd)
	agentTriggerCmd.AddCommand(atDeleteCmd)
	rootCmd.AddCommand(agentTriggerCmd)
}

var (
	atActive   bool
	atInactive bool
)

var agentTriggerCmd = &cobra.Command{
	Use:     "agent-trigger",
	Aliases: []string{"at", "trigger", "webhook-trigger", "wt"},
	Short:   "Manage agent triggers (generic webhook, Slack, Telegram, Teams, WhatsApp, schedule)",
}

var atListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agent triggers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.AgentTriggers.List(cmdContext(), &promptrails.ListAgentTriggersParams{
			Page:    atPage,
			Limit:   20,
			AgentID: atAgentID,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, t := range resp.Data {
			active := "✗"
			if t.IsActive {
				active = "✓"
			}
			secret := "No"
			if t.HasSecret {
				secret = "Yes"
			}
			lastUsed := "Never"
			if t.LastUsedAt != nil {
				lastUsed = t.LastUsedAt.Format("2006-01-02 15:04")
			}
			source := string(t.Source)
			if source == "" {
				source = "generic"
			}
			rows = append(rows, []string{t.ID, t.Name, source, t.TokenPrefix + "...", active, secret, lastUsed})
		}
		output.Table([]string{"ID", "NAME", "SOURCE", "TOKEN", "ACTIVE", "SECRET", "LAST USED"}, rows)

		if resp.Meta.TotalPages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total))
		}
		return nil
	},
}

var atGetCmd = &cobra.Command{
	Use:   "get <trigger-id>",
	Short: "Get agent trigger details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		t, err := client.AgentTriggers.Get(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(t)
		}

		fmt.Println()
		output.KeyValue("ID", t.ID)
		output.KeyValue("Name", t.Name)
		output.KeyValue("Agent ID", t.AgentID)
		source := string(t.Source)
		if source == "" {
			source = "generic"
		}
		output.KeyValue("Source", source)
		output.KeyValue("Token", t.Token)
		active := "No"
		if t.IsActive {
			active = "Yes"
		}
		output.KeyValue("Active", active)
		secret := "No"
		if t.HasSecret {
			secret = "Yes"
		}
		output.KeyValue("Has Secret", secret)
		if t.LastUsedAt != nil {
			output.KeyValue("Last Used", t.LastUsedAt.Format("2006-01-02 15:04:05"))
		} else {
			output.KeyValue("Last Used", "Never")
		}
		output.KeyValue("Created", t.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()

		hookPath := triggerHookPath(t.Source, t.Token)
		output.Info("Usage example:")
		fmt.Printf("\n  curl -X POST <BASE_URL>%s \\\n", hookPath)
		fmt.Printf("    -H \"Content-Type: application/json\" \\\n")
		fmt.Printf("    -d '{\"input\": {\"message\": \"Hello\"}}'\n\n")
		return nil
	},
}

var atCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new agent trigger",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.AgentTriggers.Create(cmdContext(), &promptrails.CreateAgentTriggerParams{
			Name:           atName,
			AgentID:        atAgentID,
			Source:         promptrails.AgentTriggerSource(atSource),
			GenerateSecret: atGenerateSecret,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp)
		}

		fmt.Println()
		output.Success(fmt.Sprintf("Agent trigger created: %s (%s)", resp.Name, resp.ID))
		fmt.Printf("\n  Token: %s\n", resp.Token)
		if resp.Secret != "" {
			fmt.Printf("  Secret: %s\n", resp.Secret)
			output.Warn("Save the secret — it cannot be retrieved later.")
			output.Info("Use the secret to compute HMAC-SHA256 signatures (X-PromptRails-Signature header).")
		}
		hookPath := triggerHookPath(resp.Source, resp.Token)
		fmt.Printf("\n  Endpoint: POST %s\n\n", hookPath)

		output.Info("Usage example:")
		fmt.Printf("\n  curl -X POST <BASE_URL>%s \\\n", hookPath)
		fmt.Printf("    -H \"Content-Type: application/json\" \\\n")
		fmt.Printf("    -d '{\"input\": {\"message\": \"Hello\"}}'\n\n")
		return nil
	},
}

var atUpdateCmd = &cobra.Command{
	Use:   "update <trigger-id>",
	Short: "Update an agent trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		params := &promptrails.UpdateAgentTriggerParams{}
		changed := false
		if cmd.Flags().Changed("name") {
			params.Name = &atName
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

		t, err := client.AgentTriggers.Update(cmdContext(), args[0], params)
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(t)
		}

		output.Success(fmt.Sprintf("Agent trigger updated: %s", t.Name))
		return nil
	},
}

var atDeleteCmd = &cobra.Command{
	Use:   "delete <trigger-id>",
	Short: "Delete an agent trigger",
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
					Title(fmt.Sprintf("Delete agent trigger %s?", args[0])).
					Description("External services using this trigger will no longer be able to invoke the agent.").
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

		if err := client.AgentTriggers.Delete(cmdContext(), args[0]); err != nil {
			return err
		}

		output.Success("Agent trigger deleted.")
		return nil
	},
}

// triggerHookPath returns the public hook URL path for a trigger source.
func triggerHookPath(source promptrails.AgentTriggerSource, token string) string {
	switch source {
	case promptrails.AgentTriggerSourceSlack:
		return "/api/v1/hooks/slack/" + token
	case promptrails.AgentTriggerSourceTelegram:
		return "/api/v1/hooks/telegram/" + token
	case promptrails.AgentTriggerSourceTeams:
		return "/api/v1/hooks/teams/" + token
	case promptrails.AgentTriggerSourceWhatsApp:
		return "/api/v1/hooks/whatsapp/" + token
	default:
		return "/api/v1/hooks/" + token
	}
}
