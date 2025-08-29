package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jepemo/miko-shell/pkg/mikoshell"
	"github.com/spf13/cobra"
)

var imagePruneForce bool

// imagePruneCmd represents the image prune command
var imagePruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove all unused container images and build cache",
	Long: `Remove all unused container images and build cache to reclaim disk space.

This command removes:
- All dangling images (not associated with any container)
- All unused images (not referenced by any container) 
- Build cache and intermediate layers

Use --force to skip the confirmation prompt.`,
	Example: `  # Prune unused images with confirmation
  miko-shell image prune

  # Prune without confirmation prompt
  miko-shell image prune --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		if configFile == "" {
			configFile = "miko-shell.yaml"
		}

		config, err := mikoshell.LoadConfigFromFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client, err := mikoshell.NewClientWithConfigFile(config, configFile)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Show what will be removed
		pruneInfo, err := client.GetPruneInfo()
		if err != nil {
			return fmt.Errorf("failed to get prune info: %w", err)
		}

		if pruneInfo.TotalImages == 0 {
			fmt.Println("No images to prune")
			return nil
		}

		fmt.Printf("This will remove:\n")
		fmt.Printf("  - %d unused image(s)\n", pruneInfo.UnusedImages)
		fmt.Printf("  - %d dangling image(s)\n", pruneInfo.DanglingImages)
		fmt.Printf("  - Build cache (~%s)\n", pruneInfo.BuildCacheSize)
		fmt.Printf("Total space to reclaim: ~%s\n\n", pruneInfo.TotalSize)

		// Confirm unless --force is used
		if !imagePruneForce {
			fmt.Print("Are you sure you want to continue? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Operation cancelled")
				return nil
			}
		}

		fmt.Println("Pruning images and build cache...")

		result, err := client.PruneImages()
		if err != nil {
			return fmt.Errorf("failed to prune images: %w", err)
		}

		fmt.Printf("Pruning completed successfully!\n")
		fmt.Printf("Removed %d image(s)\n", result.RemovedImages)
		fmt.Printf("Reclaimed space: %s\n", result.ReclaimedSpace)

		return nil
	},
}

func init() {
	imageCmd.AddCommand(imagePruneCmd)
	imagePruneCmd.Flags().BoolVarP(&imagePruneForce, "force", "f", false, "Do not prompt for confirmation")
	imagePruneCmd.Flags().StringP("config", "c", "", "Path to configuration file (default: miko-shell.yaml)")
}
