#!/bin/bash

DISTRO=
PACKAGE_MANAGER=
detect_package_manager() {
	if [[ -f /etc/os-release ]]; then
		source /etc/os-release
	else
		echo "error: file not found"
		exit 1
	fi

	DISTRO=$ID
	PACKAGE_MANAGER=""
	if [[ $DISTRO == "ubuntu" ]] || [[ $DISTRO == "debian" ]]; then
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
	echo "Installing feh..."
	install_native_packages feh
}

install_hyprpaper() {
	echo "Installing hyprpaper..."
	if [[ $PACKAGE_MANAGER == "apt"]]; then
		sudo apt update
		sudo apt install -y build-essential cmake pkg-config git \
			  libhyprlang-dev libhyprutils-dev libhyprgraphics-dev \
			  libwayland-dev libpango1.0-dev libjpeg-dev libpng-dev \
			  libwebp-dev libgles2-mesa-dev

		local tmp_dir=$(mktemp -d)
		git clone https://github.com/hyprwm/hyprpaper.git $tmp_dir
		(
			cd "$tmp_dir/hyprpaper"
			cmake --no-warn-unused-cli -DCMAKE_BUILD_TYPE:STRING=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr -S . -B ./build
			cmake --build ./build --config Release --target hyprpaper -j$(nproc)
			sudo cmake --install ./build
		)
		rm -rf $tmp_dir
	else
		install_native_packages hyprpaper
	fi
}

install_awww() {
	echo "Installing awww..."

    if ! command -v cargo &> /dev/null; then
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source "$HOME/.cargo/env"
    fi

	local tmp_dir=$(mktemp -d)
	git clone https://codeberg.org/LGFae/awww.git $tmp_dir
	(
		cd "$tmp_dir/awww"
		cargo build --release
	)
	rm -rf $tmp_dir
	mv target/release/awww target/release/awww-daemon /usr/local/bin/
}
