package cmd

import (
	"fmt"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

var imageBuildForce bool

// imageBuildCmd represents the image build command
var imageBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build container image",
	Long: `Build the container image for the miko-shell environment.

If the image already exists, it will not be rebuilt unless the --force flag is used.
The image is built based on the configuration in miko-shell.yaml.`,
	Example: `  # Build container image
  miko-shell image build

  # Force rebuild of existing image
  miko-shell image build --force`,
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

		fmt.Println("Building container image...")
		if err := client.BuildImage(imageBuildForce); err != nil {
			return fmt.Errorf("failed to build image: %w", err)
		}

		fmt.Println("Container image built successfully!")
		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageBuildCmd)
	imageBuildCmd.Flags().BoolVarP(&imageBuildForce, "force", "f", false, "Force rebuild by removing existing image first")
	imageBuildCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
}
