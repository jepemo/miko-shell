package cmd

import (
	"fmt"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a container image from the configuration",
	Long:  `Builds a container image based on the miko-shell.yaml configuration file.`,
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

		force, _ := cmd.Flags().GetBool("force")

		fmt.Println("Building container image...")
		if err := client.BuildImage(force); err != nil {
			return err
		}

		fmt.Println("Container image built successfully!")
		return nil
	},
}

func init() {
	buildCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuild by removing existing image first")
	rootCmd.AddCommand(buildCmd)
}
