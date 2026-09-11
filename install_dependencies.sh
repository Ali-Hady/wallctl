#!/bin/bash


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

	echo $PACKAGE_MANAGER
}
