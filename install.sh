#!/bin/sh
set -e

# ccstatuswidgets installer
# Usage: curl -fsSL https://raw.githubusercontent.com/warunacds/ccstatuswidgets/main/install.sh | sh

REPO="warunacds/ccstatuswidgets"
BINARY_NAME="ccw"
SYMLINK_NAME="ccstatuswidgets"

# --- helpers ----------------------------------------------------------------

info() {
    printf '  \033[34m>\033[0m %s\n' "$1"
}

success() {
    printf '  \033[32m>\033[0m %s\n' "$1"
}

error() {
    printf '  \033[31m>\033[0m %s\n' "$1" >&2
}

die() {
    error "$1"
    exit 1
}

# --- detect OS and architecture ---------------------------------------------

detect_os() {
    os="$(uname -s)"
    case "$os" in
        Darwin)  echo "darwin" ;;
        Linux)   echo "linux" ;;
        *)       die "Unsupported operating system: $os" ;;
    esac
}

detect_arch() {
    arch="$(uname -m)"
    case "$arch" in
        x86_64)  echo "amd64" ;;
        amd64)   echo "amd64" ;;
        aarch64) echo "arm64" ;;
        arm64)   echo "arm64" ;;
        *)       die "Unsupported architecture: $arch" ;;
    esac
}

# --- resolve version --------------------------------------------------------

get_latest_version() {
    url="https://api.github.com/repos/${REPO}/releases/latest"
    response="$(curl -fsSL "$url" 2>/dev/null)" || die "Failed to fetch latest release from GitHub. Check your internet connection."

    # Extract tag_name from JSON without jq (POSIX-safe)
    version="$(printf '%s' "$response" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')"

    if [ -z "$version" ]; then
        die "Could not determine latest version from GitHub API response."
    fi

    echo "$version"
}

# --- choose install directory -----------------------------------------------

choose_install_dir() {
    if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
    elif command -v sudo >/dev/null 2>&1 && [ -d "/usr/local/bin" ]; then
        echo "/usr/local/bin"
    else
        dir="$HOME/.local/bin"
        mkdir -p "$dir"
        echo "$dir"
    fi
}

needs_sudo() {
    dir="$1"
    if [ -w "$dir" ]; then
        return 1
    fi
    return 0
}

# --- install ----------------------------------------------------------------

do_install() {
    # Check for curl
    if ! command -v curl >/dev/null 2>&1; then
        die "curl is required but not found. Please install curl and try again."
    fi

    printf '\n  \033[1mccstatuswidgets installer\033[0m\n\n'

    # Detect platform
    OS="$(detect_os)"
    ARCH="$(detect_arch)"
    info "Detected platform: ${OS}/${ARCH}"

    # Resolve version
    if [ -n "${CCW_VERSION:-}" ]; then
        VERSION="$CCW_VERSION"
        info "Using specified version: ${VERSION}"
    else
        info "Fetching latest release..."
        VERSION="$(get_latest_version)"
        info "Latest version: ${VERSION}"
    fi

    # Build download URL
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}_${OS}_${ARCH}"
    info "Downloading ${DOWNLOAD_URL}"

    # Download to temp file
    tmpdir="$(mktemp -d)"
    tmpfile="${tmpdir}/${BINARY_NAME}"
    trap 'rm -rf "$tmpdir"' EXIT

    if ! curl -fsSL -o "$tmpfile" "$DOWNLOAD_URL"; then
        die "Download failed. Please check that version ${VERSION} exists and has a binary for ${OS}/${ARCH}."
    fi

    chmod +x "$tmpfile"

    # Choose install directory
    INSTALL_DIR="$(choose_install_dir)"
    info "Installing to ${INSTALL_DIR}"

    # Install binary and symlink
    binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    symlink_path="${INSTALL_DIR}/${SYMLINK_NAME}"

    if needs_sudo "$INSTALL_DIR"; then
        info "Elevated permissions required — you may be prompted for your password."
        sudo cp "$tmpfile" "$binary_path"
        sudo chmod +x "$binary_path"
        sudo rm -f "$symlink_path"
        sudo ln -s "$binary_path" "$symlink_path"
    else
        cp "$tmpfile" "$binary_path"
        chmod +x "$binary_path"
        rm -f "$symlink_path"
        ln -s "$binary_path" "$symlink_path"
    fi

    # Verify installation
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        printf '\n'
        success "Installed ${BINARY_NAME} ${VERSION} to ${binary_path}"
        printf '\n'
        error "${INSTALL_DIR} is not in your PATH."
        info "Add it by running:"
        info "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        printf '\n'
    else
        printf '\n'
        success "Installed ${BINARY_NAME} ${VERSION} to ${binary_path}"
    fi

    # Done
    printf '\n'
    info "Get started by running:"
    printf '\n'
    info "  ccw init"
    printf '\n'
}

do_install
