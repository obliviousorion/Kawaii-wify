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

TARGET_BIN="${DEST_DIR}/kawaii-wify"
INSTALLED_BIN=""

if [ -x "$TARGET_BIN" ]; then
  INSTALLED_BIN="$TARGET_BIN"
elif command -v kawaii-wify >/dev/null 2>&1; then
  INSTALLED_BIN="$(command -v kawaii-wify)"
fi

# 1. Check for existing version vs latest release
if [ -n "$INSTALLED_BIN" ] && [ -z "$FORCE" ]; then
  LATEST_TAG=$(curl -fsSL https://api.github.com/repos/${REPO}/releases/latest 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)
  if [ -n "$LATEST_TAG" ]; then
    CURRENT_VER=$("$INSTALLED_BIN" --version 2>/dev/null || echo "")
    if echo "$CURRENT_VER" | grep -q "$LATEST_TAG"; then
      echo "Kawaii-Wify is already up to date ($LATEST_TAG). Nothing to do!"
      echo "To force reinstall, run: FORCE=1 curl -fsSL ... | bash"
      exit 0
    else
      echo "Update available! Installing latest release ($LATEST_TAG)..."
    fi
  fi
fi

# 2. Stop running daemon if active
WAS_RUNNING=0
if pgrep -x "kawaii-wify" >/dev/null 2>&1; then
  echo "Detected running kawaii-wify daemon. Stopping daemon to update binary..."
  WAS_RUNNING=1
  if [ -n "$INSTALLED_BIN" ]; then
    "$INSTALLED_BIN" stop >/dev/null 2>&1 || true
    sleep 1
  fi
  # If still running, terminate gracefully
  if pgrep -x "kawaii-wify" >/dev/null 2>&1; then
    pkill -x "kawaii-wify" >/dev/null 2>&1 || true
    sleep 0.5
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

# 3. Restart daemon if it was previously running
if [ "$WAS_RUNNING" -eq 1 ]; then
  echo "Restarting kawaii-wify daemon..."
  "${DEST_DIR}/kawaii-wify" start || echo "Notice: please run 'kawaii-wify start' to resume background monitoring."
fi

if ! command -v kawaii-wify >/dev/null 2>&1; then
  echo "Notice: ${DEST_DIR} is not in your PATH. Please add it to your shell configuration:"
  echo "  export PATH=\"${DEST_DIR}:\$PATH\""
fi

if [ -n "$INSTALLED_BIN" ]; then
  echo "Upgrade complete! Run 'kawaii-wify status' to verify."
else
  echo "Installation complete! Run 'kawaii-wify --help' to get started."
fi
