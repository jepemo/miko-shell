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
		client, err := mikoshell.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		configFile, _ := cmd.Flags().GetString("config")
		if configFile != "" {
			if err := client.LoadConfigFromFile(configFile); err != nil {
				return err
			}
		} else {
			if err := client.LoadConfig(); err != nil {
				return err
			}
		}

		force, _ := cmd.Flags().GetBool("force")
		
		fmt.Println("Building container image...")
		tag, err := client.BuildImageWithForce(force)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully built image: %s\n", tag)
		return nil
	},
}

func init() {
	buildCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuild by removing existing image first")
	rootCmd.AddCommand(buildCmd)
}
