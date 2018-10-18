#!/bin/bash

repository="davrodpin/mole"
install_path="/usr/local/bin"

latest_version=$(curl --silent "https://api.github.com/repos/${repository}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Get the os type
os_type=$(uname -sm)

# Extract the first part of os_type (Linux/Darwin)
os_type=${os_type% *}

# Convert os_type to lowercase
os_type=${os_type,,}

filename="mole${latest_version#v}.${os_type}-amd64.tar.gz"

download_link="https://github.com/${repository}/releases/download/${latest_version}/${filename}"

curl -L "${download_link}" | sudo tar xz -C "${install_path}"

echo -ne "\nmole ${latest_version} installed succesfully on ${install_path}\n"
