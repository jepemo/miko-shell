package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at build time with -ldflags "-X github.com/jepemo/miko-shell/cmd.version=<value>"
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "miko-shell",
	Short: "A CLI tool for containerized development environments",
	Long: `miko-shell is a CLI tool that serves to abstract dependencies used in a local development project and use containers.
It allows creating a container image based on configuration, and connecting to containers to execute scripts in the project context.`,
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current version of the tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("miko-shell version %s\n", version)
	},
}
