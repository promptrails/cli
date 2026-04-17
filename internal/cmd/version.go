package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current CLI version. Bump this manually on each release.
// Can be overridden at build time via ldflags:
//
//	go build -ldflags "-X github.com/promptrails/cli/internal/cmd.Version=1.2.3"
var Version = "0.3.0"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("promptrails %s\n", Version)
	},
}
