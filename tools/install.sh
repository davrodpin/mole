#!/bin/bash

# MIT License
#
# Copyright (c) 2018 David Pinheiro
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

set -o pipefail

repository="davrodpin/mole"
install_path="/usr/local/bin"
curl_timeout_seconds=60

# Get the os architecture
os_arch=$(uname -m | tr '[:upper:]' '[:lower:]')

# Check if architecture is supported
if [ "${os_arch}" != "x86_64" ]; then
	echo "The ${os_arch} architecture is not supported"
	exit 1
fi

# Get the OS type
os_type=$(uname -s | tr '[:upper:]' '[:lower:]')

# Get latest version of mole available
	latest_version=$(curl --silent --location --max-time "${curl_timeout_seconds}" "https://api.github.com/repos/${repository}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ $? -ne 0 ]; then
	echo -ne "\nThere was an error trying to check what is the latest version of mole.\nPlease try again later.\n"
	exit 1
fi

filename="mole${latest_version#v}.${os_type}-amd64.tar.gz"
download_link="https://github.com/${repository}/releases/download/${latest_version}/${filename}"

curl --silent --location --max-time "${curl_timeout_seconds}" "${download_link}" | sudo tar -xz -C "${install_path}" 2>/dev/null|| {
	echo -ne "\nThere was an error trying to install the latest version of mole.\nPlease try again later.\n"
	exit 1
}

echo -ne "\nmole ${latest_version} installed succesfully on ${install_path}\n"
