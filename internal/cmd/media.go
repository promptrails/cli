package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/promptrails/cli/internal/output"
	promptrails "github.com/promptrails/go-sdk"
	"github.com/spf13/cobra"
)

var (
	mediaProvider string
	mediaType     string
	mediaModel    string
	mediaPrompt   string
	mediaInputURL string
	mediaConfigKV []string
)

func init() {
	mediaGenerateCmd.Flags().StringVar(&mediaProvider, "provider", "", "Media provider (e.g. elevenlabs, deepgram, fal, replicate, stability, runway, pika, luma)")
	mediaGenerateCmd.Flags().StringVar(&mediaType, "media-type", "", "Media type (tts, stt, image_gen, image_edit, video_gen, video_from_img)")
	mediaGenerateCmd.Flags().StringVar(&mediaModel, "model", "", "Model ID")
	mediaGenerateCmd.Flags().StringVar(&mediaPrompt, "prompt", "", "Prompt text")
	mediaGenerateCmd.Flags().StringVar(&mediaInputURL, "input-url", "", "Input URL (for STT, image editing, video from image)")
	mediaGenerateCmd.Flags().StringSliceVar(&mediaConfigKV, "config", nil, "Config key=value pairs (repeatable)")
	_ = mediaGenerateCmd.MarkFlagRequired("provider")
	_ = mediaGenerateCmd.MarkFlagRequired("media-type")
	_ = mediaGenerateCmd.MarkFlagRequired("model")

	mediaCmd.AddCommand(mediaGenerateCmd)
	rootCmd.AddCommand(mediaCmd)
}

var mediaCmd = &cobra.Command{
	Use:   "media",
	Short: "Generate media content",
}

var mediaGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate media using a provider and model",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		configMap := parseConfigKV(mediaConfigKV)

		resp, err := client.Media.Generate(cmdContext(), &promptrails.GenerateMediaParams{
			Provider:  mediaProvider,
			MediaType: mediaType,
			Model:     mediaModel,
			Prompt:    mediaPrompt,
			InputURL:  mediaInputURL,
			Config:    configMap,
		})
		if err != nil {
			return err
		}

		if getOutputFormat() == output.FormatJSON {
			return output.JSON(resp)
		}

		fmt.Println()
		output.KeyValue("Status", resp.Status)
		if resp.JobID != "" {
			output.KeyValue("Job ID", resp.JobID)
		}
		if resp.AssetURL != "" {
			output.KeyValue("Asset URL", resp.AssetURL)
		}
		if resp.TextOutput != "" {
			output.KeyValue("Text Output", resp.TextOutput)
		}
		if resp.ContentType != "" {
			output.KeyValue("Content Type", resp.ContentType)
		}
		if resp.Metadata != nil {
			metaJSON, _ := json.MarshalIndent(resp.Metadata, "  ", "  ")
			output.KeyValue("Metadata", string(metaJSON))
		}
		fmt.Println()
		return nil
	},
}

// parseConfigKV parses a slice of "key=value" strings into a map.
func parseConfigKV(pairs []string) map[string]any {
	if len(pairs) == 0 {
		return nil
	}
	result := make(map[string]any, len(pairs))
	for _, pair := range pairs {
		k, v, ok := strings.Cut(pair, "=")
		if ok {
			result[k] = v
		}
	}
	return result
}
