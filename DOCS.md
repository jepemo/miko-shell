# Miko Shell Documentation

Comprehensive user and operator guide for the miko-shell CLI. This document is the canonical reference; for a quick overview, see the project README.

## 1. Introduction

`miko-shell` packages your project into a lightweight container sExit codes: infrastructure errors (e.g., config invalid, engine missing) are returned with explanatory messages; script command failures propagate the command's exit code without extra help output.

### 1.1 Principles

- Reproducible: deterministic image tag from config hash
- Minimal host impact: no global installs; everything inside the container
- Familiar: plain YAML; Docker or Podman under the hood
- Fast feedback: caches layers and keeps startup simple

### 1.2 Core workflow

```
miko-shell.yaml  ->  build image  ->  run scripts / ad‑hoc commands
```

## 2. Installation

### 2.1 Quick Install (Recommended)

Install via script:

```bash
curl -sSL https://raw.githubusercontent.com/jepemo/miko-shell/main/install.sh | bash
```

Options:

- Pin a version:

  ```bash
  curl -sSL https://raw.githubusercontent.com/jepemo/miko-shell/main/install.sh | bash -s -- --version v1.0.0
  ```

- Uninstall completely:

  ```bash
  curl -sSL https://raw.githubusercontent.com/jepemo/miko-shell/main/install.sh | bash -s -- --uninstall
  ```

- Custom install directory:

  ```bash
  export BIN_DIR="$HOME/bin"
  curl -sSL https://raw.githubusercontent.com/jepemo/miko-shell/main/install.sh | bash
  ```

The script detects your OS/arch, fetches the release asset, and falls back to building from source if no asset matches.

### 2.2 From source (Go >= 1.23)

```bash
make build
# or
go build -o miko-shell .
```

### 2.3 Bootstrap script (No Go required)

For systems without Go installed, use the bootstrap script that downloads Go and builds the project:

```bash
# Build the project (downloads Go if needed)
./bootstrap.sh

# Show help and all available options
./bootstrap.sh --help
```

#### Bootstrap Options

The bootstrap script supports several useful options:

- **`./bootstrap.sh`** — Default behavior: build the project (downloads Go 1.23.4 if needed)

- **`./bootstrap.sh --clean`** — Clean build artifacts, similar to `make clean` but also removes:

  - `miko-shell` binary
  - `miko-shell-host` binary
  - `build/` directory
  - `.bootstrap/` directory

- **`./bootstrap.sh --clean-images`** — Remove all Docker/Podman images starting with 'miko-shell':

  - `miko-shell:*` (any tag)
  - `miko-shell-dev:*`
  - `miko-shell-test:*`
  - Any image beginning with `miko-shell`

- **`./bootstrap.sh --help`** — Show detailed usage information

#### Examples

```bash
# Clean everything and rebuild
./bootstrap.sh --clean
./bootstrap.sh

# Clean only container images
./bootstrap.sh --clean-images

# Full cleanup (build artifacts + images)
./bootstrap.sh --clean
./bootstrap.sh --clean-images
./bootstrap.sh  # rebuild
```

The bootstrap script is particularly useful for:

- CI environments without Go pre-installed
- Development on new machines
- Consistent builds across different environments

### 2.4 Prebuilt binaries

Download from Releases and place on your PATH as `miko-shell`.

### 2.5 Verify

```bash
miko-shell --help
miko-shell version
```

## 3. Quick Start

```bash
# 1) Scaffold a config
miko-shell init               # or: miko-shell init --dockerfile

# 2) Build the image (optional – auto-build on first run)
miko-shell build

# 3) List available scripts
miko-shell run

# 4) Run a script
miko-shell run test

# 5) Run an ad‑hoc command (everything after -- goes verbatim)
miko-shell run -- go env
```

Minimal `miko-shell.yaml`:

```yaml
name: my-project
container:
  provider: docker
  image: alpine:latest
  setup:
    - apk add --no-cache bash curl git
shell:
  startup:
    - echo "Welcome to my-project shell"
  scripts:
    - name: test
      commands:
        - go test ./...
    - name: greet
      commands:
        - echo "Hello $1, you are $2 years old"
```

## 4. Concepts and Architecture

### 4.1 Configuration model

Top-level keys:

