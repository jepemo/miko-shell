# Miko Shell 🐚

A containerized development environment manager that allows you to work with different tools and dependencies without installing them locally.

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
   miko-shell init
   ```

2. **Edit the configuration** (`dev-config.yaml`):

   ```yaml
   name: my-project
   container-provider: docker
   image: alpine:latest
   pre-install:
     - apk add curl git
   shell:
     init-hook:
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

- **Python**: `examples/dev-config-python.example.yaml`
- **JavaScript/Node.js**: `examples/dev-config-javascript.example.yaml`
- **Go**: `examples/dev-config-go.example.yaml`
- **Rust**: `examples/dev-config-rust.example.yaml`
- **Elixir**: `examples/dev-config-elixir.example.yaml`
- **Elixir/Phoenix**: `examples/dev-config-phoenix.example.yaml`
- **PHP**: `examples/dev-config-php.example.yaml`
- **Ruby**: `examples/dev-config-ruby.example.yaml`
- **Ruby/Rails**: `examples/dev-config-rails.example.yaml`
- **Java**: `examples/dev-config-java.example.yaml`

Copy and customize any example for your project:

```bash
cp examples/dev-config-javascript.example.yaml dev-config.yaml
# Edit dev-config.yaml to match your project needs
```

## Configuration

The `dev-config.yaml` file structure:

```yaml
# Project name (used for container image naming)
name: my-project

# Container provider to use (docker or podman)
container-provider: docker

# Base image for the container
image: alpine:latest

# Commands to run during image build
pre-install:
  - apk add yamllint

# Shell configuration
shell:
  # Commands to run before any shell/command execution
  init-hook:
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
- **container-provider**: `docker` (default) or `podman`
- **image**: Base container image to use
- **pre-install**: List of commands to run during image build
- **shell.init-hook**: Commands to run before any execution
- **shell.scripts**: Named scripts that can be executed with `miko-shell run <name>`

## How it Works

1. **Configuration**: The tool reads `dev-config.yaml` to understand project setup
2. **Project Naming**: Uses the `name` field (auto-generated from directory name) for container image naming
3. **Image Building**: Creates a container image with specified base image and dependencies
4. **Image Tagging**: Images are tagged with the project name and a hash of the configuration file
5. **Volume Mounting**: Project directory is mounted as `/workspace` in the container
6. **Script Execution**: Scripts can be run by name or commands executed directly

## Examples

### Python Project

```yaml
name: python-app
container-provider: docker
image: python:3.9-slim
pre-install:
  - pip install flask pytest
shell:
  init-hook:
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
container-provider: docker
image: node:18-alpine
pre-install:
  - npm install -g pnpm
shell:
  init-hook:
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
