#!/bin/sh
# TRKEY installer for macOS and Linux. Installs for the current user.
#
#   curl -fsSL https://github.com/enxg/TRKEY/releases/latest/download/install.sh | sh
#   curl -fsSL https://github.com/enxg/TRKEY/releases/latest/download/install.sh | sh -s -- --uninstall
#
# On Linux X11 sessions, add --system to also install into /usr/share/X11/xkb (uses sudo).
#
# TRKEY_BASE_URL overrides where release files are downloaded from (e.g. file:///path/to/dist).
# TRKEY_XKB_SYSTEM_DIR overrides the system xkb directory used by --system.

set -eu

BASE_URL="${TRKEY_BASE_URL:-https://github.com/enxg/TRKEY/releases/latest/download}"
NAME="TRKEY (US)"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

usage() {
	say "usage: install.sh [--uninstall] [--system]"
	say ""
	say "  --uninstall  remove TRKEY"
	say "  --system     Linux only: also (un)install into the system xkb directory"
	say "               so TRKEY works in X11 sessions; uses sudo"
}

action=install
system=no
for arg in "$@"; do
	case $arg in
		--uninstall) action=uninstall ;;
		--system) system=yes ;;
		-h | --help) usage; exit 0 ;;
		*) usage >&2; die "unknown option: $arg" ;;
	esac
done

command -v curl > /dev/null || die "curl is required"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

fetch() {
	curl -fsSL "$BASE_URL/$1" -o "$tmp/$1" || die "could not download $BASE_URL/$1"
}

# ---------------------------------------------------------------- macOS

mac_dir="$HOME/Library/Keyboard Layouts"

mac_install() {
	command -v unzip > /dev/null || die "unzip is required"

	fetch TRKEY-macos.zip
	mkdir -p "$mac_dir"
	rm -rf "$mac_dir/TRKEY.bundle"
	unzip -q "$tmp/TRKEY-macos.zip" -d "$mac_dir"
	[ -d "$mac_dir/TRKEY.bundle" ] || die "TRKEY-macos.zip did not contain TRKEY.bundle"

	# macOS ignores layouts marked as downloaded from the internet
	xattr -dr com.apple.quarantine "$mac_dir/TRKEY.bundle" 2> /dev/null || true

	say "Installed $NAME to $mac_dir/TRKEY.bundle"
	say ""
	say "Next:"
	say "  1. Log out and back in."
	say "  2. System Settings > Keyboard > Input Sources > Edit > + > Turkish > $NAME"
}

mac_uninstall() {
	if [ ! -d "$mac_dir/TRKEY.bundle" ]; then
		say "$NAME is not installed."
		return
	fi

	rm -rf "$mac_dir/TRKEY.bundle"
	say "Removed $mac_dir/TRKEY.bundle"
	say "Remove $NAME from System Settings > Keyboard > Input Sources, then log out and back in."
}

# ---------------------------------------------------------------- Linux

xkb_dir="${XDG_CONFIG_HOME:-$HOME/.config}/xkb"
rules="$xkb_dir/rules/evdev.xml"
system_dir="${TRKEY_XKB_SYSTEM_DIR:-/usr/share/X11/xkb}"
system_file="$system_dir/symbols/trkey"

# as_root runs a command with sudo unless the target directory is already writable
as_root() {
	if [ -w "$system_dir/symbols" ]; then
		"$@"
	else
		command -v sudo > /dev/null || die "sudo is required for --system (or run as root)"
		sudo "$@"
	fi
}

system_install() {
	[ -d "$system_dir/symbols" ] || die "$system_dir/symbols not found; is xkeyboard-config installed?"

	say "Installing to $system_file (may ask for your password)"
	as_root rm -f "$system_file"
	as_root cp "$tmp/xkb/symbols/trkey" "$system_file"
	as_root chmod 644 "$system_file"

	say ""
	say "X11: TRKEY isn't listed in layout pickers there; select it by name:"
	say "  now:       setxkbmap trkey"
	say "  permanent: put this in /etc/X11/xorg.conf.d/00-trkey.conf"
	say ""
	say "    Section \"InputClass\""
	say "        Identifier \"trkey\""
	say "        MatchIsKeyboard \"on\""
	say "        Option \"XkbLayout\" \"trkey\""
	say "    EndSection"
	say ""
	say "  or set the layout to \"trkey\" in your desktop's keyboard config."
}

