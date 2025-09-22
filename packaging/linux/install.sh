#!/usr/bin/env bash
set -euo pipefail

# Installation script for pre-push CLI
# Usage: ./install.sh [install_directory]
# Default: /usr/local/bin

APP_DIR="${1:-/usr/local/bin}"

# Use ARCHIVE_DIR if available (for makeself installers), otherwise fall back to script location
BASE_DIR="${ARCHIVE_DIR:-$(dirname "$(readlink -f "$0" || echo "$0")")}"

echo "[*] Installing pre-push CLI"
echo "[*] Target directory: $APP_DIR"
echo "[*] Source directory: $BASE_DIR"

install_one() {
  local src="$1"
  local dst="$APP_DIR/$(basename "$src")"
  # Create target directory if it doesn't exist
  mkdir -p "$APP_DIR" 2>/dev/null || true # Attempt to create, ignore if fails (e.g., no permissions)
  install -m 0755 "$src" "$dst"
  echo "  + $(basename "$src") -> $dst"
}

# Find and install the pre-push binary
if [[ -f "$BASE_DIR/pre-push" ]]; then
  install_one "$BASE_DIR/pre-push"
elif [[ -f "$BASE_DIR/pre-push.exe" ]]; then
  install_one "$BASE_DIR/pre-push.exe"
else
  echo "[ERROR] pre-push binary not found in $BASE_DIR"
  echo "[ERROR] Available files:"
  ls -la "$BASE_DIR" | head -10
  exit 1
fi

echo "[✓] Installation completed successfully!"
echo "[✓] You can now use: pre-push --help"