# wallctl

A lightweight, zero-dependency Linux wallpaper manager written in Go. `wallctl` fetches daily high-resolution wallpapers, indexes them locally, provides seamless chronological navigation, and sets your desktop background using native desktop utilities across modern Linux sessions.

---

## Demo

---

## Features

- **Zero Desktop Wrapper Fluff**: Eliminates flaky DBus script wrappers and third-party bindings. Interacts directly with native tools across KDE Plasma, GNOME, Sway, Hyprland, XFCE, and X11 standalone window managers.
- **Smart History & Batch Seeding**: Automatically seeds your local library with recent historical wallpapers on your very first run.
- **Bi-directional Navigation**: Move forward (`next`) and backward (`back`) through downloaded wallpapers without re-downloading.
- **Favorites & Aliases**: Tag wallpapers as favorites and assign memorable aliases to quickly apply them by name.
- **Atomic Local Persistence**: Metadata and state are saved in a clean, human-readable JSON store with atomic file operations to prevent corruption.
- **Native Automation**: Ships with automated `systemd` user service and timer integration to fetch daily wallpapers on schedule, with automatic detection (and installation) of the right wallpaper-setting tool for your session.

---

## Supported Desktop Environments & Compositors

`wallctl` automatically detects the active session and invokes the appropriate native utility:

| Desktop / Window Manager | Tool Used | Notes |
| :--- | :--- | :--- |
| **KDE Plasma** (Kubuntu, Fedora KDE, Arch) | `plasma-apply-wallpaperimage` | Official Plasma CLI (Wayland & X11) |
| **GNOME / Cinnamon / Budgie** | `gsettings` | Sets both Light and Dark mode (`picture-uri-dark`) |
| **Hyprland** | `swww` or `hyprctl hyprpaper` | Detects running Wayland wallpaper daemons |
| **Sway** | `swaymsg`, `swaybg`, or `wbg` | Native Wayland output IPC |
| **XFCE** | `xfconf-query` | Native XFCE backdrop configuration |
| **X11 Standalone WMs** (i3, bspwm, etc.) | `feh` | Standard `--bg-fill` fallback |

---

## Current Source: Bing Daily Archive

`wallctl` currently integrates with the **Bing Daily Image Archive**:
- **First Run Batch Seeding**: On the initial `wallctl fetch` (or whenever your store is empty), `wallctl` automatically queries Bing's archive for the past **7 days of wallpapers**, caches them chronologically, and applies the most recent one.
- **Single Updates**: Routine fetches retrieve the newest daily release and append it to your timeline.
- **Preserved Metadata**: Titles, copyright attribution, source URLs, and fetch timestamps are preserved with each image.

---

## Installation

### Prerequisites
- [Go](https://go.dev/doc/install) 1.22+ (for building from source)
- Git

### One-Step Automated Install
Clone the repository and execute the installer script:

```bash
git clone https://github.com/Ali-Hady/wallctl.git
cd wallctl
chmod +x install.sh uninstall.sh
./install.sh
```

#### What `install.sh` Does:

1. Compiles the binary with stripped debug symbols (`-ldflags="-s -w"`).
2. Installs the executable to `~/.local/bin/wallctl`.
3. Checks whether a supported wallpaper-setting tool (`swww`, `hyprctl`, `swaymsg`, `swaybg`, `wbg`, `plasma-apply-wallpaperimage`, `gsettings`, `xfconf-query`, `feh`) is already available, and warns if none are found.
4. Sets up a `systemd` user service (`~/.config/systemd/user/wallctl.service`) and a daily timer (`~/.config/systemd/user/wallctl.timer`):
   - The service passes through session environment variables (`WAYLAND_DISPLAY`, `DISPLAY`, `XDG_CURRENT_DESKTOP`, `XDG_SESSION_TYPE`, `SWAYSOCK`, `HYPRLAND_INSTANCE_SIGNATURE`) so it can correctly reach your compositor/session when triggered by systemd.
   - It waits 5 minutes after being started (`ExecStartPre=/bin/sleep 300`) before fetching, to avoid racing your desktop session on login/boot.
   - The timer runs daily at **09:00 AM** (`Persistent=true`, so a missed run fires as soon as the session is next active).
5. Imports the current compositor/session environment into the systemd user session (`systemctl --user import-environment ...`), then reloads and enables the timer.
6. Runs `./install_dependencies.sh` to install whichever native wallpaper-setting tool(s) your desktop environment needs.

> **Note**: Ensure `~/.local/bin` is in your `$PATH`. If not, add the following to your `~/.bashrc` or `~/.zshrc`:
> ```bash
> export PATH="$HOME/.local/bin:$PATH"
> ```
> The installer will remind you of this at the end if it detects `~/.local/bin` isn't already on your `$PATH`.

---

## Usage & Command Reference

### 1. Fetch Wallpapers

Fetch wallpapers from Bing and apply the newest one:

```bash
# Fetch latest (or seed last 7 on first run)
wallctl fetch

# Fetch a custom batch size
wallctl fetch --count 5

# Download and cache without changing the current desktop background
wallctl fetch --no-set
```

### 2. Navigate History

Step through your cached catalog chronologically:

```bash
# Move to the previous (older) wallpaper
wallctl back

# Move to the next (newer) wallpaper
wallctl next

# Navigate strictly through your marked favorites
wallctl back --favs
wallctl next --favs
```

### 3. Favorites & Aliases

Bookmark wallpapers so you can jump back to them:

```bash
# Mark the currently active wallpaper as favorite
wallctl favourite current

# Mark the current wallpaper as favorite and assign an alias
wallctl favourite current --alias hero

# Mark a specific image by its ID
wallctl favourite 20260909-bing --alias sunset
```

### 4. Direct Selection

Apply any cached wallpaper directly by its ID or assigned alias:

```bash
# Apply using alias
wallctl set hero

# Apply using ID
wallctl set 20260908-bing
```

### 5. Inspect History

List cached wallpapers along with their IDs, dates, titles, and favorite status:

```bash
# View last 10 entries
wallctl history

# View all saved favorites
wallctl history --favs

# Custom limit
wallctl history --limit 20
```

---

## File System & State Locations

`wallctl` adheres to the XDG Base Directory specification:

* **Metadata Store**: `~/.config/wallctl/index.json`
* **Cached Images**: `~/.cache/wallctl/images/`
* **Systemd Units**: `~/.config/systemd/user/wallctl.{service,timer}`

All metadata updates use atomic write-rename routines with temporary files to protect against incomplete writes during system shutdowns or restarts.

---

## Uninstallation

To cleanly remove the binary, disable background timers, and optionally clean up cached images and configuration:

```bash
./uninstall.sh
```

To run unattended and automatically remove all configuration and image cache directories:

```bash
./uninstall.sh --yes
```

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.