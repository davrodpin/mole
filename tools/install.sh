#!/bin/bash

repository="davrodpin/mole"
install_path="/usr/local/bin"

# Get the os architecture
os_arch=$(uname -m)

# Check if architecture is supported
if [ "${os_arch,,}" != "x86_64" ]; then
	echo "The ${os_arch} architecture is not supported"
	exit 1
fi

# Get the OS type
os_type=$(uname -s)

# Convert os_type to lowercase
os_type=${os_type,,}

latest_version=$(curl --silent "https://api.github.com/repos/${repository}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
filename="mole${latest_version#v}.${os_type}-amd64.tar.gz"

download_link="https://github.com/${repository}/releases/download/${latest_version}/${filename}"

curl -L "${download_link}" | sudo tar xz -C "${install_path}"

echo -ne "\nmole ${latest_version} installed succesfully on ${install_path}\n"
