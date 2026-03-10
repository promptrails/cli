package cmd

import (
	"context"
	"os"

	"github.com/promptrails/cli/internal/config"
	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	flagOutput    string
	flagWorkspace string
	flagAPIURL    string
	flagNoColor   bool
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:           "promptrails",
	Short:         "PromptRails CLI — manage agents, prompts, and executions",
	Long:          "Command-line interface for the PromptRails AI agent orchestration platform.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table, json (default from config)")
	rootCmd.PersistentFlags().StringVar(&flagWorkspace, "workspace", "", "Override workspace ID")
	rootCmd.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "Override API URL")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable color output")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
}

// getOutputFormat returns the effective output format.
func getOutputFormat() output.Format {
	if flagOutput != "" {
		return output.Format(flagOutput)
	}
	cfg, err := config.Load()
	if err == nil && cfg.OutputFormat != "" {
		return output.Format(cfg.OutputFormat)
	}
	return output.FormatTable
}

// resolveAPIURL returns the API URL using flag > env > config priority.
func resolveAPIURL(cfg *config.Config) string {
	if flagAPIURL != "" {
		return flagAPIURL
	}
	if envURL := os.Getenv("PROMPTRAILS_API_URL"); envURL != "" {
		return envURL
	}
	return cfg.APIURL
}

// cmdContext returns a background context for API calls.
func cmdContext() context.Context {
	return context.Background()
}

// newSDKClient creates a go-sdk client from stored config.
func newSDKClient() (*promptrails.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, err
	}

	apiKey := os.Getenv("PROMPTRAILS_API_KEY")
	if apiKey == "" {
		apiKey = creds.APIKey
	}

	client := promptrails.NewClient(apiKey,
		promptrails.WithBaseURL(resolveAPIURL(cfg)),
	)
	return client, nil
}

// requireAuth creates an SDK client and fails if not logged in.
func requireAuth() (*promptrails.Client, error) {
	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, err
	}

	envKey := os.Getenv("PROMPTRAILS_API_KEY")
	if !creds.IsLoggedIn() && envKey == "" {
		return nil, &notLoggedInError{}
	}

	return newSDKClient()
}

type notLoggedInError struct{}

func (e *notLoggedInError) Error() string {
	return "not logged in — run 'promptrails init' first"
}
