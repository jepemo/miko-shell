package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/jepemo/miko-shell/pkg/mikoshell"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a dev-config.yaml file in the current directory",
	Long:  `Creates a dev-config.yaml configuration file with default values in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := mikoshell.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		if err := client.InitProject(); err != nil {
			return err
		}

		fmt.Println("Created dev-config.yaml successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
