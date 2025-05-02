#!/usr/bin/env bash
set -euo pipefail

REPO="0xB1a60/runapp"  # Replace with your GitHub repo
VERSION="$(curl -s https://api.github.com/repos/$REPO/releases/latest | jq -r .tag_name)"
OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Normalize OS to 'macos' if it's 'darwin'
if [ "$OS" == "darwin" ]; then
  OS="macos"
fi

# Normalize architecture
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

curl -L "https://github.com/${REPO}/releases/download/${VERSION}/runapp-${OS}-${ARCH}" -o /tmp/runapp
chmod +x /tmp/runapp
sudo mv /tmp/runapp /usr/local/bin/runapp

echo "Installed to /usr/local/bin/runapp"
