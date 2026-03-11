package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	assetsType     string
	assetsProvider string
	assetsPage     int
	assetsLimit    int
)

func init() {
	assetsListCmd.Flags().StringVar(&assetsType, "type", "", "Filter by media type (audio, image, video)")
	assetsListCmd.Flags().StringVar(&assetsProvider, "provider", "", "Filter by provider")
	assetsListCmd.Flags().IntVar(&assetsPage, "page", 1, "Page number")
	assetsListCmd.Flags().IntVar(&assetsLimit, "limit", 20, "Items per page")

	assetsCmd.AddCommand(assetsListCmd)
	assetsCmd.AddCommand(assetsGetCmd)
	assetsCmd.AddCommand(assetsDeleteCmd)
	assetsCmd.AddCommand(assetsSignedURLCmd)
	rootCmd.AddCommand(assetsCmd)
}

var assetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Manage media assets",
}

var assetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List assets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Assets.List(cmdContext(), &promptrails.ListAssetsParams{
			Page:      assetsPage,
			Limit:     assetsLimit,
			MediaType: assetsType,
			Provider:  assetsProvider,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, a := range resp.Data {
			size := formatSize(a.Size)
			rows = append(rows, []string{a.ID, a.FileName, a.ContentType, a.Provider, size, a.CreatedAt.Format("2006-01-02 15:04")})
		}
		output.Table([]string{"ID", "FILE NAME", "CONTENT TYPE", "PROVIDER", "SIZE", "CREATED"}, rows)

		if resp.Meta.TotalPages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total))
		}
		return nil
	},
}

var assetsGetCmd = &cobra.Command{
	Use:   "get <asset-id>",
	Short: "Get asset details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		asset, err := client.Assets.Get(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(asset)
		}

		fmt.Println()
		output.KeyValue("ID", asset.ID)
		output.KeyValue("File Name", asset.FileName)
		output.KeyValue("Content Type", asset.ContentType)
		output.KeyValue("Media Type", asset.MediaType)
		output.KeyValue("Provider", asset.Provider)
		output.KeyValue("Size", formatSize(asset.Size))
		if asset.URL != "" {
			output.KeyValue("URL", asset.URL)
		}
		output.KeyValue("Created", asset.CreatedAt.Format("2006-01-02 15:04:05"))
		output.KeyValue("Updated", asset.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
		return nil
	},
}

var assetsDeleteCmd = &cobra.Command{
	Use:   "delete <asset-id>",
	Short: "Delete an asset",
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
					Title(fmt.Sprintf("Delete asset %s?", args[0])).
					Description("This will remove the asset from storage. This action cannot be undone.").
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

		if err := client.Assets.Delete(cmdContext(), args[0]); err != nil {
			return err
		}

		output.Success("Asset deleted.")
		return nil
	},
}

var assetsSignedURLCmd = &cobra.Command{
	Use:   "signed-url <asset-id>",
	Short: "Get a temporary signed URL for an asset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.Assets.GetSignedURL(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp)
		}

		fmt.Println()
		output.KeyValue("URL", resp.URL)
		if resp.ExpiresAt != "" {
			output.KeyValue("Expires At", resp.ExpiresAt)
		}
		fmt.Println()
		return nil
	},
}

// formatSize formats bytes into a human-readable string.
func formatSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
