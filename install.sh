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

echo "==> Setting up systemd user units..."
mkdir -p "$SYSTEMD_DIR"

# Write systemd service file
cat << EOF > "$SYSTEMD_DIR/$SERVICE_NAME.service"
[Unit]
Description=Fetch and apply daily wallpaper via wallctl
PartOf=graphical-session.target
After=graphical-session.target

[Service]
Type=oneshot
Environment="PATH=$BIN_DIR:/usr/local/bin:/usr/bin:/bin"
ExecStart=$BIN_DIR/$SERVICE_NAME fetch

[Install]
WantedBy=graphical-session.target
EOF

# Write systemd timer file
cat << EOF > "$SYSTEMD_DIR/$SERVICE_NAME.timer"
[Unit]
Description=Daily trigger for wallctl

[Timer]
OnCalendar=*-*-* 09:00:00
Persistent=true

[Install]
WantedBy=timers.target
EOF

echo "==> Reloading and enabling systemd timer..."
systemctl --user daemon-reload
systemctl --user enable --now "$SERVICE_NAME.timer"

echo ""
echo "Installation successful!"

# Verify PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
  echo "Notice: $BIN_DIR is not currently in your \$PATH."
  echo "Add the following line to your ~/.bashrc or ~/.zshrc:"
  echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
  echo "Then run: source ~/.bashrc"
else
  echo "Run 'wallctl --help' to get started."
fi