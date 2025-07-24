package cmd

import (
	"fmt"
	"strings"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

// isInfrastructureError checks if an error is related to infrastructure
// (docker/podman not available, config issues, etc.) vs script execution errors
func isInfrastructureError(err error) bool {
	errMsg := err.Error()

	// Infrastructure error patterns
	infrastructurePatterns := []string{
		"failed to create client",
		"failed to load config",
		"configuration not loaded",
		"no command specified",
		"failed to build image",
		"docker: not found",
		"podman: not found",
		"container provider",
		"failed to calculate config hash",
	}

	for _, pattern := range infrastructurePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

var runCmd = &cobra.Command{
	Use:   "run [command...]",
	Short: "Run a command inside the container",
	Long:  `Runs a command inside the container. If the command matches a script name, it will run that script.`,
	Args:  cobra.ArbitraryArgs,
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

		// Run command and handle exit codes properly
		err = client.RunCommand(args)
		if err != nil {
			// Check if this is an infrastructure error or a script execution error
			if isInfrastructureError(err) {
				return err // This will show help for infrastructure errors
			}
			// For script execution errors, just exit with the error code
			// without showing help
			cmd.SilenceUsage = true
			return err
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
	rootCmd.AddCommand(runCmd)
}
