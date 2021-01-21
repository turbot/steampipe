#!/bin/sh
# TODO(everyone): Keep this script simple and easily auditable.

set -e

if ! command -v unzip >/dev/null; then
	echo "Error: unzip is required to install Steampipe." 1>&2
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

steampipe_install="/usr/local"
bin_dir="$steampipe_install/bin"
exe="$bin_dir/steampipe"

if [ ! -d "$bin_dir" ]; then
	mkdir -p "$bin_dir"
fi

echo "Downloading from $steampipe_uri"
if command -v wget >/dev/null; then
	wget -q --show-progress -O "$exe.zip" "$steampipe_uri"
elif command -v curl >/dev/null; then
    curl --fail --location --progress-bar --output "$exe.zip" "$steampipe_uri"
else
    echo "Unable to find wget or curl. Cannot download."
    exit 1
fi

echo "Deflating downloaded archive"
unzip -d "$bin_dir" -o "$exe.zip"
echo "Removing downloaded archive"
rm "$exe.zip"
echo "Setting necessary permissions"
chmod +x "$exe"

echo "Steampipe was installed successfully to $exe"
if command -v steampipe >/dev/null; then
	echo "Run 'steampipe --help' to get started"
else
    echo "Steampipe was installed, but could not be located. Are you sure `/use/local/bin` is exported?"
fi
