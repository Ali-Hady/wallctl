#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="$HOME/.local/bin"
SYSTEMD_DIR="$HOME/.config/systemd/user"
SERVICE_NAME="wallctl"

echo "==> Building $SERVICE_NAME..."
if ! command -v go >/dev/null 2>&1; then
  echo "Error: 'go' is not installed or not in PATH." >&2
  exit 1
fi
go build -ldflags="-s -w" -o "$SERVICE_NAME" main.go

echo "==> Installing binary to $BIN_DIR..."
mkdir -p "$BIN_DIR"
install -m 755 "$SERVICE_NAME" "$BIN_DIR/$SERVICE_NAME"
rm -f "$SERVICE_NAME"

echo "==> Checking for available wallpaper tools..."
SETTERS=("swww" "hyprctl" "swaymsg" "swaybg" "wbg" "plasma-apply-wallpaperimage" "gsettings" "xfconf-query" "feh")
FOUND=0
for tool in "${SETTERS[@]}"; do
  if command -v "$tool" >/dev/null 2>&1; then
    FOUND=1
    break
  fi
done

if [[ $FOUND -eq 0 ]]; then
  echo "Warning: No supported wallpaper tool detected (e.g. swww, hyprpaper, swaybg, feh)."
  echo "         Please install one appropriate for your desktop environment."
fi

echo "==> Setting up systemd user units..."
mkdir -p "$SYSTEMD_DIR"

# Write systemd service file
cat << 'EOF' > "$SYSTEMD_DIR/$SERVICE_NAME.service"
[Unit]
Description=Fetch and apply daily wallpaper via wallctl
PartOf=graphical-session.target
After=graphical-session.target

[Service]
Type=oneshot
KillMode=process
Environment="PATH=%h/.local/bin:/usr/local/bin:/usr/bin:/bin"
PassEnvironment=WAYLAND_DISPLAY DISPLAY XDG_CURRENT_DESKTOP XDG_SESSION_TYPE SWAYSOCK HYPRLAND_INSTANCE_SIGNATURE
ExecStartPre=/bin/sleep 300
ExecStart=%h/.local/bin/wallctl fetch

[Install]
WantedBy=graphical-session.target
EOF

# Write systemd timer file
cat << 'EOF' > "$SYSTEMD_DIR/$SERVICE_NAME.timer"
[Unit]
Description=Daily trigger for wallctl

[Timer]
OnCalendar=*-*-* 09:00:00
Persistent=true

[Install]
WantedBy=timers.target
EOF

# Ensure systemd user session knows about the compositor environment if running live
if command -v systemctl >/dev/null 2>&1; then
  systemctl --user import-environment WAYLAND_DISPLAY DISPLAY XDG_CURRENT_DESKTOP XDG_SESSION_TYPE SWAYSOCK HYPRLAND_INSTANCE_SIGNATURE 2>/dev/null || true

  echo "==> Reloading and enabling systemd timer..."
  systemctl --user daemon-reload
  systemctl --user enable --now "$SERVICE_NAME.timer"
fi

echo "==> Installing all needed wallpaper tools..."
./install_dependencies.sh

echo ""
echo "Installation successful!"

if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
  echo "Notice: $BIN_DIR is not currently in your \$PATH."
  echo "Add the following line to your ~/.bashrc or ~/.zshrc:"
  echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
  echo "Then run: source ~/.bashrc"
else
  echo "Run 'wallctl --help' to get started."
fi
