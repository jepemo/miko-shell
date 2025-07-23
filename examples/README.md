# Miko Shell Configuration Examples

This directory contains example configuration files for different programming languages and development environments.

## 🎯 Quick Start

You can use any of these examples with the `-c/--config` flag:

```bash
# Use a specific configuration file
./miko-shell build -c examples/miko-shell-python.example.yaml
./miko-shell run -c examples/miko-shell-nextjs.example.yaml
./miko-shell shell -c examples/miko-shell-go.example.yaml

# Default behavior (looks for miko-shell.yaml in current directory)
./miko-shell build
./miko-shell run
./miko-shell shell
```

## 📁 Available Examples (14 configurations)

### 🐍 Python - `miko-shell-python.example.yaml`

- Base image: `python:3.9-slim`
- Pre-installed packages: pytest, flask, requests
- Scripts: test, install, lint, shell

### 🟨 JavaScript/Node.js - `miko-shell-javascript.example.yaml`

- Base image: `node:18-alpine`
- Pre-installed tools: npm, yarn, eslint, prettier
- Scripts: test, install, lint, format, build, dev, start, clean

### 🔵 Go - `miko-shell-go.example.yaml`

- Base image: `golang:1.22-alpine`
- Pre-installed tools: golangci-lint, goimports
- Scripts: test, build, run, lint, format, mod, coverage, clean

### 🦀 Rust - `miko-shell-rust.example.yaml`

- Base image: `rust:1.75-alpine`
- Pre-installed tools: clippy, rustfmt
- Scripts: test, build, release, run, lint, format, check, update, clean, doc

### 🔮 Elixir - `miko-shell-elixir.example.yaml`

- Base image: `elixir:1.15-alpine`
- Pre-installed tools: hex, rebar, phoenix
- Scripts: test, deps, compile, run, format, credo, dialyzer, phx-server, iex, clean, coverage

### 🔥 Elixir/Phoenix - `miko-shell-phoenix.example.yaml`

- Base image: `elixir:1.15-alpine`
- Pre-installed tools: hex, rebar, phoenix, nodejs, npm, postgresql-client
- Scripts: new, setup, deps, compile, test, server, routes, migrate, rollback, seed, reset, gen-_, assets-_, format, credo, dialyzer, iex, console, clean, coverage

### 🐘 PHP - `miko-shell-php.example.yaml`

- Base image: `php:8.2-cli-alpine`
- Pre-installed tools: composer, pdo extensions
- Scripts: test, install, update, run, lint, phpcs, phpcbf, psalm, phpstan, clean

### 💎 Ruby - `miko-shell-ruby.example.yaml`

- Base image: `ruby:3.2-alpine`
- Pre-installed tools: bundler, build tools
- Scripts: test, install, update, run, server, console, migrate, seed, rubocop, rubocop-fix, clean

### 🚂 Ruby/Rails - `miko-shell-rails.example.yaml`

- Base image: `ruby:3.2-alpine`
- Pre-installed tools: bundler, rails, nodejs, npm, yarn, postgresql-dev
- Scripts: new, setup, install, update, test, server, console, routes, migrate, rollback, seed, reset, create-db, drop-db, setup-db, gen-_, assets-_, rubocop, rubocop-fix, brakeman, annotate, clean, logs

### ☕ Java - `miko-shell-java.example.yaml`

- Base image: `openjdk:17-jdk-alpine`
- Pre-installed tools: Maven
- Scripts: test, compile, run, package, clean, exec, dependency, site, verify

### ⚛️ Next.js - `miko-shell-nextjs.example.yaml`

- Base image: `node:18-alpine`
- Pre-installed tools: npm, TypeScript, Tailwind CSS, ESLint, Prettier
- Scripts: dev, build, start, test, test-watch, lint, lint-fix, format, type-check, clean

### 🐍 Django - `miko-shell-django.example.yaml`

- Base image: `python:3.11-slim`
- Pre-installed tools: Django, PostgreSQL client, Redis tools, Celery
- Scripts: runserver, startproject, startapp, migrate, makemigrations, createsuperuser, collectstatic, shell, dbshell, test, check, loaddata, dumpdata, flush, celery-worker, celery-beat, clean

### 🍃 Spring Boot - `miko-shell-spring-boot.example.yaml`

- Base image: `openjdk:17-jdk-alpine`
- Pre-installed tools: Maven, Spring Boot CLI
- Scripts: run, test, build, clean, package, bootrun, dev, actuator, profile, docker-build

### 🎨 Laravel - `miko-shell-laravel.example.yaml`

- Base image: `php:8.2-cli-alpine`
- Pre-installed tools: Composer, Laravel installer, Node.js, NPM
- Scripts: serve, new, install, update, migrate, migrate-fresh, seed, tinker, test, test-unit, test-feature, queue, schedule, artisan, npm-dev, npm-build, npm-watch, clear-cache, optimize, clean
- Scripts: test, compile, package, clean, install, run, format, verify, dependency

## Usage

1. Copy the example file for your language to your project root:

   ```bash
   cp examples/miko-shell-javascript.example.yaml miko-shell.yaml
   ```

2. Customize the configuration according to your project needs:

   - Update the `name` field
   - Modify the `image` if needed
   - Add/remove packages in `pre-install`
   - Customize scripts in `shell.scripts`

3. Initialize your project:

   ```bash
   miko-shell init  # Creates default config, or use your copied config
   ```

4. Start using miko-shell:
   ```bash
   miko-shell run test
   miko-shell shell
   ```

## Interactive Demo

Try the interactive demo script to explore different language configurations:

```bash
cd examples/
./demo.sh
```

The demo script will:

- Show all available examples
- Let you select a language
- Create a temporary demo environment
- Optionally build the container image
- Provide commands to try

## Detailed Framework Guides

For specific frameworks, check out these detailed guides:

- **Phoenix**: See [PHOENIX_EXAMPLE.md](PHOENIX_EXAMPLE.md) for detailed Phoenix workflow
- **Rails**: See [RAILS_EXAMPLE.md](RAILS_EXAMPLE.md) for detailed Rails workflow

## Configuration Structure

Each example follows the same structure:

```yaml
name: project-name # Project identifier
container: # Container configuration
  provider: docker # Container provider (docker/podman)
  image: base-image:tag # Base container image
  setup: # Commands to run during image build
    - package installation commands
shell:
  startup: # Commands to run when shell starts
    - initialization commands
  scripts: # Custom scripts
    - name: script-name
      commands:
        - command1
        - command2
```

## Creating Custom Examples

To create a custom example for a new language or framework:

1. Choose an appropriate base image
2. Add necessary package installations in `pre-install`
3. Define useful scripts for the development workflow
4. Test the configuration with `miko-shell`
5. Document any special requirements or usage notes

## Contributing

Feel free to contribute new examples or improvements to existing ones by submitting a pull request!
