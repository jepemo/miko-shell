# Miko Shell - Usage Guide

## 🎯 Quick Start

### Using Default Configuration

```bash
# Create a new miko-shell.yaml file
./miko-shell init

# Build the container image
./miko-shell build

# Run a command in the container
./miko-shell run python --version

# Open an interactive shell
./miko-shell shell
```

### Using Custom Configuration Files

```bash
# Build with a specific configuration file
./miko-shell build -c examples/miko-shell-python.example.yaml

# Run commands with custom configuration
./miko-shell run -c examples/miko-shell-nextjs.example.yaml npm install
./miko-shell run -c examples/miko-shell-go.example.yaml go build

# Open shell with custom configuration
./miko-shell shell -c examples/miko-shell-django.example.yaml
```

## 📋 Available Commands

### Core Commands

- `init` - Initialize a new project with miko-shell.yaml
- `build` - Build container image from configuration
- `run` - Run commands inside the container (run without args to see available scripts)
- `shell` - Open interactive shell session

### Script Discovery

```bash
# List all available scripts with descriptions
./miko-shell run

# List scripts from specific configuration file
./miko-shell run -c examples/miko-shell-python.example.yaml
```

### Common Usage Patterns

#### Python Development

```bash
# Use Python example configuration
./miko-shell build -c examples/miko-shell-python.example.yaml
./miko-shell run -c examples/miko-shell-python.example.yaml test
./miko-shell run -c examples/miko-shell-python.example.yaml install
```

#### JavaScript/Node.js Development

```bash
# Use JavaScript example configuration
./miko-shell build -c examples/miko-shell-javascript.example.yaml
./miko-shell run -c examples/miko-shell-javascript.example.yaml npm install
./miko-shell run -c examples/miko-shell-javascript.example.yaml test
```

#### Go Development

```bash
# Use Go example configuration
./miko-shell build -c examples/miko-shell-go.example.yaml
./miko-shell run -c examples/miko-shell-go.example.yaml build
./miko-shell run -c examples/miko-shell-go.example.yaml test
```

#### Framework-Specific Examples

```bash
# Django development
./miko-shell build -c examples/miko-shell-django.example.yaml
./miko-shell run -c examples/miko-shell-django.example.yaml runserver

# Next.js development
./miko-shell build -c examples/miko-shell-nextjs.example.yaml
./miko-shell run -c examples/miko-shell-nextjs.example.yaml dev

# Spring Boot development
./miko-shell build -c examples/miko-shell-spring-boot.example.yaml
./miko-shell run -c examples/miko-shell-spring-boot.example.yaml bootrun
```

## 🔧 Configuration Options

### Command Line Flags

- `-c, --config` - Path to configuration file (default: miko-shell.yaml)
- `-h, --help` - Show help information

### Environment Variables

- `MIKO_SHELL_CONFIG` - Default configuration file path
- `DOCKER_BUILDKIT` - Enable Docker BuildKit (recommended)

## 🚀 Interactive Demo

Run the interactive demo to try different language configurations:

```bash
cd examples
./demo.sh
```

The demo provides 16 different language and framework configurations to explore.

## 📝 Configuration File Format

All configuration files follow the same YAML structure:

```yaml
name: "project-name"
container:
  provider: "docker" # or "podman"
  image: "base-image:tag"
  setup:
    - "setup command 1"
    - "setup command 2"
shell:
  startup:
    - "startup command"
  scripts:
    - name: "script-name"
      description: "What this script does (optional)"
      commands:
        - "command 1"
        - "command 2"
```

## 🌟 Best Practices

1. **Use specific configuration files** for different project types
2. **Keep configurations in version control** alongside your projects
3. **Use the demo script** to explore new language setups
4. **Test configurations** before committing to ensure they work correctly
5. **Document custom scripts** in your configuration files

## 🔍 Troubleshooting

### Common Issues

1. **Configuration file not found**

   ```bash
   # Check if file exists
   ls -la examples/miko-shell-python.example.yaml

   # Use absolute path if needed
   ./miko-shell build -c /full/path/to/config.yaml
   ```

2. **Container provider not available**

   ```bash
   # Check if Docker is running
   docker info

   # Or use Podman
   ./miko-shell build -c examples/miko-shell-python.example.yaml
   ```

3. **Permission errors**
   ```bash
   # Make sure miko-shell is executable
   chmod +x miko-shell
   ```

## 💡 Tips

- Use tab completion for configuration file paths
- Keep configuration files organized in subdirectories
- Use meaningful names for custom configuration files
- Test new configurations in isolation before using in production
- Consider using environment-specific configuration files (dev, staging, prod)
