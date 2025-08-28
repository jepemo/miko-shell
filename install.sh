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
  --uninstall            Remove ${PROJECT} from common locations

Examples:
  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash -s -- --help
  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash -s -- --version v1.0.0
  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash -s -- --uninstall
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
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/ *"tag_name": *"\([^"]*\)".*/\1/p' | head -n1
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

build_from_source() {
  say "Building from source (requires Go)"
  require_cmd go
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  # Build in current repo if .git present, else go install latest
  if [ -f go.mod ] && [ -f main.go ]; then
    say "Detected repo checkout; building locally"
    go build -o "${tmp}/${PROJECT}" .
  else
    say "Using 'go install'"
    GO111MODULE=on GOBIN="$tmp" go install "github.com/${REPO}@${VERSION:-latest}"
  fi
  install_to_path "${tmp}/${PROJECT}"
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
}

extract_and_install() {
  local archive="$1" tmp dir bin
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  case "$archive" in
    *.tar.gz|*.tgz)
      tar -C "$tmp" -xzf "$archive" ;;
    *.zip)
      require_cmd unzip
      unzip -q "$archive" -d "$tmp" ;;
    *)
      # Assume raw binary
      install_to_path "$archive"
      return
      ;;
  esac
  # Locate binary inside archive (portable: don't rely on -perm variants)
  bin="$(find "$tmp" -type f -name "${PROJECT}" 2>/dev/null | head -n1 || true)"
  if [ -z "$bin" ]; then
    bin="$(find "$tmp" -type f -name "${PROJECT}.exe" 2>/dev/null | head -n1 || true)"
  fi
  if [ -n "$bin" ]; then
    chmod +x "$bin" 2>/dev/null || true
  fi
  if [ -z "$bin" ]; then
    err "Could not locate ${PROJECT} binary in archive"
    return 1
  fi
  install_to_path "$bin"
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
  local os arch tag url tmp
  os="$(OS)"; arch="$(ARCH)";
  if [ "$os" = "unknown" ]; then err "Unsupported OS"; return 1; fi
  if [ "$arch" = "unknown" ]; then err "Unsupported ARCH"; return 1; fi

  tag="${VERSION:-}"
  if [ -z "$tag" ] || [ "$tag" = "latest" ]; then
    say "Fetching latest release tag"
    tag="$(api_latest_tag)"
    [ -n "$tag" ] || { err "Unable to determine latest release"; return 1; }
  fi
  say "Installing ${PROJECT} ${tag} for ${os}/${arch}"

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  if [ -n "${ASSET_URL:-}" ]; then
    url="${ASSET_URL}"
  else
    while read -r candidate; do
      say "Trying asset: ${candidate}"
      if curl -fsI "$candidate" >/dev/null 2>&1; then
        url="$candidate"; break
      fi
    done < <(choose_asset_candidates "$tag" "$os" "$arch")
  fi

  if [ -z "${url:-}" ]; then
    err "No matching release asset found"
    return 1
  fi

  local file="$tmp/artifact"
  download "$url" "$file"
  extract_and_install "$file"
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
      *) err "Unknown option: $1"; usage; exit 1 ;;
    esac
    shift || true
  done

  if ! install_release; then
    if [ "${BUILD_FROM_SOURCE:-0}" = "1" ]; then
      build_from_source
    else
      say "Falling back to --build-from-source (pass explicitly to skip this message)"
      build_from_source
    fi
  fi

  say "Done"
}

main "$@"