- `name` — project label; also used for image tagging
- `container` — how to build or select the base image
- `shell` — what to run on startup and named scripts to expose

Container section:

- `provider`: `docker` (default) or `podman`
- `image`: base image to use if you’re not building
- `build` (optional): custom image build
  - `dockerfile`: path to Dockerfile
  - `context`: build context (default: ".")
  - `args`: map of build-args
- `setup`: list of commands executed at image build time (install deps)

Shell section:

- `startup`: commands executed on every `run`
- `scripts[]`:
  - `name`: script name to call via `miko-shell run <name>`
  - `description` (optional)
  - `commands[]`: commands executed inside the container. Positional `$1`, `$2`, … map to arguments.

### 4.2 Image caching and tagging

`miko-shell` computes a short hash of your config and tags the built image as:

```
<normalized-name>:<config-hash>
```

When the config changes, a new tag is built; otherwise the existing image is reused.

### 4.3 Runtime environment

- The repository is mounted at `/workspace`
- The working directory is `/workspace`
- Host details are available to scripts when needed (for example via environment variables if provided by the wrapper). Typical variables:
  - `MIKO_HOST_OS`, `MIKO_HOST_ARCH` (when supported)

### 4.4 Docker and Podman

Choose your engine via `container.provider`. Everything else works the same.

## 5. Command Reference

Global flags:

- `-c, --config`: path to config (default: `miko-shell.yaml`)

### 5.1 init

Scaffold a new config.

```bash
miko-shell init           # prebuilt base image + setup commands
miko-shell init --dockerfile  # Dockerfile-driven build
```

### 5.2 build

Build the image defined by the config. Usually optional — first `run` will build as needed.

```bash
miko-shell build
miko-shell build -c examples/dev-config-go.example.yaml
miko-shell build --force  # Force rebuild by removing existing image first
```

Flags:

- `--force, -f`: Remove existing image with same tag and rebuild from scratch
- `--config, -c`: Path to configuration file

### 5.3 run

Run a named script or an ad‑hoc command inside the container.

```bash
# List available scripts (no args)
miko-shell run

# Run a named script
miko-shell run test
miko-shell run greet Alice 42

# Ad‑hoc command (everything after -- is passed verbatim)
miko-shell run -- go env
```

Exit codes: infrastructure errors (e.g., config invalid, engine missing) are returned with explanatory messages; script command failures propagate the command’s exit code without extra help output.

### 5.4 open

Open an interactive shell inside the development environment.

```bash
miko-shell open
miko-shell open -c examples/dev-config-go.example.yaml
```

This provides direct access to the containerized environment for debugging, exploration, or manual operations.

### 5.5 image

Comprehensive container image management with multiple subcommands.

```bash
# Build container image (alternative to miko-shell build)
miko-shell image build
miko-shell image build --force  # Force rebuild

# List miko-shell images
miko-shell image list
miko-shell image ls              # Alias

# Clean unused images
miko-shell image clean
miko-shell image clean --all     # Remove all miko-shell images

# Show detailed image information
miko-shell image info            # Current project's image
miko-shell image info <image-id> # Specific image

# Prune all unused images and build cache
miko-shell image prune
miko-shell image prune --force   # Skip confirmation
```

The `image` command provides a modern, Docker-like interface for managing container images:

- **`build`**: Same functionality as `miko-shell build` with improved UX
- **`list`**: View all miko-shell related images with metadata
- **`clean`**: Remove unused images to reclaim disk space
- **`info`**: Inspect image details, layers, and configuration
- **`prune`**: System-wide cleanup of unused images and build cache

### 5.6 version

Show version information.

```bash
miko-shell version
```

### 5.7 completion

Generate shell autocompletion scripts for enhanced command-line experience.

```bash
# Generate completion for bash
miko-shell completion bash

# Generate completion for zsh
miko-shell completion zsh

# Generate completion for fish
miko-shell completion fish

# Generate completion for PowerShell
miko-shell completion powershell
```

To enable autocompletion, follow the instructions provided by the command output for your specific shell.eveloper and CI job uses the same toolchain. It builds or reuses a container image from a simple YAML file, mounts your repo at `/workspace`, and runs named scripts or ad‑hoc commands.

### 5.8 version

Show version information.

```bash
miko-shell version
```

