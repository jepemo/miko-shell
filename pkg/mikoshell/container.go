package mikoshell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ContainerProvider defines the interface for container providers
type ContainerProvider interface {
	IsAvailable() bool
	BuildImage(cfg *Config, tag string) error
	RunCommand(cfg *Config, tag string, command []string) error
	RunShell(cfg *Config, tag string) error
	RunShellWithStartup(cfg *Config, tag string) error
	ImageExists(tag string) bool
	RemoveImage(tag string) error
	ListImages() ([]ImageListItem, error)
	CleanImages(all bool) ([]string, error)
	GetImageInfo(imageID string) (*ImageInfo, error)
	GetPruneInfo() (*PruneInfo, error)
	PruneImages() (*PruneResult, error)
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
	// If no startup commands and no scripts are defined, just run the shell
	if len(cfg.Shell.InitHook) == 0 && len(cfg.Shell.Scripts) == 0 {
		return d.RunShell(cfg, tag)
	}

	// 1. Script de startup original
	var startupScript strings.Builder
	startupScript.WriteString("#!/bin/sh\n")
	startupScript.WriteString("set -e\n\n")

	// Agregar comandos de startup
	for _, cmd := range cfg.Shell.InitHook {
		startupScript.WriteString(cmd + "\n\n")
	}

	// 2. Generar wrapper miko-shell
	var mikoShell strings.Builder
	mikoShell.WriteString("#!/bin/sh\n")
	mikoShell.WriteString("set -e\n\n")

	// Configurar PATH para incluir herramientas de Go
	mikoShell.WriteString("# Ensure Go tools are in PATH\n")
	mikoShell.WriteString("export PATH=\"/go/bin:/usr/local/go/bin:$PATH\"\n\n")

	// Función de ayuda
	mikoShell.WriteString("show_help() {\n")
	mikoShell.WriteString("  echo \"Miko Shell - Container development environment\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Usage:\"\n")
	mikoShell.WriteString("  echo \"  miko-shell [command]\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Available Commands:\"\n")
	mikoShell.WriteString("  echo \"  help        Show help for miko-shell\"\n")
	mikoShell.WriteString("  echo \"  list        List available scripts\"\n")
	mikoShell.WriteString("  echo \"  run         Run a script or command inside the container\"\n")
	mikoShell.WriteString("  echo \"  version     Show miko-shell version\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Run 'miko-shell run --help' for information about running scripts\"\n")
	mikoShell.WriteString("}\n\n")

	// Función para listar scripts
	mikoShell.WriteString("list_scripts() {\n")
	mikoShell.WriteString("  echo \"Available scripts:\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	for _, script := range cfg.Shell.Scripts {
		desc := script.Description
		if desc == "" {
			desc = script.Name
		}
		mikoShell.WriteString(fmt.Sprintf("  echo \"  %-15s %s\"\n", script.Name, desc))
	}
	mikoShell.WriteString("}\n\n")

	// Función para ejecutar scripts
	mikoShell.WriteString("run_script() {\n")
	mikoShell.WriteString("  script_name=\"$1\"\n")
	mikoShell.WriteString("  shift\n\n")
	mikoShell.WriteString("  case \"$script_name\" in\n")

	// Agregar case para cada script
	for _, script := range cfg.Shell.Scripts {
		mikoShell.WriteString(fmt.Sprintf("    %s)\n", script.Name))
		mikoShell.WriteString("      # Ejecutar script con argumentos pasados\n")

		// Exportar variables para los argumentos posicionales
		mikoShell.WriteString("      # Establecer argumentos posicionales\n")
		mikoShell.WriteString("      i=1\n")
		mikoShell.WriteString("      for arg in \"$@\"; do\n")
		mikoShell.WriteString("        export \"_MIKO_ARG_${i}=$arg\"\n")
		mikoShell.WriteString("        i=$((i+1))\n")
		mikoShell.WriteString("      done\n\n")

		// Ejecutar cada comando del script, reemplazando $1, $2, etc. con las variables exportadas
		for _, cmd := range script.Commands {
			// Reemplazar $1, $2, etc. con las variables _MIKO_ARG_1, _MIKO_ARG_2, etc.
			processedCmd := cmd
			for i := 1; i <= 9; i++ {
				placeholder := fmt.Sprintf("$%d", i)
				replacement := fmt.Sprintf("${_MIKO_ARG_%d:-}", i)
				processedCmd = strings.ReplaceAll(processedCmd, placeholder, replacement)
			}
			mikoShell.WriteString(fmt.Sprintf("      %s\n", processedCmd))
		}

		// Limpiar las variables de argumentos
		mikoShell.WriteString("\n      # Limpiar variables de argumentos\n")
		mikoShell.WriteString("      for j in $(seq 1 $((i-1))); do\n")
		mikoShell.WriteString("        unset \"_MIKO_ARG_${j}\"\n")
		mikoShell.WriteString("      done\n")

		mikoShell.WriteString("      return $?\n")
		mikoShell.WriteString("      ;;\n")
	}

	// Caso para comando directo (ejecuta el comando pasado directamente)
	mikoShell.WriteString("    --)\n")
	mikoShell.WriteString("      shift\n")
	mikoShell.WriteString("      \"$@\"\n")
	mikoShell.WriteString("      return $?\n")
	mikoShell.WriteString("      ;;\n")

	// Caso para script desconocido
	mikoShell.WriteString("    *)\n")
	mikoShell.WriteString("      echo \"Error: Unknown script '$script_name'\"\n")
	mikoShell.WriteString("      echo \"\"\n")
	mikoShell.WriteString("      list_scripts\n")
	mikoShell.WriteString("      return 1\n")
	mikoShell.WriteString("      ;;\n")
	mikoShell.WriteString("  esac\n")
	mikoShell.WriteString("}\n\n")

	// Función para mostrar ayuda de run
	mikoShell.WriteString("show_run_help() {\n")
	mikoShell.WriteString("  echo \"Run a script or command inside the container\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Usage:\"\n")
	mikoShell.WriteString("  echo \"  miko-shell run <script-name> [args...]  Run a script with optional arguments\"\n")
	mikoShell.WriteString("  echo \"  miko-shell run -- <command> [args...]   Run a direct command\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Available scripts:\"\n")
	mikoShell.WriteString("  echo \"\"\n")

	// Listar scripts disponibles
	for _, script := range cfg.Shell.Scripts {
		desc := script.Description
		if desc == "" {
			desc = script.Name
		}
		mikoShell.WriteString(fmt.Sprintf("  echo \"  %-15s %s\"\n", script.Name, desc))
	}
	mikoShell.WriteString("}\n\n")

	// Comando principal
	mikoShell.WriteString("# Detectar versión de la imagen\n")
	mikoShell.WriteString("MIKO_VERSION=\"$(cat /tmp/miko-version 2>/dev/null || echo 'dev')\"\n\n")
	mikoShell.WriteString("# Procesar comandos\n")
	mikoShell.WriteString("case \"$1\" in\n")

	// Comando run
	mikoShell.WriteString("  run)\n")
	mikoShell.WriteString("    shift\n")
	mikoShell.WriteString("    if [ \"$1\" = \"--help\" ] || [ \"$1\" = \"-h\" ]; then\n")
	mikoShell.WriteString("      show_run_help\n")
	mikoShell.WriteString("      exit 0\n")
	mikoShell.WriteString("    fi\n")
	mikoShell.WriteString("    if [ -z \"$1\" ]; then\n")
	mikoShell.WriteString("      echo \"Error: Missing script name or command\"\n")
	mikoShell.WriteString("      echo \"\"\n")
	mikoShell.WriteString("      show_run_help\n")
	mikoShell.WriteString("      exit 1\n")
	mikoShell.WriteString("    fi\n")
	mikoShell.WriteString("    run_script \"$@\"\n")
	mikoShell.WriteString("    exit $?\n")
	mikoShell.WriteString("    ;;\n")

	// Comando open (debe fallar dentro del contenedor)
	mikoShell.WriteString("  open)\n")
	mikoShell.WriteString("    echo \"Error: Already inside a miko-shell container\"\n")
	mikoShell.WriteString("    echo \"The 'open' command can only be used from outside the container\"\n")
	mikoShell.WriteString("    exit 1\n")
	mikoShell.WriteString("    ;;\n")

	// Comando list
	mikoShell.WriteString("  list)\n")
	mikoShell.WriteString("    list_scripts\n")
	mikoShell.WriteString("    exit 0\n")
	mikoShell.WriteString("    ;;\n")

	// Comando version
	mikoShell.WriteString("  version)\n")
	mikoShell.WriteString("    echo \"miko-shell version $MIKO_VERSION\"\n")
	mikoShell.WriteString("    exit 0\n")
	mikoShell.WriteString("    ;;\n")

	// Comando help o sin argumentos
	mikoShell.WriteString("  help|-h|--help|\"\")\n")
	mikoShell.WriteString("    show_help\n")
	mikoShell.WriteString("    exit 0\n")
	mikoShell.WriteString("    ;;\n")

	// Comando desconocido
	mikoShell.WriteString("  *)\n")
	mikoShell.WriteString("    echo \"Error: Unknown command '$1'\"\n")
	mikoShell.WriteString("    echo \"\"\n")
	mikoShell.WriteString("    show_help\n")
	mikoShell.WriteString("    exit 1\n")
	mikoShell.WriteString("    ;;\n")
	mikoShell.WriteString("esac\n")

	// Crear el comando completo que:
	// 1. Guarda la versión en un archivo
	// 2. Crea el script miko-shell
	// 3. Genera el autocompletado
	// 4. Ejecuta el script de startup
	version := "dev"
	if v := os.Getenv("MIKO_VERSION"); v != "" {
		version = v
	}

	shellCommand := fmt.Sprintf(`
# Save version information
echo "%s" > /tmp/miko-version

# Create the miko-shell wrapper
cat > /usr/local/bin/miko-shell << 'MIKO_WRAPPER_EOF'
%s
MIKO_WRAPPER_EOF
chmod +x /usr/local/bin/miko-shell

# Bash completion disabled for sh compatibility
# Bash completion for miko-shell (disabled for sh compatibility)
touch /etc/profile.d/miko-shell-completion.sh

# Setup PATH to include Go tools for all sessions
echo 'export PATH="/go/bin:/usr/local/go/bin:$PATH"' >> /etc/profile.d/miko-shell-path.sh

# Setup prompt to show we're in a miko-shell
echo 'PS1="[\[\e[1;32m\]miko-shell\[\e[0m\]] \w \$ "' >> /etc/profile.d/miko-shell-prompt.sh

# Now run the startup script
cat > /tmp/startup.sh << 'MIKO_SCRIPT_EOF'
%s
# Export PATH for interactive shell
export PATH="/go/bin:/usr/local/go/bin:$PATH"
# Start interactive shell
exec /bin/sh --login
MIKO_SCRIPT_EOF

chmod +x /tmp/startup.sh
exec /tmp/startup.sh`,
		version,
		mikoShell.String(),

		startupScript.String())

	// Run the command
	return d.runContainer(cfg, tag, []string{"/bin/sh", "-c", shellCommand}, true)
}

