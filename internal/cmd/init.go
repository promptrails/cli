package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/promptrails/cli/internal/config"
	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var initAPIKey string

func init() {
	initCmd.Flags().StringVar(&initAPIKey, "api-key", "", "API key for non-interactive setup")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up PromptRails CLI with an API key",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	banner := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("  PromptRails CLI Setup")
	fmt.Println()
	fmt.Println(banner)
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if flagAPIURL != "" {
		cfg.APIURL = flagAPIURL
	}

	apiKey := initAPIKey

	if apiKey == "" {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("API Key").
					Description("Enter your PromptRails API key (starts with pr_key_)").
					Value(&apiKey),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
	}

	// Validate the API key
	client := promptrails.NewClient(apiKey,
		promptrails.WithBaseURL(resolveAPIURL(cfg)),
	)
	_, err = client.Agents.List(cmdContext(), &promptrails.ListAgentsParams{Page: 1, Limit: 1})
	if err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}

	creds := &config.Credentials{
		APIKey: apiKey,
	}
	if err := creds.Save(); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Println()
	output.Success("API key validated successfully.")
	output.Info("Workspace is determined by the API key. Use 'promptrails agent list' to verify access.")
	fmt.Println()

	return nil
}
