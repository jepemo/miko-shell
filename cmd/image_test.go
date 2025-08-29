package cmd

import (
	"testing"
)

func TestImageCommand(t *testing.T) {
	// Test that the image command is properly configured
	if imageCmd == nil {
		t.Fatal("imageCmd should not be nil")
	}

	if imageCmd.Use != "image" {
		t.Errorf("Expected Use to be 'image', got '%s'", imageCmd.Use)
	}

	if imageCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if imageCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	// Test that subcommands are properly registered
	subcommands := imageCmd.Commands()
	expectedSubcommands := []string{"build", "list", "clean", "info", "prune"}

	// Verify each expected subcommand exists
	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range subcommands {
			if cmd.Use == expected {
				found = true
				break
			}
		}
		if !found {
			// Print available commands for debugging
			availableCommands := make([]string, len(subcommands))
			for i, cmd := range subcommands {
				availableCommands[i] = cmd.Use
			}
			t.Errorf("Expected subcommand '%s' not found. Available commands: %v", expected, availableCommands)
		}
	}
}

func TestImageBuildCommand(t *testing.T) {
	// Test that the image build command has the right flags
	if imageBuildCmd == nil {
		t.Fatal("imageBuildCmd should not be nil")
	}

	// Check for force flag
	forceFlag := imageBuildCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected --force flag to be present")
	}

	// Check for config flag
	configFlag := imageBuildCmd.Flags().Lookup("config")
	if configFlag == nil {
		t.Error("Expected --config flag to be present")
	}
}

func TestImageListCommand(t *testing.T) {
	// Test that the image list command has aliases
	if imageListCmd == nil {
		t.Fatal("imageListCmd should not be nil")
	}

	aliases := imageListCmd.Aliases
	expectedAliases := []string{"ls"}

	if len(aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(aliases))
	}

	for i, alias := range aliases {
		if alias != expectedAliases[i] {
			t.Errorf("Expected alias '%s', got '%s'", expectedAliases[i], alias)
		}
	}
}

func TestImageCleanCommand(t *testing.T) {
	// Test that the image clean command has the right flags
	if imageCleanCmd == nil {
		t.Fatal("imageCleanCmd should not be nil")
	}

	// Check for all flag
	allFlag := imageCleanCmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Error("Expected --all flag to be present")
	}
}

func TestImagePruneCommand(t *testing.T) {
	// Test that the image prune command has the right flags
	if imagePruneCmd == nil {
		t.Fatal("imagePruneCmd should not be nil")
	}

	// Check for force flag
	forceFlag := imagePruneCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected --force flag to be present")
	}
}
