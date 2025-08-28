#!/bin/bash

# bootstrap.sh - Download Go and build miko-shell project
# This script downloads a minimal Go binary and uses it to build the project

set -e  # Exit on any error

# Configuration
GO_VERSION="1.23.4"  # Updated to meet module requirements
PROJECT_NAME="miko-shell"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMP_DIR="${SCRIPT_DIR}/.bootstrap"
GO_DIR="${TEMP_DIR}/go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          log_error "Unsupported OS: $(uname -s)"; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        armv6l)         arch="armv6l" ;;
        *)              log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Download and extract Go
download_go() {
    local platform="$1"
    local go_archive="go${GO_VERSION}.${platform}.tar.gz"
    local download_url="https://golang.org/dl/${go_archive}"
    
    log_info "Detecting platform: ${platform}"
    log_info "Downloading Go ${GO_VERSION} for ${platform}..."
    
    mkdir -p "${TEMP_DIR}"
    
    if command -v wget >/dev/null 2>&1; then
        wget -q --show-progress -O "${TEMP_DIR}/${go_archive}" "${download_url}"
    elif command -v curl >/dev/null 2>&1; then
        curl -L --progress-bar -o "${TEMP_DIR}/${go_archive}" "${download_url}"
    else
        log_error "Neither wget nor curl found. Please install one of them."
        exit 1
    fi
    
    log_info "Extracting Go..."
    tar -C "${TEMP_DIR}" -xzf "${TEMP_DIR}/${go_archive}"
    rm "${TEMP_DIR}/${go_archive}"
    
    log_success "Go ${GO_VERSION} downloaded and extracted"
}

# Check if Go is already available
check_existing_go() {
    if command -v go >/dev/null 2>&1; then
        local existing_version=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Found existing Go installation: ${existing_version}"
        
        # Compare versions (basic comparison)
        if [ "${existing_version}" = "${GO_VERSION}" ]; then
            log_success "Existing Go version matches required version"
            return 0
        else
            log_warning "Existing Go version (${existing_version}) differs from required (${GO_VERSION})"
            log_info "Will use downloaded Go for consistent builds"
        fi
    fi
    return 1
}

# Build the project
build_project() {
    local go_binary="$1"
    
    log_info "Building ${PROJECT_NAME}..."
    
    # Get version from git or use 'dev'
    local version=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    
    # Set Go environment
    export GOROOT="${GO_DIR}"
    export PATH="${GO_DIR}/bin:${PATH}"
    
    # Build with version info
    "${go_binary}" build -ldflags="-X '${PROJECT_NAME}/cmd.version=${version}'" -o "${PROJECT_NAME}" .
    
    log_success "Built ${PROJECT_NAME} successfully!"
    log_info "Binary location: ${SCRIPT_DIR}/${PROJECT_NAME}"
}

# Show help information
show_help() {
    cat << EOF
Bootstrap script for ${PROJECT_NAME}

USAGE:
    ./bootstrap.sh [OPTION]

OPTIONS:
    --clean     Remove build artifacts (binaries, build/, .bootstrap/)
                Same as 'make clean' but also removes miko-shell-host
    --help      Show this help message
    
EXAMPLES:
    ./bootstrap.sh          # Build the project
    ./bootstrap.sh --clean  # Clean build artifacts
    ./bootstrap.sh --help   # Show this help
    
DESCRIPTION:
    This script downloads Go ${GO_VERSION} (if needed) and builds the ${PROJECT_NAME} project.
    It's useful for bootstrapping development on systems without Go installed.

EOF
}

# Clean up build artifacts (like make clean but also removes miko-shell-host)
clean_project() {
    log_info "Cleaning build artifacts..."
    
    # Remove binaries
    if [ -f "./${PROJECT_NAME}" ]; then
        log_info "Removing ${PROJECT_NAME}"
        rm -f "./${PROJECT_NAME}"
    fi
    
    if [ -f "./${PROJECT_NAME}-host" ]; then
        log_info "Removing ${PROJECT_NAME}-host"
        rm -f "./${PROJECT_NAME}-host"
    fi
    
    # Remove build directory
    if [ -d "build" ]; then
        log_info "Removing build directory"
        rm -rf "build"
    fi
    
    # Remove bootstrap temporary directory
    if [ -d ".bootstrap" ]; then
        log_info "Removing .bootstrap directory"
        rm -rf ".bootstrap"
    fi
    
    log_success "Clean completed successfully!"
}

# Clean up temporary files
cleanup() {
    if [ -d "${TEMP_DIR}" ]; then
        log_info "Cleaning up temporary files..."
        rm -rf "${TEMP_DIR}"
    fi
}

# Main execution
main() {
    # Check for parameters
    case "$1" in
        --clean)
            log_info "Clean mode - removing build artifacts"
            log_info "=================================="
            cd "${SCRIPT_DIR}"
            clean_project
            exit 0
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        "")
            # No parameters, proceed with build
            ;;
        *)
            log_error "Unknown parameter: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
    
    log_info "Bootstrap script for ${PROJECT_NAME}"
    log_info "======================================"
    
    # Change to script directory
    cd "${SCRIPT_DIR}"

    # Delete current binary if it exists
    if [ -f "./${PROJECT_NAME}" ]; then
        log_info "Removing existing binary: ./${PROJECT_NAME}"
        rm -f "./${PROJECT_NAME}"
    fi
    
    # Check if we need to download Go
    local go_binary=""
    
    if check_existing_go; then
        log_info "Using existing Go installation"
        go_binary="go"
    else
        log_info "Downloading Go ${GO_VERSION}..."
        local platform=$(detect_platform)
        download_go "${platform}"
        go_binary="${GO_DIR}/bin/go"
    fi
    
    # Verify Go installation
    log_info "Verifying Go installation..."
    "${go_binary}" version
    
    # Download dependencies
    log_info "Downloading Go modules..."
    "${go_binary}" mod download
    
    # Build the project
    build_project "${go_binary}"
    
    # Test the binary
    if [ -f "./${PROJECT_NAME}" ]; then
        log_info "Testing the binary..."
        "./${PROJECT_NAME}" --version
        log_success "Bootstrap completed successfully!"
        log_info "You can now run: ./${PROJECT_NAME}"
    else
        log_error "Build failed - binary not found"
        exit 1
    fi
}

# Set up cleanup trap
trap cleanup EXIT

# Run main function
main "$@"
