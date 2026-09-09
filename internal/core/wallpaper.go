package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func SetDesktopWallpaper(imagePath string) error {
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return fmt.Errorf("resolving absolute path: %w", err)
	}

	// 1. Hyprland & Modern Wayland WMs
	// Check if running under Hyprland or if common Wayland wallpaper tools are available
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		// Option A: swww (very common among Hyprland setups)
		if _, err := exec.LookPath("swww"); err == nil {
			cmd := exec.Command("swww", "img", absPath, "--transition-type", "fade")
			if err := cmd.Run(); err == nil {
				return nil
			}
		}

		// Option B: hyprpaper (official Hyprland daemon)
		if _, err := exec.LookPath("hyprctl"); err == nil {
			_ = exec.Command("hyprctl", "hyprpaper", "preload", absPath).Run()
			cmd := exec.Command("hyprctl", "hyprpaper", "wallpaper", fmt.Sprintf(",%s", absPath))
			if err := cmd.Run(); err == nil {
				// Clean up memory from previously preloaded images
				_ = exec.Command("hyprctl", "hyprpaper", "unload", "all").Run()
				return nil
			}
		}
	}

	// 2. KDE Plasma (Kubuntu, Fedora KDE, openSUSE)
	if _, err := exec.LookPath("plasma-apply-wallpaperimage"); err == nil {
		cmd := exec.Command("plasma-apply-wallpaperimage", absPath)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	// 3. GNOME / Cinnamon (Handles both Light & Dark modes)
	if _, err := exec.LookPath("gsettings"); err == nil {
		uri := "file://" + absPath
		_ = exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", uri).Run()
		if err := exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri-dark", uri).Run(); err == nil {
			return nil
		}
	}

	// 4. Sway
	if os.Getenv("SWAYSOCK") != "" {
		if _, err := exec.LookPath("swaymsg"); err == nil {
			return exec.Command("swaymsg", "output", "*", "bg", absPath, "fill").Run()
		}
	}

	// 5. Feh (Common for X11 standalone WMs: i3, bspwm, etc.)
	if _, err := exec.LookPath("feh"); err == nil {
		return exec.Command("feh", "--bg-fill", absPath).Run()
	}

	return fmt.Errorf("no supported desktop wallpaper setter found (tried swww/hyprpaper, plasma-apply-wallpaperimage, gsettings, swaymsg, feh)")
}
