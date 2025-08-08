package mikoshell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ContainerProvider defines the interface for container providers
type ContainerProvider interface {
	IsAvailable() bool
	BuildImage(cfg *Config, tag string) error
	RunCommand(cfg *Config, tag string, command []string) error
	RunShell(cfg *Config, tag string) error
	RunShellWithStartup(cfg *Config, tag string) error
	ImageExists(tag string) bool
}

// DockerProvider implements the ContainerProvider interface for Docker
type DockerProvider struct{}

// PodmanProvider implements the ContainerProvider interface for Podman
type PodmanProvider struct{}

// NewContainerProvider creates a new container provider
func NewContainerProvider(providerName string) (ContainerProvider, error) {
	switch providerName {
	case "docker":
		return &DockerProvider{}, nil
	case "podman":
		return &PodmanProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported container provider: %s", providerName)
	}
}

// Docker Provider Implementation
func (d *DockerProvider) IsAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func (d *DockerProvider) BuildImage(cfg *Config, tag string) error {
	// First, build custom image if needed
	if cfg.Container.Build != nil {
		if err := d.buildCustomImage(cfg); err != nil {
			return fmt.Errorf("failed to build custom image: %w", err)
		}
	}

	return d.buildImage(cfg, tag)
}

func (d *DockerProvider) RunCommand(cfg *Config, tag string, command []string) error {
	return d.runContainer(cfg, tag, command, false)
}

func (d *DockerProvider) RunShell(cfg *Config, tag string) error {
	return d.runContainer(cfg, tag, []string{"/bin/sh"}, true)
}

func (d *DockerProvider) RunShellWithStartup(cfg *Config, tag string) error {
	// If no startup commands are defined, just run the shell
	if len(cfg.Shell.InitHook) == 0 {
		return d.RunShell(cfg, tag)
	}

	// Create a startup script that will be written to the container
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	script.WriteString("set -e\n") // Exit on any error
	script.WriteString("\n")
	
	// Add startup commands
	for _, cmd := range cfg.Shell.InitHook {
		script.WriteString("# Startup command\n")
		script.WriteString(cmd + "\n")
		script.WriteString("\n")
	}
	
	// Start interactive shell after startup commands
	script.WriteString("# Start interactive shell\n")
	script.WriteString("exec /bin/sh\n")

	// Create the script using a here-document to avoid escaping issues
	shellCommand := fmt.Sprintf(`cat > /tmp/startup.sh << 'MIKO_SCRIPT_EOF'
%s
MIKO_SCRIPT_EOF
chmod +x /tmp/startup.sh
exec /tmp/startup.sh`, script.String())

	// Run the script
	return d.runContainer(cfg, tag, []string{"/bin/sh", "-c", shellCommand}, true)
}

func (d *DockerProvider) ImageExists(tag string) bool {
	cmd := exec.Command("docker", "image", "inspect", tag)
	return cmd.Run() == nil
}

