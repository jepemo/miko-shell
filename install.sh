#!/usr/bin/env bash
set -euo pipefail

# miko-shell installer
# - Detects OS/ARCH
# - Downloads release artifact when available
# - Falls back to building from source (requires Go)
# - Supports uninstall

REPO="jepemo/miko-shell"
PROJECT="miko-shell"

COLOR="\033[1;34m"
NC="\033[0m"

say() { printf "${COLOR}==>${NC} %s\n" "$*"; }
err() { printf "\033[1;31mERROR:${NC} %s\n" "$*" >&2; }
print_info() { printf "${COLOR}[INFO]${NC} %s\n" "$*"; }

usage() {
  cat <<EOF
Install ${PROJECT}

Usage:
  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash

Options (pass after -s --):
  --help                 Show this help and exit
  --version X.Y.Z        Install a specific version tag (default: latest)
  --bin-dir DIR          Install destination (default: /usr/local/bin or ~/.local/bin)
  --asset URL            Override download URL explicitly
  --build-from-source    Build with 'go build' if release asset not found
  --bootstrap            Use bootstrap.sh (downloads Go temporarily) as fallback
  --uninstall            Remove ${PROJECT} from common locations

Examples:
  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash -s -- --help
  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash -s -- --version v1.0.0
  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash -s -- --bootstrap
EOF
}

OS() {
  case "${OSTYPE}" in
    linux*) echo linux ;;
    darwin*) echo darwin ;;
    msys*|cygwin*|win32*) echo windows ;;
    *) echo "unknown" ;;
  esac
}

ARCH() {
  local a
  a="$(uname -m 2>/dev/null || echo unknown)"
  case "$a" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    armv7l|armv7) echo armv7 ;;
    *) echo "$a" ;;
  esac
}

require_cmd() { command -v "$1" >/dev/null 2>&1 || { err "Missing required command: $1"; return 1; }; }

