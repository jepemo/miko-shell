# Miko-Shell CLI Tool Plan

## Overview

`miko-shell` is a CLI tool designed to:

- Abstract dependencies used in local development projects using containers
- Create container images based on configuration for use in pipelines
- Connect to containers to execute scripts in the project context
- Provide a consistent development environment across different machines and projects

## Core Features

✅ **Implemented Features:**

- Container-based development environments (Docker/Podman support)
- YAML-based configuration with validation
- Script execution with positional parameter support
- Direct command execution with flag separator (`--`)
- Interactive shell access
- Custom Dockerfile support
- Automatic image building and caching
- Pre-configured examples for multiple languages

## Commands

### `init` - Initialize Configuration

Creates a `miko-shell.yaml` configuration file in the current directory.

**Default mode (pre-built image):**

```bash
miko-shell init
```

**Advanced mode (custom Dockerfile):**

```bash
miko-shell init --dockerfile  # or -d
```

✅ **Status**: Fully implemented with validation and Dockerfile generation.

### `build` - Build Container Image

Generates a container image from the configuration.

```bash
miko-shell build
```

✅ **Status**: Fully implemented with caching and error handling.

### `run` - Execute Commands

Executes commands or scripts inside the container.

**Script execution with parameters:**

```bash
miko-shell run script-name [args...]
```

**Direct command execution:**

```bash
miko-shell run command [args...]
```

**Commands with flags (using separator):**

```bash
miko-shell run -- command -flags args
```

✅ **Status**: Fully implemented with:

- Script parameter support ($1, $2, etc.)
- Direct command execution
- Flag separator (`--`) support
- Automatic image building if needed

### `open` - Interactive Shell

Opens an interactive shell session inside the container.

```bash
miko-shell open
```

✅ **Status**: Fully implemented with startup commands support.

### Additional Commands

- `help` - Show usage information ✅
- `version` - Show current version ✅

## Configuration Format

The `miko-shell.yaml` file structure:

**Default mode (pre-built image):**

```yaml
name: my-project
container:
  provider: docker
  image: alpine:latest
  setup:
    - apk add yamllint
shell:
  startup:
    - printf "Hello"
  scripts:
    - name: lint
      commands:
        - yamllint -c yamllint.config .
    - name: hello
      commands:
        - echo "Hello $1" # Script with parameter support
```

**Advanced mode (custom Dockerfile):**

```yaml
name: my-project
container:
  provider: docker
  build:
    dockerfile: ./Dockerfile
    context: .
    args:
      VERSION: "1.0"
shell:
  startup:
    - printf "Hello"
  scripts:
    - name: lint
      commands:
        - yamllint -c yamllint.config .
```

### Configuration Options

- **name**: Project name used for container image naming (auto-generated from directory name, normalized)
- **container**: Container configuration section
  - **provider**: Container provider to use (`docker` or `podman`). Default: `docker`
  - **image**: Base image to use in the project
  - **setup**: List of commands to install packages during image build
  - **build**: Custom Dockerfile build configuration (alternative to `image`)
    - **dockerfile**: Path to custom Dockerfile
    - **context**: Build context directory (optional)
    - **args**: Build arguments (optional)
- **shell**: Shell configuration options
  - **startup**: List of commands to execute before any shell/command execution
  - **scripts**: List of scripts available for execution in the container context
    - **name**: Script name
    - **commands**: List of commands to execute in that script
    - **✅ NEW**: Scripts support positional arguments (`$1`, `$2`, etc.)

## How It Works

### Image Building Process

1. **Configuration Reading**: Tool reads `miko-shell.yaml` to understand project setup
2. **Project Naming**: Uses the `name` field (auto-generated from directory name) for container image naming
3. **Image Building**: Creates a container image with specified base image and dependencies
4. **Image Tagging**: Images are tagged with project name and hash of configuration file
5. **Caching**: Uses Docker/Podman layer caching for efficient rebuilds

### Command Execution Process

1. **Image Verification**: Ensures container image exists (builds if necessary)
2. **Volume Mounting**: Project directory is mounted as `/workspace` in the container
3. **Working Directory**: Sets `/workspace` as working directory inside container
4. **Startup Commands**: Executes commands specified in `shell.startup`
5. **Command Resolution**:
   - If command matches a script name → executes script with parameter support
   - Otherwise → executes command directly in container
6. **Cleanup**: Container is automatically removed after execution (`--rm` flag)

### Script Parameter Handling

