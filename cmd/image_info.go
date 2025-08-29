package cmd

import (
	"fmt"
	"strings"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

// imageInfoCmd represents the image info command
var imageInfoCmd = &cobra.Command{
	Use:   "info",
	Args:  cobra.MaximumNArgs(1),
	Short: "Show detailed information about a container image",
	Long: `Show detailed information about a specific container image.

If no image ID is provided, shows information about the current project's image
based on the miko-shell.yaml configuration.

Usage: miko-shell image info [IMAGE_ID]`,
	Example: `  # Show info for current project's image
  miko-shell image info

  # Show info for specific image
  miko-shell image info abc123def456`,
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

		var imageID string
		if len(args) > 0 {
			imageID = args[0]
		}

		imageInfo, err := client.GetImageInfo(imageID)
		if err != nil {
			return fmt.Errorf("failed to get image info: %w", err)
		}

		// Print image information
		fmt.Printf("Image Information:\n")
		fmt.Printf("=================\n\n")

		fmt.Printf("ID:          %s\n", imageInfo.ID)
		fmt.Printf("Tag:         %s\n", imageInfo.Tag)
		fmt.Printf("Size:        %s\n", imageInfo.Size)
		fmt.Printf("Created:     %s\n", imageInfo.Created.Format("2006-01-02 15:04:05"))
		fmt.Printf("Platform:    %s\n", imageInfo.Platform)

		if len(imageInfo.Labels) > 0 {
			fmt.Printf("\nLabels:\n")
			for key, value := range imageInfo.Labels {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}

		if len(imageInfo.Layers) > 0 {
			fmt.Printf("\nLayers (%d):\n", len(imageInfo.Layers))
			for i, layer := range imageInfo.Layers {
				fmt.Printf("  %d. %s (%s)\n", i+1, layer.ID[:12], layer.Size)
			}
		}

		if len(imageInfo.Env) > 0 {
			fmt.Printf("\nEnvironment Variables:\n")
			for _, env := range imageInfo.Env {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					fmt.Printf("  %s=%s\n", parts[0], parts[1])
				}
			}
		}

		if len(imageInfo.ExposedPorts) > 0 {
			fmt.Printf("\nExposed Ports:\n")
			for _, port := range imageInfo.ExposedPorts {
				fmt.Printf("  %s\n", port)
			}
		}

		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageInfoCmd)
	imageInfoCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
}
