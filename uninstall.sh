#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="wallctl"
BIN_DIR="$HOME/.local/bin"
SYSTEMD_DIR="$HOME/.config/systemd/user"
CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/$SERVICE_NAME"
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/$SERVICE_NAME"

AUTO_CONFIRM=false

# Check for non-interactive flag
while [[ $# -gt 0 ]]; do
  case "$1" in
    -y|--yes)
      AUTO_CONFIRM=true
      shift
      ;;
    -h|--help)
      echo "Usage: $(basename "$0") [-y|--yes]"
      echo "Uninstalls wallctl binary, services, and prompts for data removal."
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

echo "==> Stopping and disabling systemd timer & service..."
systemctl --user disable --now "$SERVICE_NAME.timer" 2>/dev/null || true
systemctl --user stop "$SERVICE_NAME.service" 2>/dev/null || true

echo "==> Removing systemd service and timer files..."
rm -f "$SYSTEMD_DIR/$SERVICE_NAME.service"
rm -f "$SYSTEMD_DIR/$SERVICE_NAME.timer"
systemctl --user daemon-reload

echo "==> Removing binary from $BIN_DIR..."
rm -f "$BIN_DIR/$SERVICE_NAME"

# Prompt for Cache Deletion
if [ -d "$CACHE_DIR" ]; then
  if [ "$AUTO_CONFIRM" = true ]; then
    rm -rf "$CACHE_DIR"
    echo "==> Cache removed."
  else
    echo ""
    read -rp "Delete downloaded wallpapers from cache ($CACHE_DIR)? [y/N] " response
    case "$response" in
      [yY][eE][sS]|[yY])
        rm -rf "$CACHE_DIR"
        echo "==> Cache directory removed."
        ;;
      *)
        echo "==> Cache preserved at: $CACHE_DIR"
        ;;
    esac
  fi
fi

# Prompt for Config Deletion
if [ -d "$CONFIG_DIR" ]; then
  if [ "$AUTO_CONFIRM" = true ]; then
    rm -rf "$CONFIG_DIR"
    echo "==> Config removed."
  else
    echo ""
    read -rp "Delete configuration files ($CONFIG_DIR)? [y/N] " response
    case "$response" in
      [yY][eE][sS]|[yY])
        rm -rf "$CONFIG_DIR"
        echo "==> Configuration directory removed."
        ;;
      *)
        echo "==> Configuration preserved at: $CONFIG_DIR"
        ;;
    esac
  fi
fi

echo ""
echo "wallctl uninstallation complete."