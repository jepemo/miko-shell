package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/jepemo/miko-shell/pkg/mikoshell"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Access the container shell",
	Long:  `Opens an interactive shell session inside the container.`,
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
	shellCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
	rootCmd.AddCommand(shellCmd)
}
