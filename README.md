# Miko Shell 🐚

A containerized development environment manager that allows you to work with different tools and dependencies without i### Configuration Options

- **name**: Project name used for container image naming (auto-generated from directory name)
- **container.provider**: `docker` (default) or `podman`
- **container.image**: Base container image to use
- **container.setup**: List of commands to run during image build
- **container.build**: Custom Dockerfile build configuration (alternative to image)
  - **dockerfile**: Path to custom Dockerfile
  - **context**: Build context directory (optional)
  - **args**: Build arguments (optional)
- **shell.startup**: Commands to run before any execution
- **shell.scripts**: Named scripts that can be executed with `miko-shell run <name>`g them locally.

[![CI](https://github.com/jepemo/miko-shell/actions/workflows/ci.yml/badge.svg)](https://github.com/jepemo/miko-shell/actions/workflows/ci.yml)
[![Release](https://github.com/jepemo/miko-shell/actions/workflows/release.yml/badge.svg)](https://github.com/jepemo/miko-shell/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jepemo/miko-shell)](https://goreportcard.com/report/github.com/jepemo/miko-shell)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- 🐳 **Container-based**: Uses Docker or Podman for isolated environments
- 🔧 **Configurable**: YAML-based configuration for easy customization
- 🚀 **Fast**: Efficient image building and caching
- 📦 **Portable**: Works on Linux, macOS, and Windows
- 🔄 **Scripts**: Define custom scripts for common tasks
- 🛠️ **Flexible**: Support for multiple container providers

## Quick Start

### Option 1: Using Bootstrap Script (Recommended for clean environments)

The bootstrap script automatically downloads Go and builds the project:

```bash
./bootstrap.sh
```

This will:

- Download Go 1.23.4 (if not already available with the right version)
- Download dependencies
- Build the project
- Test the binary

### Option 2: Using Go (if you have Go installed)

```bash
# Build the project
make build

# Or directly with go
go build -o miko-shell .
```

## Installation

### Download from Releases

Download the latest release from the [releases page](https://github.com/jepemo/miko-shell/releases).

### Build from Source

```bash
git clone https://github.com/jepemo/miko-shell.git
cd miko-shell
make build
```

### Using Go

```bash
go install github.com/jepemo/miko-shell@latest
```

## Quick Start

1. **Initialize a new project**:

   ```bash
   # Basic setup with pre-built image
   miko-shell init

   # Advanced setup with custom Dockerfile
   miko-shell init --dockerfile
   ```

2. **Edit the configuration** (`miko-shell.yaml`):

   ```yaml
   name: my-project
   container:
     provider: docker
     image: alpine:latest
     setup:
       - apk add curl git
   shell:
     startup:
       - echo "Welcome to my development environment!"
     scripts:
       - name: test
         commands:
           - echo "Running tests..."
           - go test ./...
   ```

3. **Run commands**:

   ```bash
   miko-shell run echo "Hello from container!"
   miko-shell run test  # Run custom script
   ```

4. **Open interactive shell**:
   ```bash
   miko-shell shell
   ```

## Configuration Examples

Pre-configured examples for different languages are available in the `examples/` directory:

- **Python**: `examples/miko-shell-python.example.yaml`
- **JavaScript/Node.js**: `examples/miko-shell-javascript.example.yaml`
- **Go**: `examples/miko-shell-go.example.yaml`
- **Rust**: `examples/miko-shell-rust.example.yaml`
- **Elixir**: `examples/miko-shell-elixir.example.yaml`
- **Elixir/Phoenix**: `examples/miko-shell-phoenix.example.yaml`
- **PHP**: `examples/miko-shell-php.example.yaml`
- **Ruby**: `examples/miko-shell-ruby.example.yaml`
- **Ruby/Rails**: `examples/miko-shell-rails.example.yaml`
- **Java**: `examples/miko-shell-java.example.yaml`

Copy and customize any example for your project:

```bash
cp examples/miko-shell-javascript.example.yaml miko-shell.yaml
# Edit miko-shell.yaml to match your project needs
```

## Initialization Options

### Pre-built Image Mode (Default)

```bash
miko-shell init
```

Creates a configuration that uses a pre-built container image (Alpine Linux) with package installation commands in the `setup` section:

```yaml
name: my-project
container:
  provider: docker
  image: alpine:latest
  setup:
    - apk add --no-cache curl git
```

### Custom Dockerfile Mode

```bash
miko-shell init --dockerfile  # or -d
```

Creates a configuration that uses a custom Dockerfile with build arguments, plus generates a sample Dockerfile:

```yaml
name: my-project
container:
  provider: docker
  build:
    dockerfile: ./Dockerfile
    context: .
    args:
      VERSION: "1.0"
```

This mode also creates a `Dockerfile` with basic Alpine setup that you can customize for your specific needs.

## Configuration

The `miko-shell.yaml` file structure:

```yaml
# Project name (used for container image naming)
name: my-project

# Container configuration
container:
  # Container provider to use (docker or podman)
  provider: docker

  # Base image for the container
  image: alpine:latest

  # Commands to run during image build
  setup:
    - apk add yamllint

# Shell configuration
shell:
  # Commands to run before any shell/command execution
  startup:
    - printf "Hello"

  # Custom scripts that can be executed with 'miko-shell run <script-name>'
  scripts:
    - name: lint
      commands:
        - yamllint example.yaml
    - name: test
      commands:
        - echo "Running tests..."
        - echo "All tests passed!"
```

### Configuration Options

- **name**: Project name used for container image naming (auto-generated from directory name)
- **provider**: `docker` (default) or `podman`
- **image**: Base container image to use
- **setup**: List of commands to run during image build
- **shell.startup**: Commands to run before any execution
- **shell.scripts**: Named scripts that can be executed with `miko-shell run <name>`

## How it Works

1. **Configuration**: The tool reads `miko-shell.yaml` to understand project setup
2. **Project Naming**: Uses the `name` field (auto-generated from directory name) for container image naming
3. **Image Building**: Creates a container image with specified base image and dependencies
4. **Image Tagging**: Images are tagged with the project name and a hash of the configuration file
5. **Volume Mounting**: Project directory is mounted as `/workspace` in the container
6. **Script Execution**: Scripts can be run by name or commands executed directly

## Examples

### Python Project

```yaml
name: python-app
container:
  provider: docker
  image: python:3.9-slim
  setup:
    - pip install flask pytest
shell:
  startup:
    - echo "Python environment ready"
  scripts:
    - name: test
      commands:
        - python -m pytest tests/
    - name: run
      commands:
        - python app.py
```

### Node.js Project

```yaml
name: nodejs-app
container:
  provider: docker
  image: node:18-alpine
  setup:
    - npm install -g pnpm
shell:
  startup:
    - echo "Node.js environment ready"
  scripts:
    - name: install
      commands:
        - pnpm install
    - name: test
      commands:
        - pnpm test
    - name: dev
      commands:
        - pnpm dev
```

### Custom Dockerfile Build

You can also use custom Dockerfiles instead of pre-built images:

```yaml
name: my-custom-app
container:
  provider: docker
  build:
    dockerfile: ./Dockerfile
    context: .
    args:
      NODE_VERSION: "18"
      ENVIRONMENT: "development"
shell:
  startup:
    - echo "Custom environment ready"
  scripts:
    - name: build
      commands:
        - npm run build
```

This approach gives you complete control over the container environment while still benefiting from miko-shell's script management and development workflow.

## Requirements

- Docker or Podman installed and accessible
- Go 1.20+ for building from source

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is open source and available under the MIT License.
