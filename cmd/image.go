package cmd

import (
	"github.com/spf13/cobra"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage container images",
	Long: `Manage container images for miko-shell environments.

This command provides subcommands to build, list, clean, and manage
container images used by miko-shell for development environments.`,
	Example: `  # Build container image
  miko-shell image build

  # List miko-shell images
  miko-shell image list

  # Clean project images
  miko-shell image clean`,
}

func init() {
	rootCmd.AddCommand(imageCmd)
}
