package mikoshell

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

const ConfigFileName = "dev-config.yaml"

// Config represents the project configuration
type Config struct {
	Name              string   `yaml:"name"`
	ContainerProvider string   `yaml:"container-provider"`
	Image            string   `yaml:"image"`
	PreInstall       []string `yaml:"pre-install"`
	Shell            Shell    `yaml:"shell"`
}

// Shell represents the shell configuration
type Shell struct {
	InitHook []string `yaml:"init-hook"`
	Scripts  []Script `yaml:"scripts"`
}

// Script represents a shell script
type Script struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description,omitempty"`
	Commands    interface{} `yaml:"commands"`
}

// ConfigExists checks if the configuration file exists in the current directory
func ConfigExists() bool {
	_, err := os.Stat(ConfigFileName)
	return err == nil
}

// LoadConfig loads the configuration from dev-config.yaml
func LoadConfig() (*Config, error) {
	if !ConfigExists() {
		return nil, fmt.Errorf("dev-config.yaml not found. Run 'miko-shell init' first")
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
	if config.ContainerProvider == "" {
		config.ContainerProvider = "docker"
	}

	// Validate container provider
	if config.ContainerProvider != "docker" && config.ContainerProvider != "podman" {
		return nil, fmt.Errorf("invalid container-provider: %s. Must be 'docker' or 'podman'", config.ContainerProvider)
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
	if config.ContainerProvider == "" {
		config.ContainerProvider = "docker"
	}

	// Validate container provider
	if config.ContainerProvider != "docker" && config.ContainerProvider != "podman" {
		return nil, fmt.Errorf("invalid container-provider: %s. Must be 'docker' or 'podman'", config.ContainerProvider)
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
