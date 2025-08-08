package mikoshell

import (
	"os"
	"strings"
	"testing"
)

// MockContainerProvider implements ContainerProvider for testing
type MockContainerProvider struct{}

func (m *MockContainerProvider) IsAvailable() bool {
	return true // Always available in tests
}

func (m *MockContainerProvider) BuildImage(cfg *Config, tag string) error {
	return nil // Mock successful build
}

func (m *MockContainerProvider) RunCommand(cfg *Config, tag string, command []string) error {
	return nil // Mock successful command
}

func (m *MockContainerProvider) RunShell(cfg *Config, tag string) error {
	return nil // Mock successful shell
}

func (m *MockContainerProvider) RunShellWithStartup(cfg *Config, tag string) error {
	return nil // Mock successful shell with startup
}

func (m *MockContainerProvider) ImageExists(tag string) bool {
	return true // Always exists in tests
}

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() should return a non-nil client")
	}

	if client.workingDir == "" {
		t.Error("NewClient() should set working directory")
	}

	// Verify working directory is set correctly
	expectedDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	if client.workingDir != expectedDir {
		t.Errorf("Expected working directory '%s', got '%s'", expectedDir, client.workingDir)
	}
}

func TestClient_InitProject(t *testing.T) {
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
	tempDir, err := os.MkdirTemp("", "test-init-project")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	t.Run("successful init", func(t *testing.T) {
		err := client.InitProject(false)
		if err != nil {
			t.Fatalf("InitProject() failed: %v", err)
		}

		// Check if config file was created
		if !ConfigExists() {
			t.Error("InitProject() should create config file")
		}

		// Check if config file has expected content
		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load created config: %v", err)
		}

		expectedName := GetCurrentDirName()
		if config.Name != expectedName {
			t.Errorf("Expected name '%s', got '%s'", expectedName, config.Name)
		}

		if config.Container.Provider != "docker" {
			t.Errorf("Expected provider 'docker', got '%s'", config.Container.Provider)
		}

		if config.Container.Image != "alpine:latest" {
			t.Errorf("Expected image 'alpine:latest', got '%s'", config.Container.Image)
		}
	})

	t.Run("config already exists", func(t *testing.T) {
		// Try to init again - should fail
		err := client.InitProject(false)
		if err == nil {
			t.Error("InitProject() should fail when config already exists")
		}
	})

	// Clean up for dockerfile test
	os.Remove(ConfigFileName)

	t.Run("successful init with dockerfile", func(t *testing.T) {
		err := client.InitProject(true)
		if err != nil {
			t.Fatalf("InitProject(true) failed: %v", err)
		}

		// Check if config file was created
		if !ConfigExists() {
			t.Error("InitProject(true) should create config file")
		}

		// Check if Dockerfile was created
		if _, err := os.Stat("Dockerfile"); os.IsNotExist(err) {
			t.Error("InitProject(true) should create Dockerfile")
		}

		// Load and verify config content
		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load created config: %v", err)
		}

		if config.Container.Build == nil {
			t.Error("Expected build configuration, got nil")
		}

		if config.Container.Build.Dockerfile != "./Dockerfile" {
			t.Errorf("Expected dockerfile './Dockerfile', got '%s'", config.Container.Build.Dockerfile)
		}

		// Clean up Dockerfile
		os.Remove("Dockerfile")
	})
}

func TestClient_LoadConfig(t *testing.T) {
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
	tempDir, err := os.MkdirTemp("", "test-load-config")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	t.Run("no config file", func(t *testing.T) {
		err := client.LoadConfig()
		if err == nil {
			t.Error("LoadConfig() should fail when no config file exists")
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		// Create config file
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
      cmds:
        - echo "test command"
`
		if err := os.WriteFile(ConfigFileName, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Set mock provider before loading config
		client.SetProvider(&MockContainerProvider{})

		err := client.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		if client.config == nil {
			t.Error("LoadConfig() should set config")
		}

		if client.provider == nil {
			t.Error("LoadConfig() should set provider")
		}

		if client.config.Name != "test-project" {
			t.Errorf("Expected config name 'test-project', got '%s'", client.config.Name)
		}
	})

	t.Run("invalid container provider", func(t *testing.T) {
		// Create config with invalid provider
		configContent := `name: test-project
container:
  provider: invalid-provider
  image: alpine:latest
`
		if err := os.WriteFile(ConfigFileName, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		err := client.LoadConfig()
		if err == nil {
			t.Error("LoadConfig() should fail for invalid container provider")
		}
	})
}

func TestClient_GetImageTag(t *testing.T) {
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
	tempDir, err := os.MkdirTemp("", "test-image-tag")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	t.Run("no config loaded", func(t *testing.T) {
		_, err := client.GetImageTag()
		if err == nil {
			t.Error("GetImageTag() should fail when no config is loaded")
		}
	})

	t.Run("with config loaded", func(t *testing.T) {
		// Create config file
		configContent := `name: test-project
container:
  provider: docker
  image: alpine:latest
`
		if err := os.WriteFile(ConfigFileName, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Set mock provider before loading config
		client.SetProvider(&MockContainerProvider{})

		err := client.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		tag, err := client.GetImageTag()
		if err != nil {
			t.Fatalf("GetImageTag() failed: %v", err)
		}

		if tag == "" {
			t.Error("GetImageTag() should return non-empty tag")
		}

		// Tag should start with project name
		if !strings.HasPrefix(tag, "test-project:") {
			t.Errorf("Expected tag to start with 'test-project:', got '%s'", tag)
		}
	})
}

func TestClient_GetConfig(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	// Initially should be nil
	if client.GetConfig() != nil {
		t.Error("GetConfig() should return nil when no config is loaded")
	}

	// Set a mock config
	client.config = &Config{
		Name: "test",
		Container: Container{
			Provider: "docker",
			Image:    "alpine:latest",
		},
	}

	config := client.GetConfig()
	if config == nil {
		t.Fatal("GetConfig() should return config when one is set")
	}

	if config.Name != "test" {
		t.Errorf("Expected config name 'test', got '%s'", config.Name)
	}
}

func TestClient_RunCommand(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	t.Run("no config loaded", func(t *testing.T) {
		err := client.RunCommand([]string{"echo", "test"})
		if err == nil {
			t.Error("RunCommand() should fail when no config is loaded")
		}
	})

	t.Run("no command specified", func(t *testing.T) {
		client.config = &Config{Name: "test"}
		err := client.RunCommand([]string{})
		if err == nil {
			t.Error("RunCommand() should fail when no command is specified")
		}
	})
}

func TestClient_OpenShell(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	t.Run("no config loaded", func(t *testing.T) {
		err := client.OpenShell()
		if err == nil {
			t.Error("OpenShell() should fail when no config is loaded")
		}
	})
}
