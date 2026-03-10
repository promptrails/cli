package cmd

import (
	"fmt"

	"github.com/promptrails/cli/internal/config"
	"github.com/promptrails/cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication and workspace context",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		creds, err := config.LoadCredentials()
		if err != nil {
			return err
		}

		fmt.Println()
		output.KeyValue("API URL", resolveAPIURL(cfg))
		output.KeyValue("Output Format", cfg.OutputFormat)

		if creds.IsLoggedIn() {
			masked := creds.APIKey[:10] + "..." + creds.APIKey[len(creds.APIKey)-4:]
			output.KeyValue("API Key", masked)
		} else {
			output.KeyValue("Auth", "Not configured — run 'promptrails init'")
		}

		if cfg.WorkspaceID != "" {
			wsDisplay := cfg.WorkspaceID
			if cfg.WorkspaceName != "" {
				wsDisplay = fmt.Sprintf("%s (%s)", cfg.WorkspaceName, cfg.WorkspaceID)
			}
			output.KeyValue("Workspace", wsDisplay)
		} else {
			output.KeyValue("Workspace", "Determined by API key")
		}

		fmt.Println()
		return nil
	},
}
