package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const execTimeout = 8 * time.Second

// setterAttempt represents one candidate strategy to apply a wallpaper.
type setterAttempt struct {
	name      string
	available func() bool
	apply     func(absPath string) error
}

// runCmd executes a binary under a bounded timeout and captures combined output.
// If the command fails, the output is bundled directly into the returned error.
func runCmd(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	if err != nil {
		if trimmed != "" {
			return trimmed, fmt.Errorf("%w: %s", err, trimmed)
		}
		return trimmed, err
	}
	return trimmed, nil
}

// commandExists verifies if a given binary is present in the current PATH.
// It searches directories in memory without spawning a subprocess.
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// hyprlandAttempts handles setting the wallpaper across common Hyprland helpers.
func hyprlandAttempts() []setterAttempt {
	return []setterAttempt{
		{
			name: "swww",
			available: func() bool {
				return commandExists("swww")
			},
			apply: func(p string) error {
				if _, err := runCmd("swww", "query"); err != nil {
					cmd := exec.Command("swww-daemon")
					cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					if err := cmd.Start(); err != nil {
						return fmt.Errorf("starting swww-daemon: %w", err)
					}
					time.Sleep(250 * time.Millisecond)
				}
				_, err := runCmd("swww", "img", p, "--transition-type", "fade")
				return err
			},
		},
		{
			name: "hyprpaper",
			available: func() bool {
				return commandExists("hyprpaper")
			},
			apply: func(p string) error {
			    if _, err := runCmd("hyprctl", "hyprpaper", "listactive"); err != nil {
					cmd := exec.Command("hyprpaper")
					cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					if err := cmd.Start(); err != nil {
					    return fmt.Errorf("starting hyprpaper: %w", err)
					}
					time.Sleep(250 * time.Millisecond)
				}
				if _, err := runCmd("hyprctl", "hyprpaper", "wallpaper", fmt.Sprintf(",%s", p)); err != nil {
					return err
				}
				return nil
			},
		},
	}
}

