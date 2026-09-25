#!/usr/bin/env bash
set -e

REPO="obliviousorion/Kawaii-wify"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux)
    case "$ARCH" in
      x86_64|amd64)
        ASSET="kawaii-wify-linux-amd64"
        ;;
      aarch64|arm64)
        ASSET="kawaii-wify-linux-arm64"
        ;;
      *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
    esac
    ;;
  darwin)
    case "$ARCH" in
      x86_64|amd64)
        ASSET="kawaii-wify-darwin-amd64"
        ;;
      arm64)
        ASSET="kawaii-wify-darwin-arm64"
        ;;
      *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unsupported operating system: $OS"
    exit 1
    ;;
esac

DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
DEST_DIR="/usr/local/bin"

if [ ! -w "$DEST_DIR" ]; then
  if command -v sudo >/dev/null 2>&1; then
    USE_SUDO="sudo"
  else
    DEST_DIR="$HOME/.local/bin"
    mkdir -p "$DEST_DIR"
    USE_SUDO=""
  fi
fi

TMP_FILE="$(mktemp)"

echo "Downloading ${ASSET} from GitHub..."
curl -fsSL -o "$TMP_FILE" "$DOWNLOAD_URL" || {
  echo "Error: Failed to download release from $DOWNLOAD_URL"
  echo "Please check if a release has been published on GitHub."
  rm -f "$TMP_FILE"
  exit 1
}

chmod +x "$TMP_FILE"

echo "Installing to ${DEST_DIR}/kawaii-wify..."
$USE_SUDO mv "$TMP_FILE" "${DEST_DIR}/kawaii-wify"

echo "Successfully installed kawaii-wify to ${DEST_DIR}/kawaii-wify"

if ! command -v kawaii-wify >/dev/null 2>&1; then
  echo "Notice: ${DEST_DIR} is not in your PATH. Please add it to your shell configuration:"
  echo "  export PATH=\"${DEST_DIR}:\$PATH\""
fi
