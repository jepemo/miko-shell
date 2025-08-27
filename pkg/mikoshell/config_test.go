package mikoshell

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "uppercase to lowercase",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "spaces to dashes",
			input:    "hello world",
			expected: "hello-world",
		},
		{
			name:     "special characters",
			input:    "hello@world!",
			expected: "hello-world",
		},
		{
			name:     "accented characters",
			input:    "café",
			expected: "cafe",
		},
		{
			name:     "multiple dashes",
			input:    "hello---world",
			expected: "hello-world",
		},
		{
			name:     "leading and trailing dashes",
			input:    "-hello-world-",
			expected: "hello-world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "project",
		},
		{
			name:     "only special characters",
			input:    "@#$%",
			expected: "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetCurrentDirName(t *testing.T) {
	// Save original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	}()

	// Create temporary directory with specific name
	tempDir, err := os.MkdirTemp("", "test-project-name")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	result := GetCurrentDirName()
	expected := NormalizeName(filepath.Base(tempDir))

	if result != expected {
		t.Errorf("GetCurrentDirName() = %q, want %q", result, expected)
	}
}

func TestConfigExists(t *testing.T) {
	// Test in directory without config
	if ConfigExists() {
		t.Error("ConfigExists() should return false when no config file exists")
	}

	// Create temporary config file
	tempFile, err := os.CreateTemp("", ConfigFileName)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Save original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	}()

	// Change to directory with config file
	tempDir := filepath.Dir(tempFile.Name())
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Rename temp file to config name
	configPath := filepath.Join(tempDir, ConfigFileName)
	if err := os.Rename(tempFile.Name(), configPath); err != nil {
		t.Fatalf("Failed to rename temp file: %v", err)
	}
	defer os.Remove(configPath)

	if !ConfigExists() {
		t.Error("ConfigExists() should return true when config file exists")
	}
}

func TestLoadConfig(t *testing.T) {
	// Save original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	}()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-config")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	t.Run("no config file", func(t *testing.T) {
		_, err := LoadConfig()
		if err == nil {
			t.Error("LoadConfig() should return error when no config file exists")
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		configContent := `name: test-project
container:
  provider: docker
  image: alpine:latest
  setup:
    - apk add curl
shell:
  startup:
    - echo "Hello"
  scripts:
    - name: test
      commands:
        - echo "test command"
`
		if err := os.WriteFile(ConfigFileName, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		if config.Name != "test-project" {
			t.Errorf("Expected name 'test-project', got '%s'", config.Name)
		}
		if config.Container.Provider != "docker" {
			t.Errorf("Expected provider 'docker', got '%s'", config.Container.Provider)
		}
		if config.Container.Image != "alpine:latest" {
			t.Errorf("Expected image 'alpine:latest', got '%s'", config.Container.Image)
		}
		if len(config.Container.Setup) != 1 || config.Container.Setup[0] != "apk add curl" {
			t.Errorf("Expected setup ['apk add curl'], got %v", config.Container.Setup)
		}
		if len(config.Shell.Scripts) != 1 || config.Shell.Scripts[0].Name != "test" {
			t.Errorf("Expected script named 'test', got %v", config.Shell.Scripts)
		}
	})

	t.Run("invalid container provider", func(t *testing.T) {
		configContent := `name: test-project
container:
  provider: invalid-provider
  image: alpine:latest
`
		if err := os.WriteFile(ConfigFileName, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		_, err := LoadConfig()
		if err == nil {
			t.Error("LoadConfig() should return error for invalid container provider")
		}
	})
}

func TestConfig_GetScript(t *testing.T) {
	config := &Config{
		Shell: Shell{
			Scripts: []Script{
				{Name: "test", Commands: []string{"echo test"}},
				{Name: "build", Commands: []string{"go build"}},
			},
		},
	}

	t.Run("existing script", func(t *testing.T) {
		script, exists := config.GetScript("test")
		if !exists {
			t.Error("GetScript() should return true for existing script")
		}
		if script.Name != "test" {
			t.Errorf("Expected script name 'test', got '%s'", script.Name)
		}
		if len(script.Commands) != 1 || script.Commands[0] != "echo test" {
			t.Errorf("Expected commands ['echo test'], got %v", script.Commands)
		}
	})

	t.Run("non-existing script", func(t *testing.T) {
		_, exists := config.GetScript("nonexistent")
		if exists {
			t.Error("GetScript() should return false for non-existing script")
		}
	})
}