// applyGsettings applies wallpaper settings for both light and dark GNOME/Cinnamon themes.
func applyGsettings(absPath string) error {
	fileURI := "file://" + absPath
	schemas := []struct {
		schema string
		key    string
	}{
		{"org.gnome.desktop.background", "picture-uri"},
		{"org.gnome.desktop.background", "picture-uri-dark"},
	}

	var errs []string
	for _, s := range schemas {
		if _, err := runCmd("gsettings", "set", s.schema, s.key, fileURI); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) == len(schemas) {
		return fmt.Errorf("gsettings failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// applyXfconf dynamically discovers active XFCE display backdrops and applies the image.
func applyXfconf(absPath string) error {
	out, err := runCmd("xfconf-query", "-c", "xfce4-desktop", "-p", "/backdrop", "-l")
	if err != nil {
		return err
	}

	applied := false
	for _, prop := range strings.Split(out, "\n") {
		prop = strings.TrimSpace(prop)
		if strings.HasSuffix(prop, "/last-image") || strings.HasSuffix(prop, "/image-path") {
			if _, err := runCmd("xfconf-query", "-c", "xfce4-desktop", "-p", prop, "-s", absPath); err == nil {
				applied = true
			}
		}
	}

	if !applied {
		return fmt.Errorf("no active XFCE monitor backdrop properties located")
	}
	return nil
}

// SetDesktopWallpaper detects the active desktop/compositor and applies
// imagePath using the first available native tool.
func SetDesktopWallpaper(imagePath string) error {
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return fmt.Errorf("resolving absolute path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("reading image file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("target path is a directory: %s", absPath)
	}

	var attempts []setterAttempt
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))

	// ------------------------------------------------------------------------
	// Priority 1: Full Desktop Environments (KDE, GNOME, XFCE, Cinnamon, MATE)
	// Checked first so native tools handle both X11 and Wayland sessions
	// without falling through to standalone layer tools like swaybg.
	// ------------------------------------------------------------------------
	if strings.Contains(desktop, "kde") {
		attempts = append(attempts, setterAttempt{
			name:      "plasma-apply-wallpaperimage",
			available: func() bool { return commandExists("plasma-apply-wallpaperimage") },
			apply: func(p string) error {
				_, err := runCmd("plasma-apply-wallpaperimage", p)
				return err
			},
		})
	}

	if strings.Contains(desktop, "gnome") || strings.Contains(desktop, "cinnamon") || strings.Contains(desktop, "unity") {
		attempts = append(attempts, setterAttempt{
			name:      "gsettings",
			available: func() bool { return commandExists("gsettings") },
			apply:     applyGsettings,
		})
	}

	if strings.Contains(desktop, "xfce") {
		attempts = append(attempts, setterAttempt{
			name:      "xfconf-query",
			available: func() bool { return commandExists("xfconf-query") },
			apply:     applyXfconf,
		})
	}

	// ------------------------------------------------------------------------
	// Priority 2: Compositor-Specific Wayland Tiling WMs
	// ------------------------------------------------------------------------
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		attempts = append(attempts, hyprlandAttempts()...)
	}

	if os.Getenv("SWAYSOCK") != "" {
		attempts = append(attempts, setterAttempt{
			name:      "swaymsg",
			available: func() bool { return commandExists("swaymsg") },
			apply: func(p string) error {
				_, err := runCmd("swaymsg", "output", "*", "bg", p, "fill")
				return err
			},
		})
	}

	// ------------------------------------------------------------------------
	// Priority 3: Standalone Layer-Shell Tools (Wayland WMs like River, Labwc, etc.)
	// Only probed if WAYLAND_DISPLAY is present and the DE checks did not match.
	// ------------------------------------------------------------------------
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		attempts = append(attempts,
			setterAttempt{
				name:      "swaybg",
				available: func() bool { return commandExists("swaybg") },
				apply: func(p string) error {
					_ = exec.Command("pkill", "-x", "swaybg").Run()
					cmd := exec.Command("swaybg", "-i", p, "-m", "fill")
					cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					return cmd.Start()
				},
			},
			setterAttempt{
				name:      "wbg",
				available: func() bool { return commandExists("wbg") },
				apply: func(p string) error {
					_ = exec.Command("pkill", "-x", "wbg").Run()
					cmd := exec.Command("wbg", p)
					cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					return cmd.Start()
				},
			},
		)
	}

	// ------------------------------------------------------------------------
	// Priority 4: Generic DE Fallbacks (in case XDG_CURRENT_DESKTOP was unset)
	// ------------------------------------------------------------------------
	attempts = append(attempts,
		setterAttempt{
			name:      "plasma-apply-wallpaperimage (fallback)",
			available: func() bool { return commandExists("plasma-apply-wallpaperimage") },
			apply: func(p string) error {
				_, err := runCmd("plasma-apply-wallpaperimage", p)
				return err
			},
		},
		setterAttempt{
			name:      "gsettings (fallback)",
			available: func() bool { return commandExists("gsettings") },
			apply:     applyGsettings,
		},
		setterAttempt{
			name:      "xfconf-query (fallback)",
			available: func() bool { return commandExists("xfconf-query") },
			apply:     applyXfconf,
		},
	)

	// ------------------------------------------------------------------------
	// Priority 5: Standalone X11 WMs (i3, bspwm, Openbox, dwm)
	// ------------------------------------------------------------------------
	if os.Getenv("DISPLAY") != "" {
		attempts = append(attempts, setterAttempt{
			name:      "feh",
			available: func() bool { return commandExists("feh") },
			apply: func(p string) error {
				_, err := runCmd("feh", "--bg-fill", p)
				return err
			},
		})
	}

	var tried, failures []string
	for _, a := range attempts {
		if !a.available() {
			continue
		}
		tried = append(tried, a.name)
		if err := a.apply(absPath); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", a.name, err))
			log.Printf("wallctl: %s failed: %v", a.name, err)
			continue
		}
		log.Printf("wallctl: wallpaper set via %s", a.name)
		return nil
	}

	if len(tried) == 0 {
		return fmt.Errorf("no supported wallpaper setter found on system (checked environment and $PATH)")
	}
	return fmt.Errorf(
		"no working wallpaper setter found (tried: %s) — failures: %s",
		strings.Join(tried, ", "), strings.Join(failures, "; "),
	)
}
