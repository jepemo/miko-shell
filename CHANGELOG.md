# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release of Miko Shell
- Container-based development environment management
- Support for Docker and Podman container providers
- YAML-based configuration system
- Custom script execution within containers
- Interactive shell access to containers
- Automatic image building and caching
- Comprehensive test suite with 76.4% code coverage
- CI/CD pipeline with GitHub Actions
- Cross-platform binary releases (Linux, macOS, Windows)
- Dependabot integration for dependency updates
- Linting with golangci-lint
- Makefile for development tasks

### Features

- `miko-shell init` - Initialize new projects
- `miko-shell build` - Build container images
- `miko-shell run <command>` - Execute commands in containers
- `miko-shell shell` - Open interactive shell
- Configuration-based project setup
- Custom script definitions
- Pre-install commands for image customization
- Shell init hooks for environment setup

### Documentation

- Comprehensive README with usage examples
- Testing documentation with coverage reports
- Contributing guidelines
- License (MIT)

### Development

- Go 1.20+ support
- Modular package structure (`pkg/mikoshell`)
- Unit tests for all major components
- GitHub Actions for CI/CD
- Multi-platform build support
- Code quality checks and linting

## [0.1.0] - 2025-01-XX

### Added

- Initial project structure
- Basic CLI framework with Cobra
- Container provider abstraction
- Configuration management
- Image building capabilities
- Command execution in containers
- Interactive shell support
