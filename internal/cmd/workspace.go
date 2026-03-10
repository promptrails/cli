package cmd

import (
	"github.com/promptrails/cli/internal/config"
	"github.com/promptrails/cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	workspaceCmd.AddCommand(workspaceCurrentCmd)
	rootCmd.AddCommand(workspaceCmd)
}

var workspaceCmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"ws"},
	Short:   "Manage workspace context",
}

var workspaceCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.WorkspaceID == "" {
			output.Info("Workspace is determined by the API key.")
			return nil
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(map[string]string{
				"id":   cfg.WorkspaceID,
				"name": cfg.WorkspaceName,
			})
		}

		output.KeyValue("ID", cfg.WorkspaceID)
		output.KeyValue("Name", cfg.WorkspaceName)
		return nil
	},
}