## 6. Examples Library

The `examples/` directory includes ready‑to‑use configs for:

- Go, Python, Node/Next.js, Rust
- Ruby/Rails, PHP/Laravel
- Elixir/Phoenix, Django, Java/Spring Boot

Use them directly with `-c` or copy to your project as a starting point. See `examples/README.md` and `examples/USAGE.md`.

## 7. Configuration Patterns

### 7.1 Prebuilt base image + setup

```yaml
container:
  provider: docker
  image: golang:1.23-alpine
  setup:
    - apk add --no-cache make git
```

### 7.2 Custom Dockerfile build

```yaml
container:
  provider: docker
  build:
    dockerfile: Dockerfile
    context: .
    args:
      NODE_VERSION: "20"
  setup:
    - npm i -g pnpm
```

### 7.3 Scripts and arguments

```yaml
shell:
  startup:
    - echo "Starting dev environment"
  scripts:
    - name: dev
      description: Run dev server
      commands:
        - npm run dev
    - name: greet
      commands:
        - echo "Hello $1"
```

Call as `miko-shell run greet Alice` -> prints `Hello Alice`.

## 8. CI/CD usage

Run the same scripts in CI without installing language toolchains on runners.

GitHub Actions (minimal):

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build image
        run: |
          miko-shell build
      - name: Test
        run: |
          miko-shell run test
```

GitLab CI:

```yaml
stages: [build, test]

build:
  stage: build
  image: docker:stable
  services: [docker:dind]
  script:
    - miko-shell build

test:
  stage: test
  image: docker:stable
  services: [docker:dind]
  script:
    - miko-shell run test
```

## 9. Troubleshooting

| Symptom                     | Likely cause                        | Fix                                                        |
| --------------------------- | ----------------------------------- | ---------------------------------------------------------- |
| `miko-shell.yaml not found` | Missing config                      | Run `miko-shell init` or pass `-c`                         |
| `invalid provider`          | Typo in `container.provider`        | Use `docker` or `podman`                                   |
| Engine not found            | Docker/Podman not installed/running | Install and start your engine                              |
| Script not listed           | Name mismatch                       | Run `miko-shell run` to list; check `shell.scripts[].name` |
| Command exits with non‑zero | Command failed inside container     | Fix the underlying command; exit code is preserved         |
| Too many cached images      | Multiple miko-shell builds          | Use `miko-shell image clean` or `miko-shell image prune`   |
| Disk space issues           | Build artifacts accumulation        | Use `miko-shell image prune` for complete cleanup          |

### 9.1 Maintenance Commands

For ongoing maintenance and cleanup:

```bash
# Clean build artifacts (binaries, temp directories)
./bootstrap.sh --clean

# Clean all miko-shell container images
./bootstrap.sh --clean-images

# Full cleanup and rebuild
./bootstrap.sh --clean
./bootstrap.sh --clean-images
./bootstrap.sh

# Using miko-shell build with force rebuild
miko-shell build --force  # Removes existing image and rebuilds

# Modern image management commands
miko-shell image list     # See all miko-shell images
miko-shell image clean    # Remove unused images
miko-shell image clean --all  # Remove all miko-shell images
miko-shell image prune    # System-wide cleanup with confirmation
miko-shell image prune --force  # System-wide cleanup without confirmation
miko-shell image info     # Inspect current project's image
```

miko-shell build --force # Removes existing image and rebuilds

```

## 10. FAQ

Q: Docker or Podman?

A: Both are supported — set `container.provider` accordingly.

Q: How do I pass arguments to scripts?

A: Positional arguments are available as `$1`, `$2`, … inside each command in `commands[]`.

Q: Where does it run?

A: All commands run inside the container with your project mounted at `/workspace`.

Q: Do I need to run `build` first?

A: Not strictly — the first `run`/`open` will build if needed. Running `build` proactively helps surface build errors early.

Q: How do I get an interactive shell?

A: Use `miko-shell open` to get a shell inside the development environment.

Q: What's the difference between `build` and `image build`?

A: Both do the same thing. `image build` is the modern interface with additional flags and better UX. Use `image` commands for comprehensive image management.

## 11. License

MIT. See `LICENSE`.

## 12. Links

- Examples: `examples/`
- README (overview): `README.md`
```
