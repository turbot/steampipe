#!/bin/sh
# TODO(everyone): Keep this script simple and easily auditable.

set -e

if ! command -v tar >/dev/null; then
	echo "Error: `tar` is required to install Steampipe." 1>&2
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
	"Darwin x86_64") target="darwin_amd64.zip" ;;
	"Darwin arm64") echo "Error: ARM is not supported yet." 1>&2;exit 1 ;;
	*) target="linux_amd64.tar.gz" ;;
	esac
fi

if [ $# -eq 0 ]; then
	steampipe_uri="https://github.com/turbot/steampipe/releases/latest/download/steampipe_${target}"
else
	steampipe_uri="https://github.com/turbot/steampipe/releases/download/${1}/steampipe_${target}"
fi

bin_dir="/usr/local/bin"
exe="$bin_dir/steampipe"

test -z "$tmp_dir" && tmp_dir="$(mktemp -d)"
mkdir -p "${tmp_dir}"
tmp_dir="${tmp_dir%/}"
case $(uname -sm) in
	"Darwin x86_64") zip_location="steampipe.zip" ;;
	"Darwin arm64") echo "Error: ARM is not supported yet." 1>&2;exit 1 ;;
	*) zip_location="steampipe.tar.gz" ;;
esac

echo "Downloading from $steampipe_uri"
if command -v wget >/dev/null; then
	if ! wget -q --show-progress -O "$zip_location" "$steampipe_uri"; then
        echo "Could not find version $1"
        exit 1
    fi
elif command -v curl >/dev/null; then
    if ! curl --silent --fail --location --progress-bar --output "$zip_location" "$steampipe_uri"; then
        echo "Could not find version $1"
        exit 1
    fi
else
    echo "Unable to find wget or curl. Cannot download."
    exit 1
fi

echo "Deflating downloaded archive"
tar -xf "$zip_location" -C "$tmp_dir"
echo "Installing"
install -d "$bin_dir"
install "$tmp_dir/steampipe" "$bin_dir"
chmod +x $exe
echo "Removing downloaded archive"
rm "$zip_location"

echo "Steampipe was installed successfully to $exe"
if command -v steampipe >/dev/null; then
	echo "Run 'steampipe --help' to get started"
else
    echo "Steampipe was installed, but could not be located. Are you sure `/use/local/bin` is exported?"
fi
