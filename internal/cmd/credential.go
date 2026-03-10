package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	credProvider string
	credName     string
)

func init() {
	credCreateCmd.Flags().StringVar(&credProvider, "provider", "", "Provider (openai, anthropic, google_genai, deepseek, fireworks, xai, openrouter)")
	credCreateCmd.Flags().StringVar(&credName, "name", "", "Credential name")
	_ = credCreateCmd.MarkFlagRequired("provider")
	_ = credCreateCmd.MarkFlagRequired("name")

	credCmd.AddCommand(credListCmd)
	credCmd.AddCommand(credCreateCmd)
	credCmd.AddCommand(credDeleteCmd)
	credCmd.AddCommand(credCheckCmd)
	rootCmd.AddCommand(credCmd)
}

var credCmd = &cobra.Command{
	Use:     "credential",
	Aliases: []string{"cred"},
	Short:   "Manage provider credentials",
}

var credListCmd = &cobra.Command{
	Use:   "list",
	Short: "List credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Credentials.List(cmdContext(), &promptrails.ListCredentialsParams{
			Page:  1,
			Limit: 50,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, c := range resp.Data {
			def := ""
			if c.IsDefault {
				def = "●"
			}
			valid := "✓"
			if !c.IsValid {
				valid = "✗"
			}
			rows = append(rows, []string{c.ID, c.Name, c.Provider, def, valid})
		}
		output.Table([]string{"ID", "NAME", "PROVIDER", "DEFAULT", "VALID"}, rows)
		return nil
	},
}

var credCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new credential",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		var apiKey string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("API Key").
					Description(fmt.Sprintf("Enter the API key for %s", credProvider)).
					EchoMode(huh.EchoModePassword).
					Value(&apiKey),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}

		cred, err := client.Credentials.Create(cmdContext(), &promptrails.CreateCredentialParams{
			Name:     credName,
			Provider: credProvider,
			APIKey:   apiKey,
		})
		if err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Credential created: %s (%s)", cred.Name, cred.ID))
		return nil
	},
}

var credDeleteCmd = &cobra.Command{
	Use:   "delete <credential-id>",
	Short: "Delete a credential",
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
					Title(fmt.Sprintf("Delete credential %s?", args[0])).
					Description("This may break agents using this credential.").
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

		if err := client.Credentials.Delete(cmdContext(), args[0]); err != nil {
			return err
		}

		output.Success("Credential deleted.")
		return nil
	},
}

var credCheckCmd = &cobra.Command{
	Use:   "check <credential-id>",
	Short: "Test a credential's validity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		_, err = client.Credentials.CheckConnection(cmdContext(), args[0])
		if err != nil {
			output.Error(fmt.Sprintf("Credential check failed: %s", err))
			return nil
		}

		output.Success("Credential is valid.")
		return nil
	},
}