func (d *DockerProvider) buildCustomImage(cfg *Config) error {
	build := cfg.Container.Build
	customTag := cfg.Name + ":custom"

	// Check if custom image already exists
	if d.ImageExists(customTag) {
		return nil
	}

	args := []string{"build", "-t", customTag, "-f", build.Dockerfile}

	// Add build args if specified
	for key, value := range build.Args {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Add context path
	args = append(args, build.Context)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (d *DockerProvider) buildImage(cfg *Config, tag string) error {
	dockerfile := d.generateDockerfile(cfg)

	cmd := exec.Command("docker", "build", "-t", tag, "-f", "-", ".")
	cmd.Stdin = strings.NewReader(dockerfile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (d *DockerProvider) runContainer(cfg *Config, tag string, command []string, interactive bool) error {
	args := []string{"run", "--rm"}

	if interactive {
		args = append(args, "-it")
	}

	// Add host platform environment variables
	hostOS, hostArch, err := detectHostPlatform()
	if err == nil {
		args = append(args, "-e", fmt.Sprintf("MIKO_HOST_OS=%s", hostOS))
		args = append(args, "-e", fmt.Sprintf("MIKO_HOST_ARCH=%s", hostArch))
	}

	// Mount current directory
	workingDir, _ := os.Getwd()
	args = append(args, "-v", fmt.Sprintf("%s:/workspace", workingDir))
	args = append(args, "-w", "/workspace")

	args = append(args, tag)
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (d *DockerProvider) generateDockerfile(cfg *Config) string {
	var dockerfile strings.Builder

	// Handle custom build or base image
	if cfg.Container.Build != nil {
		// For custom builds, we'll build the custom image first
		// This function generates a runtime Dockerfile that uses the custom image
		dockerfile.WriteString(fmt.Sprintf("FROM %s\n", cfg.Name+":custom"))
	} else {
		dockerfile.WriteString(fmt.Sprintf("FROM %s\n", cfg.Container.Image))
	}

	dockerfile.WriteString("WORKDIR /workspace\n")

	// Add setup commands
	for _, cmd := range cfg.Container.Setup {
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", cmd))
	}

	dockerfile.WriteString("CMD [\"/bin/sh\"]\n")

	return dockerfile.String()
}

// Podman Provider Implementation
func (p *PodmanProvider) IsAvailable() bool {
	_, err := exec.LookPath("podman")
	return err == nil
}

func (p *PodmanProvider) BuildImage(cfg *Config, tag string) error {
	// First, build custom image if needed
	if cfg.Container.Build != nil {
		if err := p.buildCustomImage(cfg); err != nil {
			return fmt.Errorf("failed to build custom image: %w", err)
		}
	}

	return p.buildImage(cfg, tag)
}

func (p *PodmanProvider) RunCommand(cfg *Config, tag string, command []string) error {
	return p.runContainer(cfg, tag, command, false)
}

func (p *PodmanProvider) RunShell(cfg *Config, tag string) error {
	return p.runContainer(cfg, tag, []string{"/bin/sh"}, true)
}

func (p *PodmanProvider) RunShellWithStartup(cfg *Config, tag string) error {
	// If no startup commands are defined, just run the shell
	if len(cfg.Shell.InitHook) == 0 {
		return p.RunShell(cfg, tag)
	}

	// Create a startup script that will be written to the container
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	script.WriteString("set -e\n") // Exit on any error
	script.WriteString("\n")
	
	// Add startup commands
	for _, cmd := range cfg.Shell.InitHook {
		script.WriteString("# Startup command\n")
		script.WriteString(cmd + "\n")
		script.WriteString("\n")
	}
	
	// Start interactive shell after startup commands
	script.WriteString("# Start interactive shell\n")
	script.WriteString("exec /bin/sh\n")

	// Create the script using a here-document to avoid escaping issues
	shellCommand := fmt.Sprintf(`cat > /tmp/startup.sh << 'MIKO_SCRIPT_EOF'
%s
MIKO_SCRIPT_EOF
chmod +x /tmp/startup.sh
exec /tmp/startup.sh`, script.String())

	// Run the script
	return p.runContainer(cfg, tag, []string{"/bin/sh", "-c", shellCommand}, true)
}

func (p *PodmanProvider) ImageExists(tag string) bool {
	cmd := exec.Command("podman", "image", "inspect", tag)
	return cmd.Run() == nil
}

func (p *PodmanProvider) buildCustomImage(cfg *Config) error {
	build := cfg.Container.Build
	customTag := cfg.Name + ":custom"

	// Check if custom image already exists
	if p.ImageExists(customTag) {
		return nil
	}

	args := []string{"build", "-t", customTag, "-f", build.Dockerfile}

	// Add build args if specified
	for key, value := range build.Args {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Add context path
	args = append(args, build.Context)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (p *PodmanProvider) buildImage(cfg *Config, tag string) error {
	dockerfile := p.generateDockerfile(cfg)

	cmd := exec.Command("podman", "build", "-t", tag, "-f", "-", ".")
	cmd.Stdin = strings.NewReader(dockerfile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (p *PodmanProvider) runContainer(cfg *Config, tag string, command []string, interactive bool) error {
	args := []string{"run", "--rm"}

	if interactive {
		args = append(args, "-it")
	}

	// Add host platform environment variables
	hostOS, hostArch, err := detectHostPlatform()
	if err == nil {
		args = append(args, "-e", fmt.Sprintf("MIKO_HOST_OS=%s", hostOS))
		args = append(args, "-e", fmt.Sprintf("MIKO_HOST_ARCH=%s", hostArch))
	}

	// Mount current directory
	workingDir, _ := os.Getwd()
	args = append(args, "-v", fmt.Sprintf("%s:/workspace", workingDir))
	args = append(args, "-w", "/workspace")

	args = append(args, tag)
	args = append(args, command...)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (p *PodmanProvider) generateDockerfile(cfg *Config) string {
	var dockerfile strings.Builder

	// Handle custom build or base image
	if cfg.Container.Build != nil {
		dockerfile.WriteString(fmt.Sprintf("FROM %s\n", cfg.Name+":custom"))
	} else {
		dockerfile.WriteString(fmt.Sprintf("FROM %s\n", cfg.Container.Image))
	}

	dockerfile.WriteString("WORKDIR /workspace\n")

	// Add setup commands
	for _, cmd := range cfg.Container.Setup {
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", cmd))
	}

	dockerfile.WriteString("CMD [\"/bin/sh\"]\n")

	return dockerfile.String()
}