system_uninstall() {
	[ -e "$system_file" ] || [ -L "$system_file" ] || return 0

	say "Removing $system_file (may ask for your password)"
	as_root rm -f "$system_file"
	say "If you set XkbLayout \"trkey\" in /etc/X11/xorg.conf.d, remove that too."
}

linux_install() {
	command -v tar > /dev/null || die "tar is required"

	fetch TRKEY-linux.tar.gz
	tar -xzf "$tmp/TRKEY-linux.tar.gz" -C "$tmp"
	if [ ! -f "$tmp/xkb/symbols/trkey" ] || [ ! -f "$tmp/xkb/rules/evdev.xml" ]; then
		die "TRKEY-linux.tar.gz is missing xkb/symbols/trkey or xkb/rules/evdev.xml"
	fi

	mkdir -p "$xkb_dir/symbols" "$xkb_dir/rules"

	rm -f "$xkb_dir/symbols/trkey"
	cp "$tmp/xkb/symbols/trkey" "$xkb_dir/symbols/trkey"

	if [ ! -f "$rules" ]; then
		cp "$tmp/xkb/rules/evdev.xml" "$rules"
	elif grep -q '<name>trkey</name>' "$rules"; then
		: # already registered
	elif grep -q '</layoutList>' "$rules"; then
		cp "$rules" "$rules.bak"
		sed -n '/<layout>/,/<\/layout>/p' "$tmp/xkb/rules/evdev.xml" > "$tmp/block"
		awk -v block="$tmp/block" '
			/<\/layoutList>/ && !done {
				while ((getline line < block) > 0) print line
				done = 1
			}
			{ print }
		' "$rules.bak" > "$rules"
		say "Added $NAME to $rules (backup: $rules.bak)"
	else
		die "$rules has no <layoutList>; add the entry from xkb/rules/evdev.xml by hand"
	fi

	say "Installed $NAME to $xkb_dir"
	say ""
	say "Wayland (GNOME, KDE Plasma, ...):"
	say "  1. Log out and back in (needed the first time, so the desktop finds $xkb_dir)."
	say "  2. Add \"$NAME\" in your keyboard layout settings (listed under Turkish)."
	say ""

	if [ "$system" = yes ]; then
		system_install
	else
		say "X11 sessions don't read $xkb_dir; rerun with --system to install for X11."
	fi
}

linux_uninstall() {
	removed=no

	if [ -f "$xkb_dir/symbols/trkey" ] || [ -L "$xkb_dir/symbols/trkey" ]; then
		rm -f "$xkb_dir/symbols/trkey"
		removed=yes
	fi

	if [ -f "$rules" ] && grep -q '<name>trkey</name>' "$rules"; then
		cp "$rules" "$rules.bak"
		# drop the <layout> block that contains <name>trkey</name>
		awk '
			/<layout>/ { buf = $0 "\n"; inblock = 1; next }
			inblock {
				buf = buf $0 "\n"
				if (/<\/layout>/) {
					if (buf !~ /<name>trkey<\/name>/) printf "%s", buf
					inblock = 0
				}
				next
			}
			{ print }
		' "$rules.bak" > "$rules"

		if ! grep -q '<layout>' "$rules"; then
			rm -f "$rules" "$rules.bak"
		fi
		removed=yes
	fi

	if [ "$removed" = yes ]; then
		say "Removed $NAME from $xkb_dir"
		say "Remove it from your keyboard layout settings if it's still listed."
	fi

	if [ "$system" = yes ]; then
		if [ -e "$system_file" ] || [ -L "$system_file" ]; then
			system_uninstall
			removed=yes
		fi
	elif [ -e "$system_file" ] || [ -L "$system_file" ]; then
		say "Still installed system-wide at $system_file; rerun with --uninstall --system to remove it."
	fi

	if [ "$removed" = no ]; then
		say "$NAME is not installed."
	fi
}

# ----------------------------------------------------------------

case $(uname -s) in
	Darwin)
		[ "$system" = no ] || die "--system is only for Linux"
		"mac_$action"
		;;
	Linux) "linux_$action" ;;
	*) die "unsupported system: $(uname -s) (on Windows, use TRKEY-setup.exe)" ;;
esac
