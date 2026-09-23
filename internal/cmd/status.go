package cmd

import (
	"fmt"
	"strings"

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
			output.KeyValue("API Key", maskSecret(creds.APIKey))
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

// maskSecret renders a credential as its last four characters: enough to tell
// which key is configured, little enough that a screenshot or a shoulder does
// not give it away. It also handles a key shorter than that, which the
// previous fixed-width slice would have panicked on.
func maskSecret(secret string) string {
	const shown = 4
	if len(secret) <= shown {
		return strings.Repeat("*", len(secret))
	}
	return "..." + secret[len(secret)-shown:]
}
