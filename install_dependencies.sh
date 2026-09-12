#!/bin/bash

DISTRO=
PACKAGE_MANAGER=
DESKTOP=
SESSION_TYPE=
detect_package_manager() {
	if [[ -f /etc/os-release ]]; then
		source /etc/os-release
	else
		echo "error: file not found"
		exit 1
	fi

	DISTRO=$ID
	PACKAGE_MANAGER=""
	if [[ $DISTRO == "ubuntu" || $DISTRO == "debian" ]]; then
		PACKAGE_MANAGER="apt"
	elif [[ $DISTRO == "fedora" ]]; then
		PACKAGE_MANAGER="dnf"
	elif [[ $DISTRO == "arch" ]]; then
		PACKAGE_MANAGER="pacman"
	fi

	for par in $ID_LIKE; do
		case $par in
			arch)
				PACKAGE_MANAGER="pacman"
				break
				;;
			ubuntu|debian)
				PACKAGE_MANAGER="apt"
				break
				;;
			fedora|rhel)
				PACKAGE_MANAGER="dnf"
				break
				;;
		esac
	done
}

install_native_packages() {
	local packages=("$@")
	case $PACKAGE_MANAGER in
		apt)
			sudo apt update
			sudo apt install -y "${packages[@]}"
			;;
		dnf)
			sudo dnf install -y "${packages[@]}"
			;;
		pacman)
			sudo pacman -S --noconfirm "${packages[@]}"
			;;
	esac
}

install_feh() {
	if command -v feh &> /dev/null; then
		return 0
	fi
	echo "Installing feh..."
	install_native_packages feh
}

install_hyprpaper() {
	if command -v hyprpaper &> /dev/null; then
		return 0
	fi
	echo "Installing hyprpaper..."
	if [[ $PACKAGE_MANAGER == "apt" ]]; then
		sudo apt update
		sudo apt install -y build-essential cmake pkg-config git \
			libhyprlang-dev libhyprutils-dev libhyprgraphics-dev \
			libwayland-dev libpango1.0-dev libjpeg-dev libpng-dev \
			libwebp-dev libgles2-mesa-dev

		local tmp_dir=$(mktemp -d)
		git clone https://github.com/hyprwm/hyprpaper.git "$tmp_dir/hyprpaper"
		(
			cd "$tmp_dir/hyprpaper"
			cmake --no-warn-unused-cli -DCMAKE_BUILD_TYPE:STRING=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr -S . -B ./build
			cmake --build ./build --config Release --target hyprpaper -j$(nproc)
			sudo cmake --install ./build
		)
		rm -rf "$tmp_dir"
	else
		install_native_packages hyprpaper
	fi
}

install_awww() {
	if command -v awww &> /dev/null; then
		return 0
	fi

	echo "Installing awww..."
	if [[ $PACKAGE_MANAGER == "pacman" ]]; then
		install_native_packages awww
	else
		if ! command -v cargo &> /dev/null; then
			curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
			source "$HOME/.cargo/env"
		fi

		local tmp_dir=$(mktemp -d)
		git clone https://codeberg.org/LGFae/awww.git "$tmp_dir/awww"
		(
			cd "$tmp_dir/awww"
			cargo build --release
			sudo mv target/release/awww target/release/awww-daemon /usr/local/bin/
		)
		rm -rf "$tmp_dir"
	fi
}

install_wallpaper_tools() {
    DESKTOP="${XDG_CURRENT_DESKTOP,,}"
    SESSION_TYPE="${XDG_SESSION_TYPE,,}"

    case "$DESKTOP" in
        *kde*|*plasma*)
            # Plasma sets wallpaper natively (X11 or Wayland) via
            # plasma-apply-wallpaperimage — nothing to install
            return
            ;;
        *gnome*|*cinnamon*|*budgie*)
            # gsettings-based, already present
            return
            ;;
        *xfce*)
            # xfconf-query is native to XFCE, already present
            return
            ;;
    esac

    if [[ $SESSION_TYPE == *"wayland"* ]]; then
        case "$DESKTOP" in
            *hyprland*)
                install_hyprpaper
				install_awww
                ;;
            *sway*)
				install_awww
                ;;
            *)
                install_awww
                ;;
        esac
    elif [[ $SESSION_TYPE == *"x11"* ]]; then
        install_feh
    fi
}

main() {
	detect_package_manager
	install_wallpaper_tools
}

main "$@"
