package cmd

import (
	"fmt"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a miko-shell.yaml file in the current directory",
	Long: `Creates a miko-shell.yaml configuration file with default values in the current directory.

By default, creates a configuration using a pre-built Alpine image with setup commands.
Use --dockerfile flag to create a configuration with custom Dockerfile support.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := mikoshell.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		useDockerfile, _ := cmd.Flags().GetBool("dockerfile")
		if err := client.InitProject(useDockerfile); err != nil {
			return err
		}

		fmt.Println("Created miko-shell.yaml successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("dockerfile", "d", false, "Generate configuration with custom Dockerfile instead of pre-built image")
}
