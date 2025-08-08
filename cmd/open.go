package cmd

import (
	"fmt"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open an interactive development environment",
	Long:  `Opens an interactive shell session inside the container environment.`,
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

		return client.OpenShell()
	},
}

func init() {
	openCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
	rootCmd.AddCommand(openCmd)
}
