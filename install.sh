#!/bin/sh
# TODO(everyone): Keep this script simple and easily auditable.

set -e

if ! command -v unzip >/dev/null; then
	echo "Error: `unzip` is required to install Steampipe." 1>&2
	exit 1
fi

if ! command -v install >/dev/null; then
	echo "Error: `install` is required to install Steampipe." 1>&2
	exit 1
fi

if [ "$OS" = "Windows_NT" ]; then
	echo "Error: Windows is not supported yet." 1>&2
	exit 1
else
	case $(uname -sm) in
	"Darwin x86_64") target="darwin_amd64" ;;
	"Darwin arm64") echo "Error: ARM is not supported yet." 1>&2;exit 1 ;;
	*) target="linux_amd64" ;;
	esac
fi

if [ $# -eq 0 ]; then
	steampipe_uri="https://github.com/turbot/steampipe/releases/latest/download/steampipe_${target}.zip"
else
	steampipe_uri="https://github.com/denoland/deno/releases/download/${1}/steampipe_${target}.zip"
fi

bin_dir="/usr/local/bin"
exe="$bin_dir/steampipe"

test -z "$tmp_dir" && tmp_dir="$(mktemp -d)"
mkdir -p "${tmp_dir}"
tmp_dir="${tmp_dir%/}"
zip_location="$tmp_dir/steampipe.zip"

echo "Downloading from $steampipe_uri"
if command -v wget >/dev/null; then
	wget -q --show-progress -O "$zip_location" "$steampipe_uri"
elif command -v curl >/dev/null; then
    curl --fail --location --progress-bar --output "$zip_location" "$steampipe_uri"
else
    echo "Unable to find wget or curl. Cannot download."
    exit 1
fi

echo "Deflating downloaded archive"
unzip -d "$tmp_dir" -o "$zip_location"
echo "Installing"
install -d "$bin_dir"
install "$tmp_dir/steampipe" "$bin_dir"
echo "Removing downloaded archive"
rm "$zip_location"

echo "Steampipe was installed successfully to $exe"
if command -v steampipe >/dev/null; then
	echo "Run 'steampipe --help' to get started"
else
    echo "Steampipe was installed, but could not be located. Are you sure `/use/local/bin` is exported?"
fi
