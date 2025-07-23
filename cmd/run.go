package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/jepemo/miko-shell/pkg/mikoshell"
)

var runCmd = &cobra.Command{
	Use:   "run [command...]",
	Short: "Run a command inside the container",
	Long:  `Runs a command inside the container. If the command matches a script name, it will run that script.`,
	Args:  cobra.MinimumNArgs(0),
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

		// If no arguments provided, show available scripts
		if len(args) == 0 {
			return client.ListScripts()
		}

		return client.RunCommand(args)
	},
}

func init() {
	runCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
	rootCmd.AddCommand(runCmd)
}
