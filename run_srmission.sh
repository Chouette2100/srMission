#!/bin/sh
set -eu

usage() {
	cat <<'EOF'
Usage:
  ./run_srmission.sh [mission] [noOfRooms] [viewingTime] [comment]

Options:
	-h, --help   Show this help and exit

Defaults (from .vscode/launch.json):
  mission    = daily
  noOfRooms  = 2
  viewingTime= 40
  comment    = 39

Environment defaults:
  SOPS_AGE_KEY_FILE=/home/chouette/.config/age/key2.txt
  SR_HEADLESS=1
  SR_BROWSER_BIN=/etc/profiles/per-user/chouette/bin/google-chrome-stable
  SR_WEBGL_MODE=off
	SR_CLEAN_JAR=0
EOF
}

is_positive_int() {
	case "$1" in
		""|*[!0-9]*)
			return 1
			;;
		*)
			[ "$1" -ge 1 ]
			;;
	esac
}

case "${1:-}" in
	-h|--help)
		usage
		exit 0
		;;
esac

if [ "$#" -gt 4 ]; then
	echo "Error: too many arguments." >&2
	usage >&2
	exit 1
fi

mission="${1:-daily}"
no_of_rooms="${2:-2}"
viewing_time="${3:-40}"
comment="${4:-39}"

if [ -z "$mission" ]; then
	echo "Error: mission must not be empty." >&2
	exit 1
fi

if ! is_positive_int "$no_of_rooms"; then
	echo "Error: noOfRooms must be a positive integer: $no_of_rooms" >&2
	exit 1
fi

if ! is_positive_int "$viewing_time"; then
	echo "Error: viewingTime must be a positive integer: $viewing_time" >&2
	exit 1
fi

: "${SOPS_AGE_KEY_FILE:=/home/chouette/.config/age/key2.txt}"
: "${SR_HEADLESS:=1}"
: "${SR_BROWSER_BIN:=/etc/profiles/per-user/chouette/bin/google-chrome-stable}"
: "${SR_WEBGL_MODE:=off}"
: "${SR_CLEAN_JAR:=0}"

export SOPS_AGE_KEY_FILE
export SR_HEADLESS
export SR_BROWSER_BIN
export SR_WEBGL_MODE
export SR_CLEAN_JAR

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

if [ ! -f "Env.enc.yml" ]; then
	echo "Error: Env.enc.yml is not found in $SCRIPT_DIR" >&2
	exit 1
fi

if [ ! -f "DBConfig.enc.yml" ]; then
	echo "Error: DBConfig.enc.yml is not found in $SCRIPT_DIR" >&2
	exit 1
fi

exec go run . "$mission" "$no_of_rooms" "$viewing_time" "$comment"