func (d *DockerProvider) ImageExists(tag string) bool {
	cmd := exec.Command("docker", "image", "inspect", tag)
	return cmd.Run() == nil
}

func (d *DockerProvider) RemoveImage(tag string) error {
	cmd := exec.Command("docker", "rmi", "-f", tag)
	return cmd.Run()
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
	// If no startup commands and no scripts are defined, just run the shell
	if len(cfg.Shell.InitHook) == 0 && len(cfg.Shell.Scripts) == 0 {
		return p.RunShell(cfg, tag)
	}

	// 1. Script de startup original
	var startupScript strings.Builder
	startupScript.WriteString("#!/bin/sh\n")
	startupScript.WriteString("set -e\n\n")

	// Agregar comandos de startup
	for _, cmd := range cfg.Shell.InitHook {
		startupScript.WriteString(cmd + "\n\n")
	}

	// 2. Generar wrapper miko-shell
	var mikoShell strings.Builder
	mikoShell.WriteString("#!/bin/sh\n")
	mikoShell.WriteString("set -e\n\n")

	// Configurar PATH para incluir herramientas de Go
	mikoShell.WriteString("# Ensure Go tools are in PATH\n")
	mikoShell.WriteString("export PATH=\"/go/bin:/usr/local/go/bin:$PATH\"\n\n")

	// Función de ayuda
	mikoShell.WriteString("show_help() {\n")
	mikoShell.WriteString("  echo \"Miko Shell - Container development environment\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Usage:\"\n")
	mikoShell.WriteString("  echo \"  miko-shell [command]\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Available Commands:\"\n")
	mikoShell.WriteString("  echo \"  help        Show help for miko-shell\"\n")
	mikoShell.WriteString("  echo \"  list        List available scripts\"\n")
	mikoShell.WriteString("  echo \"  run         Run a script or command inside the container\"\n")
	mikoShell.WriteString("  echo \"  version     Show miko-shell version\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Run 'miko-shell run --help' for information about running scripts\"\n")
	mikoShell.WriteString("}\n\n")

	// Función para listar scripts
	mikoShell.WriteString("list_scripts() {\n")
	mikoShell.WriteString("  echo \"Available scripts:\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	for _, script := range cfg.Shell.Scripts {
		desc := script.Description
		if desc == "" {
			desc = script.Name
		}
		mikoShell.WriteString(fmt.Sprintf("  echo \"  %-15s %s\"\n", script.Name, desc))
	}
	mikoShell.WriteString("}\n\n")

	// Función para ejecutar scripts
	mikoShell.WriteString("run_script() {\n")
	mikoShell.WriteString("  script_name=\"$1\"\n")
	mikoShell.WriteString("  shift\n\n")
	mikoShell.WriteString("  case \"$script_name\" in\n")

	// Agregar case para cada script
	for _, script := range cfg.Shell.Scripts {
		mikoShell.WriteString(fmt.Sprintf("    %s)\n", script.Name))
		mikoShell.WriteString("      # Ejecutar script con argumentos pasados\n")

		// Exportar variables para los argumentos posicionales
		mikoShell.WriteString("      # Establecer argumentos posicionales\n")
		mikoShell.WriteString("      i=1\n")
		mikoShell.WriteString("      for arg in \"$@\"; do\n")
		mikoShell.WriteString("        export \"_MIKO_ARG_${i}=$arg\"\n")
		mikoShell.WriteString("        i=$((i+1))\n")
		mikoShell.WriteString("      done\n\n")

		// Ejecutar cada comando del script, reemplazando $1, $2, etc. con las variables exportadas
		for _, cmd := range script.Commands {
			// Reemplazar $1, $2, etc. con las variables _MIKO_ARG_1, _MIKO_ARG_2, etc.
			processedCmd := cmd
			for i := 1; i <= 9; i++ {
				placeholder := fmt.Sprintf("$%d", i)
				replacement := fmt.Sprintf("${_MIKO_ARG_%d:-}", i)
				processedCmd = strings.ReplaceAll(processedCmd, placeholder, replacement)
			}
			mikoShell.WriteString(fmt.Sprintf("      %s\n", processedCmd))
		}

		// Limpiar las variables de argumentos
		mikoShell.WriteString("\n      # Limpiar variables de argumentos\n")
		mikoShell.WriteString("      for j in $(seq 1 $((i-1))); do\n")
		mikoShell.WriteString("        unset \"_MIKO_ARG_${j}\"\n")
		mikoShell.WriteString("      done\n")

		mikoShell.WriteString("      return $?\n")
		mikoShell.WriteString("      ;;\n")
	}

	// Caso para comando directo (ejecuta el comando pasado directamente)
	mikoShell.WriteString("    --)\n")
	mikoShell.WriteString("      shift\n")
	mikoShell.WriteString("      \"$@\"\n")
	mikoShell.WriteString("      return $?\n")
	mikoShell.WriteString("      ;;\n")

	// Caso para script desconocido
	mikoShell.WriteString("    *)\n")
	mikoShell.WriteString("      echo \"Error: Unknown script '$script_name'\"\n")
	mikoShell.WriteString("      echo \"\"\n")
	mikoShell.WriteString("      list_scripts\n")
	mikoShell.WriteString("      return 1\n")
	mikoShell.WriteString("      ;;\n")
	mikoShell.WriteString("  esac\n")
	mikoShell.WriteString("}\n\n")

	// Función para mostrar ayuda de run
	mikoShell.WriteString("show_run_help() {\n")
	mikoShell.WriteString("  echo \"Run a script or command inside the container\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Usage:\"\n")
	mikoShell.WriteString("  echo \"  miko-shell run <script-name> [args...]  Run a script with optional arguments\"\n")
	mikoShell.WriteString("  echo \"  miko-shell run -- <command> [args...]   Run a direct command\"\n")
	mikoShell.WriteString("  echo \"\"\n")
	mikoShell.WriteString("  echo \"Available scripts:\"\n")
	mikoShell.WriteString("  echo \"\"\n")

	// Listar scripts disponibles
	for _, script := range cfg.Shell.Scripts {
		desc := script.Description
		if desc == "" {
			desc = script.Name
		}
		mikoShell.WriteString(fmt.Sprintf("  echo \"  %-15s %s\"\n", script.Name, desc))
	}
	mikoShell.WriteString("}\n\n")

	// Comando principal
	mikoShell.WriteString("# Detectar versión de la imagen\n")
	mikoShell.WriteString("MIKO_VERSION=\"$(cat /tmp/miko-version 2>/dev/null || echo 'dev')\"\n\n")
	mikoShell.WriteString("# Procesar comandos\n")
	mikoShell.WriteString("case \"$1\" in\n")

	// Comando run
	mikoShell.WriteString("  run)\n")
	mikoShell.WriteString("    shift\n")
	mikoShell.WriteString("    if [ \"$1\" = \"--help\" ] || [ \"$1\" = \"-h\" ]; then\n")
	mikoShell.WriteString("      show_run_help\n")
	mikoShell.WriteString("      exit 0\n")
	mikoShell.WriteString("    fi\n")
	mikoShell.WriteString("    if [ -z \"$1\" ]; then\n")
	mikoShell.WriteString("      echo \"Error: Missing script name or command\"\n")
	mikoShell.WriteString("      echo \"\"\n")
	mikoShell.WriteString("      show_run_help\n")
	mikoShell.WriteString("      exit 1\n")
	mikoShell.WriteString("    fi\n")
	mikoShell.WriteString("    run_script \"$@\"\n")
	mikoShell.WriteString("    exit $?\n")
	mikoShell.WriteString("    ;;\n")

	// Comando open (debe fallar dentro del contenedor)
	mikoShell.WriteString("  open)\n")
	mikoShell.WriteString("    echo \"Error: Already inside a miko-shell container\"\n")
	mikoShell.WriteString("    echo \"The 'open' command can only be used from outside the container\"\n")
	mikoShell.WriteString("    exit 1\n")
	mikoShell.WriteString("    ;;\n")

	// Comando list
	mikoShell.WriteString("  list)\n")
	mikoShell.WriteString("    list_scripts\n")
	mikoShell.WriteString("    exit 0\n")
	mikoShell.WriteString("    ;;\n")

	// Comando version
	mikoShell.WriteString("  version)\n")
	mikoShell.WriteString("    echo \"miko-shell version $MIKO_VERSION\"\n")
	mikoShell.WriteString("    exit 0\n")
	mikoShell.WriteString("    ;;\n")

	// Comando help o sin argumentos
	mikoShell.WriteString("  help|-h|--help|\"\")\n")
	mikoShell.WriteString("    show_help\n")
	mikoShell.WriteString("    exit 0\n")
	mikoShell.WriteString("    ;;\n")

	// Comando desconocido
	mikoShell.WriteString("  *)\n")
	mikoShell.WriteString("    echo \"Error: Unknown command '$1'\"\n")
	mikoShell.WriteString("    echo \"\"\n")
	mikoShell.WriteString("    show_help\n")
	mikoShell.WriteString("    exit 1\n")
	mikoShell.WriteString("    ;;\n")
	mikoShell.WriteString("esac\n")

	// Crear el comando completo que:
	// 1. Guarda la versión en un archivo
	// 2. Crea el script miko-shell
	// 3. Genera el autocompletado
	// 4. Ejecuta el script de startup
	version := "dev"
	if v := os.Getenv("MIKO_VERSION"); v != "" {
		version = v
	}

	shellCommand := fmt.Sprintf(`
# Save version information
echo "%s" > /tmp/miko-version

# Create the miko-shell wrapper
cat > /usr/local/bin/miko-shell << 'MIKO_WRAPPER_EOF'
%s
MIKO_WRAPPER_EOF
chmod +x /usr/local/bin/miko-shell

# Bash completion disabled for sh compatibility
# Bash completion for miko-shell (disabled for sh compatibility)
touch /etc/profile.d/miko-shell-completion.sh

# Setup PATH to include Go tools for all sessions
echo 'export PATH="/go/bin:/usr/local/go/bin:$PATH"' >> /etc/profile.d/miko-shell-path.sh

# Setup prompt to show we're in a miko-shell
echo 'PS1="[\[\e[1;32m\]miko-shell\[\e[0m\]] \w \$ "' >> /etc/profile.d/miko-shell-prompt.sh

# Now run the startup script
cat > /tmp/startup.sh << 'MIKO_SCRIPT_EOF'
%s
# Export PATH for interactive shell
export PATH="/go/bin:/usr/local/go/bin:$PATH"
# Start interactive shell
exec /bin/sh --login
MIKO_SCRIPT_EOF

chmod +x /tmp/startup.sh
exec /tmp/startup.sh`,
		version,
		mikoShell.String(),

		startupScript.String())

	// Run the command
	return p.runContainer(cfg, tag, []string{"/bin/sh", "-c", shellCommand}, true)
}

