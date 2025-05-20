#!/usr/bin/env bash
set -euo pipefail

# Refresh base packages
sudo apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y \
    build-essential clang lld lldb gcc-10 gcc-10-multilib \
    gcc-i686-linux-gnu gcc-x86-64-linux-gnu binutils-i686-linux-gnu binutils-x86-64-linux-gnu \
    gcc-multilib make cmake gmake bmake \
    ninja-build meson pkg-config git wget curl unzip \
    python3 python3-venv python3-pip \
    qemu-system-x86 qemu-utils qemu-kvm \
    nodejs npm \
    rustc cargo

# Go 1.23
GO_VERSION=1.23.8
ARCH=$(uname -m)
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -o /tmp/go.tgz
sudo tar -C /usr/local -xzf /tmp/go.tgz
export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
echo 'export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH' >> ~/.profile

# Python environment
python3 -m venv .venv
source .venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt
pip install pytest pytest-xdist pexpect

deactivate

# Install Node.js 18 via NodeSource
NODE_MAJOR=18
curl -fsSL https://deb.nodesource.com/setup_${NODE_MAJOR}.x | sudo -E bash -
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs

# Install Rust 1.75.0 via rustup
RUST_VERSION=1.75.0
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain ${RUST_VERSION} --profile minimal
source "$HOME/.cargo/env"

# Install protoc via official script
curl -fsSL https://protobuf.dev/install.sh | bash -s -- -b /usr/local/bin -d /tmp/protoc

# Success
exit 0
