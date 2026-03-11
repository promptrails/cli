package cmd

import (
	"fmt"

	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	mediaModelsProvider  string
	mediaModelsMediaType string
)

func init() {
	mediaModelsListCmd.Flags().StringVar(&mediaModelsProvider, "provider", "", "Filter by provider")
	mediaModelsListCmd.Flags().StringVar(&mediaModelsMediaType, "media-type", "", "Filter by media type (tts, stt, image_gen, image_edit, video_gen, video_from_img)")

	mediaModelsCmd.AddCommand(mediaModelsListCmd)
	rootCmd.AddCommand(mediaModelsCmd)
}

var mediaModelsCmd = &cobra.Command{
	Use:     "media-models",
	Aliases: []string{"mm"},
	Short:   "Browse available media models",
}

var mediaModelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List media models",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		resp, err := client.MediaModels.List(cmdContext(), &promptrails.ListMediaModelsParams{
			Page:      1,
			Limit:     50,
			Provider:  mediaModelsProvider,
			MediaType: mediaModelsMediaType,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp.Data)
		}

		var rows [][]string
		for _, m := range resp.Data {
			active := ""
			if m.IsActive {
				active = "active"
			}
			rows = append(rows, []string{m.ID, m.Name, m.Provider, m.MediaType, active})
		}
		output.Table([]string{"ID", "NAME", "PROVIDER", "MEDIA TYPE", "STATUS"}, rows)

		if resp.Meta.TotalPages > 1 {
			output.Info(fmt.Sprintf("Page %d of %d (%d total)", resp.Meta.Page, resp.Meta.TotalPages, resp.Meta.Total))
		}
		return nil
	},
}
