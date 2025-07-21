package mikoshell

import (
	"fmt"
	"os"
	"strings"
)

// Client provides the main functionality of the miko-shell tool
type Client struct {
	workingDir string
	config     *Config
	provider   ContainerProvider
	configFile string
}

// NewClient creates a new miko-shell client instance
func NewClient() (*Client, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	return &Client{
		workingDir: workingDir,
	}, nil
}

// LoadConfig loads the configuration file
func (c *Client) LoadConfig() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	c.config = cfg
	c.configFile = ConfigFileName

	// Initialize the container provider
	provider, err := NewContainerProvider(cfg.ContainerProvider)
	if err != nil {
		return fmt.Errorf("failed to create container provider: %w", err)
	}

	if !provider.IsAvailable() {
		return fmt.Errorf("container provider '%s' is not available. Please install %s first", cfg.ContainerProvider, cfg.ContainerProvider)
	}

	c.provider = provider
	return nil
}

// LoadConfigFromFile loads the configuration from a specific file
func (c *Client) LoadConfigFromFile(filePath string) error {
	cfg, err := LoadConfigFromFile(filePath)
	if err != nil {
		return err
	}

	c.config = cfg
	c.configFile = filePath

	// Initialize the container provider
	provider, err := NewContainerProvider(cfg.ContainerProvider)
	if err != nil {
		return fmt.Errorf("failed to create container provider: %w", err)
	}

	if !provider.IsAvailable() {
		return fmt.Errorf("container provider '%s' is not available. Please install %s first", cfg.ContainerProvider, cfg.ContainerProvider)
	}

	c.provider = provider
	return nil
}

// InitProject creates a new dev-config.yaml file
func (c *Client) InitProject() error {
	if ConfigExists() {
		return fmt.Errorf("dev-config.yaml already exists in current directory")
	}

	// Get the normalized directory name
	projectName := GetCurrentDirName()

	defaultConfig := `name: ` + projectName + `
container-provider: docker
image: alpine:latest
pre-install:
  - apk add --no-cache curl git
shell:
  init-hook:
    - echo "Welcome to your development environment!"
    - echo "Project ` + projectName + `"
    - pwd
  scripts:
    - name: hello
      description: "Say hello and show system info"
      commands: |
        echo "Hello from miko-shell!"
        uname -a
        df -h /
    - name: test
      description: "Run a simple test"
      commands:
        - echo "Running tests..."
        - echo "All tests passed!"
    - name: build
      description: "Build the project"
      commands:
        - echo "Building project..."
        - echo "Build completed successfully!"
`

	if err := os.WriteFile(ConfigFileName, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	return nil
}

// BuildImage builds the container image
func (c *Client) BuildImage() (string, error) {
	if c.config == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	hash, err := GetConfigHashFromFile(c.configFile)
	if err != nil {
		return "", fmt.Errorf("failed to calculate config hash: %w", err)
	}

	tag := fmt.Sprintf("%s:%s", c.config.Name, hash)

	if err := c.provider.BuildImage(c.config, tag); err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}

	return tag, nil
}

// RunCommand executes a command in the container
func (c *Client) RunCommand(args []string) error {
	if c.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	tag, err := c.ensureImageExists()
	if err != nil {
		return err
	}

	// Check if the command is a script
	commandName := args[0]
	if script, exists := c.config.GetScript(commandName); exists {
		// Run the script commands
		command := []string{"/bin/sh", "-c", script.GetCommandsAsString()}
		return c.provider.RunCommand(c.config, tag, command)
	}

	// Run the command directly
	return c.provider.RunCommand(c.config, tag, args)
}

// OpenShell opens an interactive shell in the container
func (c *Client) OpenShell() error {
	if c.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	tag, err := c.ensureImageExists()
	if err != nil {
		return err
	}

	return c.provider.RunShell(c.config, tag)
}

// GetImageTag returns the current image tag
func (c *Client) GetImageTag() (string, error) {
	if c.config == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	hash, err := GetConfigHashFromFile(c.configFile)
	if err != nil {
		return "", fmt.Errorf("failed to calculate config hash: %w", err)
	}

	return fmt.Sprintf("%s:%s", c.config.Name, hash), nil
}

// GetCommandsAsString converts Commands field to a shell command string
func (s *Script) GetCommandsAsString() string {
	switch v := s.Commands.(type) {
	case []interface{}:
		// Handle array of commands
		var commands []string
		for _, cmd := range v {
			commands = append(commands, fmt.Sprintf("%v", cmd))
		}
		return strings.Join(commands, " && ")
	case string:
		// Handle single string (multiline block)
		return v
	case []string:
		// Handle array of strings
		return strings.Join(v, " && ")
	default:
		return ""
	}
}

// ListScripts displays all available scripts with their descriptions
func (c *Client) ListScripts() error {
	if c.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if len(c.config.Shell.Scripts) == 0 {
		fmt.Println("No scripts available in this configuration.")
		return nil
	}

	fmt.Println("Available scripts:")
	fmt.Println()

	for _, script := range c.config.Shell.Scripts {
		if script.Description != "" {
			fmt.Printf("  %s - %s\n", script.Name, script.Description)
		} else {
			fmt.Printf("  %s\n", script.Name)
		}
	}

	fmt.Println()
	fmt.Println("Usage: ./miko-shell run <script-name>")
	return nil
}

// GetConfig returns the current configuration
func (c *Client) GetConfig() *Config {
	return c.config
}

// ensureImageExists checks if the image exists and builds it if necessary
func (c *Client) ensureImageExists() (string, error) {
	tag, err := c.GetImageTag()
	if err != nil {
		return "", err
	}

	if !c.provider.ImageExists(tag) {
		if _, err := c.BuildImage(); err != nil {
			return "", fmt.Errorf("failed to build image: %w", err)
		}
	}

	return tag, nil
}
