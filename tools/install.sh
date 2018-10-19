#!/bin/bash
set -o pipefail

repository="davrodpin/mole"
install_path="/usr/local/bin"
temporary_file="/tmp/mole.tar.gz"

# Get the os architecture
os_arch=$(uname -m)

# Check if architecture is supported
if [ "${os_arch,,}" != "x86_64" ]; then
	echo "The ${os_arch} architecture is not supported"
	exit 1
fi

# Get the OS type
os_type=$(uname -s | tr '[:upper:]' '[:lower:]')

# Get latest version of mole available
latest_version=$(curl --silent --location --max-time 60 "https://api.github.com/repos/${repository}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ $? -ne 0 ]; then
	echo -ne "\nThere was an error trying to check what is the latest version of mole.\nPlease try again later.\n"
	exit 1
fi

filename="mole${latest_version#v}.${os_type}-amd64.tar.gz"
download_link="https://github.com/${repository}/releases/download/${latest_version}/${filename}"

# Download latest version of mole available
curl --location --max-time 60 "${download_link}" -o "${temporary_file}"
if [ $? -ne 0 ]; then
	echo -ne "\nThere was an error trying download the latest version of mole.\nPlease try again later.\n"
	exit 1
fi

# Extract the downloaded mole.tar.gz
sudo tar -xzf "${temporary_file}" -C "${install_path}"
if [ $? -ne 0 ]; then
	echo -ne "\nThere was an error trying extract the latest version of mole.\nPlease try again later.\n"
	exit 1
fi

echo -ne "\nmole ${latest_version} installed succesfully on ${install_path}\n"
