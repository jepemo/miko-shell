package mikoshell

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

const ConfigFileName = "miko-shell.yaml"

// Config represents the project configuration
type Config struct {
	Name      string    `yaml:"name"`
	Container Container `yaml:"container"`
	Shell     Shell     `yaml:"shell"`
}

// Container represents the container configuration
type Container struct {
	Provider string          `yaml:"provider"`
	Image    string          `yaml:"image,omitempty"`
	Build    *ContainerBuild `yaml:"build,omitempty"`
	Setup    []string        `yaml:"setup,omitempty"`
}

// ContainerBuild represents custom image build configuration
type ContainerBuild struct {
	Dockerfile string            `yaml:"dockerfile"`
	Context    string            `yaml:"context,omitempty"`
	Args       map[string]string `yaml:"args,omitempty"`
}

// Shell represents the shell configuration
type Shell struct {
	InitHook []string `yaml:"startup"`
	Scripts  []Script `yaml:"scripts"`
}

// Script represents a shell script
type Script struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Commands    []string `yaml:"commands"`
}

// ConfigExists checks if the configuration file exists in the current directory
func ConfigExists() bool {
	_, err := os.Stat(ConfigFileName)
	return err == nil
}

// LoadConfig loads the configuration from miko-shell.yaml
func LoadConfig() (*Config, error) {
	if !ConfigExists() {
		return nil, fmt.Errorf("miko-shell.yaml not found. Run 'miko-shell init' first")
	}

	data, err := os.ReadFile(ConfigFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Container.Provider == "" {
		config.Container.Provider = "docker"
	}

	// Validate container provider
	if config.Container.Provider != "docker" && config.Container.Provider != "podman" {
		return nil, fmt.Errorf("invalid provider: %s. Must be 'docker' or 'podman'", config.Container.Provider)
	}

	// Validate that either image or build is specified
	if config.Container.Image == "" && config.Container.Build == nil {
		return nil, fmt.Errorf("either 'container.image' or 'container.build' must be specified")
	}

	// Validate build configuration if present
	if config.Container.Build != nil {
		if config.Container.Build.Dockerfile == "" {
			return nil, fmt.Errorf("'container.build.dockerfile' is required when using custom build")
		}
		if config.Container.Build.Context == "" {
			config.Container.Build.Context = "."
		}
	}

	return &config, nil
}

// LoadConfigFromFile loads the configuration from a specific file
func LoadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file '%s': %w", filePath, err)
	}

	// Set defaults
	if config.Container.Provider == "" {
		config.Container.Provider = "docker"
	}

	// Validate container provider
	if config.Container.Provider != "docker" && config.Container.Provider != "podman" {
		return nil, fmt.Errorf("invalid provider: %s. Must be 'docker' or 'podman'", config.Container.Provider)
	}

	// Validate that either image or build is specified
	if config.Container.Image == "" && config.Container.Build == nil {
		return nil, fmt.Errorf("either 'container.image' or 'container.build' must be specified")
	}

	// Validate build configuration if present
	if config.Container.Build != nil {
		if config.Container.Build.Dockerfile == "" {
			return nil, fmt.Errorf("'container.build.dockerfile' is required when using custom build")
		}
		if config.Container.Build.Context == "" {
			config.Container.Build.Context = "."
		}
	}

	return &config, nil
}

// GetConfigHash calculates a hash of the configuration file
func GetConfigHash() (string, error) {
	return GetConfigHashFromFile(ConfigFileName)
}

// GetConfigHashFromFile calculates a hash of the specified configuration file
func GetConfigHashFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))[:12], nil
}

// GetScript returns a script by name
func (c *Config) GetScript(name string) (*Script, bool) {
	for _, script := range c.Shell.Scripts {
		if script.Name == name {
			return &script, true
		}
	}
	return nil, false
}

// NormalizeName normalizes a directory name to be used as a container image name
func NormalizeName(name string) string {
	// Remove accents and normalize unicode
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, name)

	// Convert to lowercase
	normalized = strings.ToLower(normalized)

	// Replace spaces and special characters with dashes
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	normalized = reg.ReplaceAllString(normalized, "-")

	// Remove leading and trailing dashes
	normalized = strings.Trim(normalized, "-")

	// If empty after normalization, use default
	if normalized == "" {
		normalized = "project"
	}

	return normalized
}

// GetCurrentDirName returns the normalized name of the current directory
func GetCurrentDirName() string {
	workingDir, err := os.Getwd()
	if err != nil {
		return "project"
	}

	dirName := filepath.Base(workingDir)
	return NormalizeName(dirName)
}

// detectHostPlatform detects the host OS and architecture
func detectHostPlatform() (string, string, error) {
	var hostOS, hostArch string

	// Detect OS
	switch runtime.GOOS {
	case "linux":
		hostOS = "linux"
	case "darwin":
		hostOS = "darwin"
	case "windows":
		hostOS = "windows"
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Detect architecture
	switch runtime.GOARCH {
	case "amd64":
		hostArch = "amd64"
	case "arm64":
		hostArch = "arm64"
	case "arm":
		hostArch = "armv6l"
	default:
		return "", "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	return hostOS, hostArch, nil
}
