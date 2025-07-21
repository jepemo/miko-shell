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
	return d.buildImage(cfg, tag)
}

func (d *DockerProvider) RunCommand(cfg *Config, tag string, command []string) error {
	return d.runContainer(cfg, tag, command, false)
}

func (d *DockerProvider) RunShell(cfg *Config, tag string) error {
	return d.runContainer(cfg, tag, []string{"/bin/sh"}, true)
}

func (d *DockerProvider) ImageExists(tag string) bool {
	cmd := exec.Command("docker", "image", "inspect", tag)
	return cmd.Run() == nil
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
	
	dockerfile.WriteString(fmt.Sprintf("FROM %s\n", cfg.Image))
	dockerfile.WriteString("WORKDIR /workspace\n")
	
	// Add pre-install commands
	for _, cmd := range cfg.PreInstall {
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", cmd))
	}
	
	// Add init hook commands
	if len(cfg.Shell.InitHook) > 0 {
		dockerfile.WriteString("RUN ")
		dockerfile.WriteString(strings.Join(cfg.Shell.InitHook, " && "))
		dockerfile.WriteString("\n")
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
	return p.buildImage(cfg, tag)
}

func (p *PodmanProvider) RunCommand(cfg *Config, tag string, command []string) error {
	return p.runContainer(cfg, tag, command, false)
}

func (p *PodmanProvider) RunShell(cfg *Config, tag string) error {
	return p.runContainer(cfg, tag, []string{"/bin/sh"}, true)
}

func (p *PodmanProvider) ImageExists(tag string) bool {
	cmd := exec.Command("podman", "image", "inspect", tag)
	return cmd.Run() == nil
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
	
	dockerfile.WriteString(fmt.Sprintf("FROM %s\n", cfg.Image))
	dockerfile.WriteString("WORKDIR /workspace\n")
	
	// Add pre-install commands
	for _, cmd := range cfg.PreInstall {
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", cmd))
	}
	
	// Add init hook commands
	if len(cfg.Shell.InitHook) > 0 {
		dockerfile.WriteString("RUN ")
		dockerfile.WriteString(strings.Join(cfg.Shell.InitHook, " && "))
		dockerfile.WriteString("\n")
	}
	
	dockerfile.WriteString("CMD [\"/bin/sh\"]\n")
	
	return dockerfile.String()
}