api_latest_tag() {
  require_cmd curl || return 1
  local result
  result="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | sed -n 's/ *"tag_name": *"\([^"]*\)".*/\1/p' | head -n1 || true)"
  if [ -z "$result" ]; then
    # Fallback to a known good version if API fails
    echo "v1.0.0"
  else
    echo "$result"
  fi
}

download() {
  local url="$1" out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fL --proto '=https' --tlsv1.2 -o "$out" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -O "$out" "$url"
  else
    err "Neither curl nor wget found"
    return 1
  fi
}

detect_install_dir() {
  local target="/usr/local/bin"
  if [ -n "${BIN_DIR:-}" ]; then
    echo "$BIN_DIR"; return
  fi
  if [ -w "$target" ]; then
    echo "$target"; return
  fi
  mkdir -p "$HOME/.local/bin" >/dev/null 2>&1 || true
  echo "$HOME/.local/bin"
}

uninstall() {
  say "Uninstalling ${PROJECT}"
  local removed=0
  for d in /usr/local/bin /usr/bin "$HOME/.local/bin" "$HOME/bin"; do
    if [ -x "$d/${PROJECT}" ]; then
      say "Removing $d/${PROJECT}"
      rm -f "$d/${PROJECT}" && removed=1 || true
    fi
  done
  if [ "$removed" = 0 ]; then
    say "No existing ${PROJECT} binary found in standard locations"
  fi
}

bootstrap_build() {
  say "Using bootstrap method (downloads Go temporarily)"
  local tmp_bootstrap bootstrap_script
  tmp_bootstrap="$(mktemp -d)"
  
  # Download bootstrap script
  bootstrap_script="$tmp_bootstrap/bootstrap.sh"
  if download "https://raw.githubusercontent.com/${REPO}/main/bootstrap.sh" "$bootstrap_script"; then
    chmod +x "$bootstrap_script"
    cd "$tmp_bootstrap"
    # Download repo and build
    if download "https://github.com/${REPO}/archive/main.tar.gz" "$tmp_bootstrap/repo.tar.gz"; then
      tar -xzf repo.tar.gz --strip-components=1
      bash bootstrap.sh
      if [ -f "miko-shell" ]; then
        install_to_path "$tmp_bootstrap/miko-shell"
        rm -rf "$tmp_bootstrap"
        return 0
      fi
    fi
  fi
  rm -rf "$tmp_bootstrap"
  err "Bootstrap build failed"
  return 1
}
build_from_source() {
  say "Building from source (requires Go)"
  if ! command -v go >/dev/null 2>&1; then
    err "Go is required to build from source but not found in PATH"
    err "Please install Go from https://golang.org/dl/ or use a release binary"
    return 1
  fi
  local tmp_build
  tmp_build="$(mktemp -d)"
  
  # Build in current repo if .git present, else go install latest
  if [ -f go.mod ] && [ -f main.go ]; then
    say "Detected repo checkout; building locally"
    go build -o "${tmp_build}/${PROJECT}" .
  else
    say "Using 'go install'"
    GO111MODULE=on GOBIN="$tmp_build" go install "github.com/${REPO}@${VERSION:-latest}"
  fi
  install_to_path "${tmp_build}/${PROJECT}"
  rm -rf "$tmp_build"
}

install_to_path() {
  local src="$1" dest_dir
  dest_dir="$(detect_install_dir)"
  mkdir -p "$dest_dir"
  say "Installing to $dest_dir/${PROJECT}"
  if command -v install >/dev/null 2>&1; then
    install -m 0755 "$src" "$dest_dir/${PROJECT}"
  else
    cp -f "$src" "$dest_dir/${PROJECT}"
    chmod 0755 "$dest_dir/${PROJECT}"
  fi
  say "Installed: $("$dest_dir/${PROJECT}" --version 2>/dev/null || echo "${PROJECT}")"
  
  # Check if binary is in PATH and provide instructions if not
  if ! command -v "${PROJECT}" >/dev/null 2>&1; then
    say ""
    say "📝 Add to PATH: export PATH=\"$dest_dir:\$PATH\""
    say "   For permanent: echo 'export PATH=\"$dest_dir:\$PATH\"' >> ~/.$(basename "$SHELL")rc"
    say "   Then reload: source ~/.$(basename "$SHELL")rc"
    say ""
  fi
}

extract_and_install() {
  local archive="$1" tmp_extract binary_path
  
  # Create temp directory
  tmp_extract="$(mktemp -d)"
  cd "$tmp_extract"
  
  # Extract based on file type
  print_info "Extracting archive..."
  if [[ "$archive" == *.zip ]]; then
    if ! command -v unzip >/dev/null 2>&1; then
      err "unzip is required to extract zip files"
      rm -rf "$tmp_extract"
      return 1
    fi
    unzip -q "$archive"
  else
    tar -xzf "$archive"
  fi
  
  # Find the binary (might be in subdirectory)
  binary_path=$(find . -name "$PROJECT" -type f | head -1)
  
  if [ -z "$binary_path" ]; then
    binary_path=$(find . -name "${PROJECT}.exe" -type f | head -1)
  fi
  
  if [ -z "$binary_path" ]; then
    err "Could not locate $PROJECT binary in archive"
    say "Archive contents:"
    find . -type f 2>/dev/null || true
    rm -rf "$tmp_extract"
    return 1
  fi
  
  # Make executable and install
  chmod +x "$binary_path"
  install_to_path "$binary_path"
  
  # Cleanup
  rm -rf "$tmp_extract"
}

choose_asset_candidates() {
  local tag="$1" os="$2" arch="$3"
  cat <<EOF
https://github.com/${REPO}/releases/download/${tag}/${PROJECT}_${tag#v}_${os}_${arch}.tar.gz
https://github.com/${REPO}/releases/download/${tag}/${PROJECT}_${os}_${arch}.tar.gz
https://github.com/${REPO}/releases/download/${tag}/${PROJECT}_${os}_${arch}.zip
https://github.com/${REPO}/releases/download/${tag}/${PROJECT}-${os}-${arch}.tar.gz
https://github.com/${REPO}/releases/download/${tag}/${PROJECT}-${os}-${arch}.zip
https://github.com/${REPO}/releases/download/${tag}/${PROJECT}_${os}_${arch}
https://github.com/${REPO}/releases/download/${tag}/${PROJECT}-${os}-${arch}
EOF
}

install_release() {
  local os arch tag url tmp_release
  os="$(OS)"; arch="$(ARCH)";
  if [ "$os" = "unknown" ]; then err "Unsupported OS"; return 1; fi
  if [ "$arch" = "unknown" ]; then err "Unsupported ARCH"; return 1; fi

  tag="${VERSION:-}"
  if [ -z "$tag" ] || [ "$tag" = "latest" ]; then
    say "Fetching latest release tag"
    tag="$(api_latest_tag)"
    [ -n "$tag" ] || { err "Unable to determine latest release"; return 1; }
  fi
  say "Looking for ${PROJECT} ${tag} release for ${os}/${arch}"

  tmp_release="$(mktemp -d)"

  if [ -n "${ASSET_URL:-}" ]; then
    url="${ASSET_URL}"
  else
    while read -r candidate; do
      say "Checking: $(basename "$candidate")"
      if curl -fsI "$candidate" >/dev/null 2>&1; then
        url="$candidate"; break
      fi
    done < <(choose_asset_candidates "$tag" "$os" "$arch")
  fi

  if [ -z "${url:-}" ]; then
    say "No matching release asset found for ${os}/${arch}"
    rm -rf "$tmp_release"
    return 1
  fi

  local file="$tmp_release/artifact"
  say "Downloading: $(basename "$url")"
  if download "$url" "$file"; then
    extract_and_install "$file"
    rm -rf "$tmp_release"
  else
    rm -rf "$tmp_release"
    return 1
  fi
}

main() {
  if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then usage; exit 0; fi
  # Parse args
  while [ $# -gt 0 ]; do
    case "$1" in
      --help|-h) usage; exit 0 ;;
      --version) VERSION="$2"; shift ;;
      --bin-dir) BIN_DIR="$2"; shift ;;
      --asset|--asset-url) ASSET_URL="$2"; shift ;;
      --uninstall) uninstall; exit 0 ;;
      --build-from-source) BUILD_FROM_SOURCE=1 ;;
      --bootstrap) BOOTSTRAP=1 ;;
      *) err "Unknown option: $1"; usage; exit 1 ;;
    esac
    shift || true
  done

  if ! install_release; then
    say "No precompiled release found for your platform"
    if [ "${BOOTSTRAP:-0}" = "1" ]; then
      bootstrap_build
    elif [ "${BUILD_FROM_SOURCE:-0}" = "1" ]; then
      build_from_source
    else
      say "Options:"
      say "  1. Try bootstrap (downloads Go temporarily):"
      say "     curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash -s -- --bootstrap"
      say "  2. Install Go and retry with --build-from-source"
      say "  3. Check https://github.com/${REPO}/releases for manual download"
      err "Installation failed - no precompiled binary available"
      exit 1
    fi
  fi

  say "Done"
}

main "$@"
