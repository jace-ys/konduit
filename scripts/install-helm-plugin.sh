#!/bin/bash
set -euo pipefail

VERSION=$(grep 'version:' "$HELM_PLUGIN_DIR/plugin.yaml" | cut -d'"' -f2)

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
esac

URL="https://github.com/jace-ys/konduit/releases/download/v${VERSION}/konduit_${VERSION}_${OS}_${ARCH}.tar.gz"

echo "Downloading Konduit v${VERSION} for ${OS}/${ARCH}..."
mkdir -p "$HELM_PLUGIN_DIR/bin"

curl -fsSL "$URL" | tar xz -C "$HELM_PLUGIN_DIR/bin"

echo "Konduit plugin for Helm installed successfully!"