func (p *PodmanProvider) ImageExists(tag string) bool {
	cmd := exec.Command("podman", "image", "inspect", tag)
	return cmd.Run() == nil
}

func (p *PodmanProvider) RemoveImage(tag string) error {
	cmd := exec.Command("podman", "rmi", "-f", tag)
	return cmd.Run()
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

// ListImages implementation for DockerProvider
func (d *DockerProvider) ListImages() ([]ImageListItem, error) {
	// This is a simplified implementation
	// In a real implementation, you would parse docker images output
	// and filter for miko-shell related images
	return []ImageListItem{}, nil
}

// CleanImages implementation for DockerProvider
func (d *DockerProvider) CleanImages(all bool) ([]string, error) {
	// This is a simplified implementation
	// In a real implementation, you would:
	// 1. List miko-shell images
	// 2. Remove unused ones (or all if all=true)
	// 3. Return list of removed image IDs
	return []string{}, nil
}

// GetImageInfo implementation for DockerProvider
func (d *DockerProvider) GetImageInfo(imageID string) (*ImageInfo, error) {
	// This is a simplified implementation
	// In a real implementation, you would use "docker inspect" to get detailed info
	return &ImageInfo{
		ID:           imageID,
		Tag:          imageID,
		Size:         "Unknown",
		Created:      time.Now(),
		Platform:     "linux/amd64",
		Labels:       make(map[string]string),
		Layers:       []LayerInfo{},
		Env:          []string{},
		ExposedPorts: []string{},
	}, nil
}

// GetPruneInfo implementation for DockerProvider
func (d *DockerProvider) GetPruneInfo() (*PruneInfo, error) {
	// This is a simplified implementation
	// In a real implementation, you would analyze docker system df output
	return &PruneInfo{
		TotalImages:    0,
		UnusedImages:   0,
		DanglingImages: 0,
		BuildCacheSize: "0B",
		TotalSize:      "0B",
	}, nil
}

// PruneImages implementation for DockerProvider
func (d *DockerProvider) PruneImages() (*PruneResult, error) {
	// This is a simplified implementation
	// In a real implementation, you would run "docker system prune"
	return &PruneResult{
		RemovedImages:  0,
		ReclaimedSpace: "0B",
	}, nil
}

// ListImages implementation for PodmanProvider
func (p *PodmanProvider) ListImages() ([]ImageListItem, error) {
	// This is a simplified implementation
	// In a real implementation, you would parse podman images output
	// and filter for miko-shell related images
	return []ImageListItem{}, nil
}

// CleanImages implementation for PodmanProvider
func (p *PodmanProvider) CleanImages(all bool) ([]string, error) {
	// This is a simplified implementation
	// In a real implementation, you would:
	// 1. List miko-shell images
	// 2. Remove unused ones (or all if all=true)
	// 3. Return list of removed image IDs
	return []string{}, nil
}

// GetImageInfo implementation for PodmanProvider
func (p *PodmanProvider) GetImageInfo(imageID string) (*ImageInfo, error) {
	// This is a simplified implementation
	// In a real implementation, you would use "podman inspect" to get detailed info
	return &ImageInfo{
		ID:           imageID,
		Tag:          imageID,
		Size:         "Unknown",
		Created:      time.Now(),
		Platform:     "linux/amd64",
		Labels:       make(map[string]string),
		Layers:       []LayerInfo{},
		Env:          []string{},
		ExposedPorts: []string{},
	}, nil
}

// GetPruneInfo implementation for PodmanProvider
func (p *PodmanProvider) GetPruneInfo() (*PruneInfo, error) {
	// This is a simplified implementation
	// In a real implementation, you would analyze podman system df output
	return &PruneInfo{
		TotalImages:    0,
		UnusedImages:   0,
		DanglingImages: 0,
		BuildCacheSize: "0B",
		TotalSize:      "0B",
	}, nil
}

// PruneImages implementation for PodmanProvider
func (p *PodmanProvider) PruneImages() (*PruneResult, error) {
	// This is a simplified implementation
	// In a real implementation, you would run "podman system prune"
	return &PruneResult{
		RemovedImages:  0,
		ReclaimedSpace: "0B",
	}, nil
}
