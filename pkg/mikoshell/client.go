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

	// Initialize the container provider only if not already set (for testing)
	if c.provider == nil {
		provider, err := NewContainerProvider(cfg.Container.Provider)
		if err != nil {
			return fmt.Errorf("failed to create container provider: %w", err)
		}

		if !provider.IsAvailable() {
			return fmt.Errorf("container provider '%s' is not available. Please install %s first", cfg.Container.Provider, cfg.Container.Provider)
		}

		c.provider = provider
	}
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

	// Initialize the container provider only if not already set (for testing)
	if c.provider == nil {
		provider, err := NewContainerProvider(cfg.Container.Provider)
		if err != nil {
			return fmt.Errorf("failed to create container provider: %w", err)
		}

		if !provider.IsAvailable() {
			return fmt.Errorf("container provider '%s' is not available. Please install %s first", cfg.Container.Provider, cfg.Container.Provider)
		}

		c.provider = provider
	}
	return nil
}

// InitProject creates a new miko-shell.yaml file
func (c *Client) InitProject(useDockerfile bool) error {
	if ConfigExists() {
		return fmt.Errorf("miko-shell.yaml already exists in current directory")
	}

	// Get the normalized directory name
	projectName := GetCurrentDirName()

	var defaultConfig string
	if useDockerfile {
		defaultConfig = c.generateDockerfileConfig(projectName)
	} else {
		defaultConfig = c.generateImageConfig(projectName)
	}

	if err := os.WriteFile(ConfigFileName, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Create Dockerfile if using --dockerfile option
	if useDockerfile {
		if err := c.createSampleDockerfile(); err != nil {
			return fmt.Errorf("failed to create Dockerfile: %w", err)
		}
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
		// Run the script commands with parameters
		scriptArgs := args[1:] // Get the remaining arguments
		commandStr := script.GetCommandsAsStringWithArgs(scriptArgs)
		command := []string{"/bin/sh", "-c", commandStr}
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

	return c.provider.RunShellWithStartup(c.config, tag)
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
	return s.GetCommandsAsStringWithArgs([]string{})
}

// GetCommandsAsStringWithArgs converts Commands field to a shell command string with arguments
func (s *Script) GetCommandsAsStringWithArgs(args []string) string {
	// Join all commands with &&
	command := strings.Join(s.Commands, " && ")

	// If there are no arguments, return the command as is
	if len(args) == 0 {
		return command
	}

	// Prepare argument variables for the shell script
	argSetup := "set -- "
	for _, arg := range args {
		// Escape single quotes in arguments
		escaped := strings.ReplaceAll(arg, "'", "'\"'\"'")
		argSetup += "'" + escaped + "' "
	}
	
	// Combine argument setup with the actual command
	return argSetup + "; " + command
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

// generateImageConfig generates configuration using pre-built image
func (c *Client) generateImageConfig(projectName string) string {
	return `name: ` + projectName + `
container:
  provider: docker
  image: alpine:latest
  setup:
    - apk add --no-cache curl git
shell:
  startup:
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
}

// generateDockerfileConfig generates configuration using custom Dockerfile
func (c *Client) generateDockerfileConfig(projectName string) string {
	return `name: ` + projectName + `
container:
  provider: docker
  build:
    dockerfile: ./Dockerfile
    context: .
    args:
      VERSION: "1.0"
shell:
  startup:
    - echo "Welcome to your custom development environment!"
    - echo "Project ` + projectName + `"
    - pwd
  scripts:
    - name: hello
      description: "Say hello and show system info"
      commands: |
        echo "Hello from miko-shell!"
        uname -a
        df -h /
    - name: example
      description: "Example custom command"
      commands:
        - echo "This is a custom environment built from Dockerfile"
    - name: build
      description: "Build the project"
      commands:
        - echo "Building project..."
        - echo "Build completed successfully!"
`
}

// createSampleDockerfile creates a sample Dockerfile
func (c *Client) createSampleDockerfile() error {
	dockerfileContent := `FROM alpine:latest

# Install basic tools
RUN apk add --no-cache curl git

# Set working directory
WORKDIR /workspace

# Add sample setup
RUN echo "Custom Dockerfile setup completed"

# Default command
CMD ["sh"]
`

	return os.WriteFile("Dockerfile", []byte(dockerfileContent), 0644)
}

// SetProvider sets the container provider (useful for testing)
func (c *Client) SetProvider(provider ContainerProvider) {
	c.provider = provider
}
