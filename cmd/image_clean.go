package cmd

import (
	"fmt"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

var imageCleanAll bool

// imageCleanCmd represents the image clean command
var imageCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove container images",
	Long: `Remove container images related to miko-shell environments.

By default, this command removes unused images. Use --all to remove all miko-shell images,
including the ones that might be in use.`,
	Example: `  # Remove unused miko-shell images
  miko-shell image clean

  # Remove all miko-shell images (including active ones)
  miko-shell image clean --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		if configFile == "" {
			configFile = "miko-shell.yaml"
		}

		config, err := mikoshell.LoadConfigFromFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client, err := mikoshell.NewClientWithConfig(config)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		fmt.Println("Cleaning container images...")

		removed, err := client.CleanImages(imageCleanAll)
		if err != nil {
			return fmt.Errorf("failed to clean images: %w", err)
		}

		if len(removed) == 0 {
			fmt.Println("No images were removed")
			return nil
		}

		fmt.Printf("Removed %d image(s):\n", len(removed))
		for _, imageID := range removed {
			fmt.Printf("  - %s\n", imageID)
		}

		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageCleanCmd)
	imageCleanCmd.Flags().BoolVarP(&imageCleanAll, "all", "a", false, "Remove all miko-shell images, including active ones")
	imageCleanCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
}