- Arguments passed after script name are available as `$1`, `$2`, `$3`, etc.
- Uses shell `set --` command to properly set positional parameters
- Supports multiple commands within a single script
- Compatible with both string and array command formats

### Direct Command Execution

- Commands can be executed directly without defining scripts
- Use `--` separator for commands with flags to avoid conflicts
- Full access to container environment and mounted workspace

## Usage Examples

### Script Parameters

```yaml
shell:
  scripts:
    - name: hello
      commands:
        - echo "Hello $1"
    - name: greet
      commands:
        - echo "Hello $1, you are $2 years old"
    - name: deploy
      commands:
        - echo "Deploying to $1 environment"
        - echo "Using version $2"
```

Usage:

```bash
miko-shell run hello world                    # Output: "Hello world"
miko-shell run greet John 25                  # Output: "Hello John, you are 25 years old"
miko-shell run deploy production v1.2.3       # Output: "Deploying to production environment"
                                               #         "Using version v1.2.3"
```

### Direct Commands

```bash
# Simple commands
miko-shell run echo "Hello World"
miko-shell run ls -l

# Commands with flags (use -- separator)
miko-shell run -- ls -la
miko-shell run -- grep -r "pattern" .
miko-shell run -- curl -X POST https://api.example.com
miko-shell run -- docker --version

# Complex commands
miko-shell run -- "ls -la | grep .yaml > files.txt"
```

### Language-Specific Examples

**Python Project:**

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
    - name: install
      commands:
        - pip install $1
```

**Node.js Project:**

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
    - name: build
      commands:
        - pnpm build --target $1 # Parameter support for build targets
```

## System Requirements

- **Container Runtime**: Docker or Podman installed and accessible
- **Build Requirements**: Go 1.20+ for building from source
- **Operating System**: Linux, macOS, or Windows
- **Architecture**: Support for AMD64 and ARM64

## Implementation Status

### ✅ Completed Features

- [x] Project initialization with templates
- [x] Container image building with caching
- [x] Script execution with positional parameter support
- [x] Direct command execution with flag separator (`--`)
- [x] Interactive shell access
- [x] Docker and Podman provider support
- [x] Custom Dockerfile support
- [x] Configuration validation
- [x] Error handling and user feedback
- [x] Pre-configured language examples
- [x] Automatic cleanup of containers
- [x] Volume mounting and workspace setup
- [x] Startup command execution
- [x] Help and version commands

### Implementation Details

**Core Libraries:**

- `cobra` - CLI command handling ✅
- `yaml.v3` - YAML configuration parsing ✅
- `golang.org/x/text` - Text normalization for project names ✅

**Key Features:**

- SHA256 hashing for configuration-based image tagging ✅
- Automatic container cleanup with `--rm` flag ✅
- Proper error handling with descriptive messages ✅
- Container provider detection and validation ✅
- Support for both string and array command formats in scripts ✅
- Argument escaping and shell safety ✅

**Testing:**

- Unit tests for core functionality ✅
- Mock container providers for testing ✅
- Configuration validation tests ✅
- Error handling verification ✅

## Available Examples

Pre-configured examples in `examples/` directory:

- Python (`miko-shell-python.example.yaml`)
- JavaScript/Node.js (`miko-shell-javascript.example.yaml`)
- Go (`miko-shell-go.example.yaml`)
- Rust (`miko-shell-rust.example.yaml`)
- Elixir (`miko-shell-elixir.example.yaml`)
- Phoenix (`miko-shell-phoenix.example.yaml`)
- PHP (`miko-shell-php.example.yaml`)
- Ruby (`miko-shell-ruby.example.yaml`)
- Rails (`miko-shell-rails.example.yaml`)
- Java (`miko-shell-java.example.yaml`)
- Laravel (`miko-shell-laravel.example.yaml`)
- Next.js (`miko-shell-nextjs.example.yaml`)
- Spring Boot (`miko-shell-spring-boot.example.yaml`)
- Django (`miko-shell-django.example.yaml`)

## Usage Workflow

1. **Initialize**: `miko-shell init` (or `miko-shell init --dockerfile`)
2. **Configure**: Edit `miko-shell.yaml` to match project needs
3. **Execute**:
   - Run scripts: `miko-shell run script-name [args]`
   - Run commands: `miko-shell run command` or `miko-shell run -- command -flags`
   - Interactive shell: `miko-shell shell`
4. **Build** (optional): `miko-shell build` for explicit image building

The tool automatically handles image building, caching, and cleanup, providing a seamless containerized development experience.
