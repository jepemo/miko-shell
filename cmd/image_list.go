package cmd

import (
	"fmt"
	"strings"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

// imageListCmd represents the image list command
var imageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List container images",
	Long: `List all container images related to miko-shell environments.

This command shows existing images that have been built for miko-shell projects,
along with their basic information like image ID, size, and creation date.`,
	Aliases: []string{"ls"},
	Example: `  # List all miko-shell images
  miko-shell image list
  
  # Using alias
  miko-shell image ls`,
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

		images, err := client.ListImages()
		if err != nil {
			return fmt.Errorf("failed to list images: %w", err)
		}

		if len(images) == 0 {
			fmt.Println("No miko-shell images found")
			return nil
		}

		// Print header
		fmt.Printf("%-20s %-15s %-10s %-20s\n", "IMAGE ID", "TAG", "SIZE", "CREATED")
		fmt.Println(strings.Repeat("-", 67))

		// Print images
		for _, image := range images {
			fmt.Printf("%-20s %-15s %-10s %-20s\n",
				image.ID[:12],
				image.Tag,
				image.Size,
				image.Created.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageListCmd)
	imageListCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
}
