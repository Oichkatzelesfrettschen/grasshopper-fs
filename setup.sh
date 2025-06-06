#!/usr/bin/env bash
set -euo pipefail

# ensure we run from the repository root
cd "$(dirname "$(readlink -f "$0")")"

# log file for any installation failures
LOG_FILE=/tmp/setup-failures.log
echo "" > "$LOG_FILE"

# attempt a pip install fallback when apt packages fail, if applicable
apt_pip_fallback() {
  pkg="$1"
  if [[ "$pkg" == python3-* ]]; then
    pip_pkg="${pkg#python3-}"
    if ! pip3 install --no-cache-dir "$pip_pkg" >/dev/null 2>&1; then
      echo "PIP fallback failed: $pip_pkg" >> "$LOG_FILE"
    fi
  fi
}

# helper to install python packages with failure logging
pip_install() {
  pkg="$1"
  if ! pip3 install --no-cache-dir "$pkg" >/dev/null 2>&1; then
    echo "PIP install failed: $pkg" >> "$LOG_FILE"
  fi
}
export DEBIAN_FRONTEND=noninteractive

# helper to pin to the repo's exact version if it exists
apt_pin_install() {
  pkg="$1"
  ver=$(apt-cache show "$pkg" 2>/dev/null | awk '/^Version:/{print $2; exit}')
  if [ -n "$ver" ]; then
    if ! apt-get install -y "${pkg}=${ver}" >/dev/null 2>&1; then
      echo "APT install failed: $pkg" >> "$LOG_FILE"
      apt_pip_fallback "$pkg"
    fi
  else
    if ! apt-get install -y "$pkg" >/dev/null 2>&1; then
      echo "APT install failed: $pkg" >> "$LOG_FILE"
      apt_pip_fallback "$pkg"
    fi
  fi
}

# enable foreign architectures for cross-compilation
for arch in i386 armel armhf arm64 riscv64 powerpc ppc64el ia64; do
  dpkg --add-architecture "$arch"
done

apt-get update -y || echo "APT update failed" >> "$LOG_FILE"

# core build tools, formatters, analysis, science libs
for pkg in \
  build-essential gcc g++ clang lld llvm \
  clang-format uncrustify astyle editorconfig pre-commit \
  make bmake ninja-build cmake meson \
  autoconf automake libtool m4 gawk flex bison byacc \
  pkg-config file ca-certificates curl git unzip \
  libopenblas-dev liblapack-dev libeigen3-dev \
  strace ltrace linux-perf systemtap systemtap-sdt-dev crash \
  valgrind kcachegrind trace-cmd kernelshark \
  libasan6 libubsan1 likwid hwloc; do
  apt_pin_install "$pkg"
done

# Python & deep-learning / MLOps
for pkg in \
  python3 python3-pip python3-dev python3-venv python3-wheel \
  python3-numpy python3-scipy python3-pandas \
  python3-matplotlib python3-scikit-learn \
  python3-torch python3-torchvision python3-torchaudio \
  python3-onnx python3-onnxruntime; do
  apt_pin_install "$pkg"
done

for pkg in \
  tensorflow-cpu jax jaxlib \
  tensorflow-model-optimization mlflow onnxruntime-tools; do
  pip_install "$pkg"
done

# QEMU emulation for foreign binaries
for pkg in \
  qemu-user-static \
  qemu-system-x86 qemu-system-arm qemu-system-aarch64 \
  qemu-system-riscv64 qemu-system-ppc qemu-system-ppc64 qemu-utils; do
  apt_pin_install "$pkg"
done

# multi-arch cross-compilers
for pkg in \
  bcc bin86 elks-libc \
  gcc-ia64-linux-gnu g++-ia64-linux-gnu \
  gcc-i686-linux-gnu g++-i686-linux-gnu \
  gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
  gcc-arm-linux-gnueabi g++-arm-linux-gnueabi \
  gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf \
  gcc-riscv64-linux-gnu g++-riscv64-linux-gnu \
  gcc-powerpc-linux-gnu g++-powerpc-linux-gnu \
  gcc-powerpc64-linux-gnu g++-powerpc64-linux-gnu \
  gcc-powerpc64le-linux-gnu g++-powerpc64le-linux-gnu \
  gcc-m68k-linux-gnu g++-m68k-linux-gnu \
  gcc-hppa-linux-gnu g++-hppa-linux-gnu \
  gcc-loongarch64-linux-gnu g++-loongarch64-linux-gnu \
  gcc-mips-linux-gnu g++-mips-linux-gnu \
  gcc-mipsel-linux-gnu g++-mipsel-linux-gnu \
  gcc-mips64-linux-gnuabi64 g++-mips64-linux-gnuabi64 \
  gcc-mips64el-linux-gnuabi64 g++-mips64el-linux-gnuabi64; do
  apt_pin_install "$pkg"
done

# high-level language runtimes and tools
for pkg in \
  golang-go nodejs npm typescript \
  rustc cargo clippy rustfmt \
  lua5.4 liblua5.4-dev luarocks \
  ghc cabal-install hlint stylish-haskell \
  sbcl ecl clisp cl-quicklisp slime cl-asdf \
  ldc gdc dmd-compiler dub libphobos-dev \
  chicken-bin libchicken-dev chicken-doc \
  openjdk-17-jdk maven gradle dotnet-sdk-8 mono-complete \
  swift swift-lldb swiftpm kotlin gradle-plugin-kotlin \
  ruby ruby-dev gem bundler php-cli php-dev composer phpunit \
  r-base r-base-dev dart flutter gnat gprbuild gfortran gnucobol \
  fpc lazarus zig nim nimble crystal shards gforth; do
  apt_pin_install "$pkg"
done

# GUI & desktop-dev frameworks
for pkg in \
  libqt5-dev qtcreator libqt6-dev \
  libgtk1.2-dev libgtk2.0-dev libgtk-3-dev libgtk-4-dev \
  libfltk1.3-dev xorg-dev libx11-dev libxext-dev \
  libmotif-dev openmotif cde \
  xfce4-dev-tools libxfce4ui-2-dev lxde-core lxqt-dev-tools \
  libefl-dev libeina-dev \
  libwxgtk3.0-dev libwxgtk3.0-gtk3-dev \
  libsdl2-dev libsdl2-image-dev libsdl2-ttf-dev \
  libglfw3-dev libglew-dev; do
  apt_pin_install "$pkg"
done

# containers, virtualization, HPC, debug
for pkg in \
  docker.io podman buildah virt-manager libvirt-daemon-system qemu-kvm \
  gdb lldb perf gcovr lcov bcc-tools bpftrace \
  openmpi-bin libopenmpi-dev mpich; do
  apt_pin_install "$pkg"
done

# IA-16 (8086/286) cross-compiler
IA16_VER=$(curl -fsSL https://api.github.com/repos/tkchia/gcc-ia16/releases/latest | awk -F\" '/tag_name/{print $4; exit}')
if [ -z "$IA16_VER" ] || ! curl -fsSL "https://github.com/tkchia/gcc-ia16/releases/download/${IA16_VER}/ia16-elf-gcc-linux64.tar.xz" | tar -Jx -C /opt; then
  echo "IA16 toolchain install failed" >> "$LOG_FILE"
fi
echo 'export PATH=/opt/ia16-elf-gcc/bin:$PATH' > /etc/profile.d/ia16.sh
export PATH=/opt/ia16-elf-gcc/bin:$PATH

# protoc installer (pinned)
PROTO_VERSION=25.1
if ! curl -fsSL "https://raw.githubusercontent.com/protocolbuffers/protobuf/v${PROTO_VERSION}/protoc-${PROTO_VERSION}-linux-x86_64.zip" -o /tmp/protoc.zip; then
  echo "protoc download failed" >> "$LOG_FILE"
else
  unzip -d /usr/local /tmp/protoc.zip >/dev/null 2>&1 || echo "protoc unzip failed" >> "$LOG_FILE"
  rm /tmp/protoc.zip
fi

# gmake alias
command -v gmake >/dev/null 2>&1 || ln -s "$(command -v make)" /usr/local/bin/gmake

# Install Go 1.23 or newer
GO_VERSION=1.23.8
ARCH=$(uname -m)
if ! curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -o /tmp/go.tgz; then
  echo "Go download failed" >> "$LOG_FILE"
else
  rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tgz || echo "Go extract failed" >> "$LOG_FILE"
fi
export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
echo 'export PATH=$HOME/go/bin:/usr/local/go/bin:$PATH' > /etc/profile.d/go.sh

# Fetch Go modules so offline commands succeed
if ! go mod download >/dev/null 2>&1; then
  echo "Go mod download failed" >> "$LOG_FILE"
fi
if ! go mod vendor >/dev/null 2>&1; then
  echo "Go mod vendor failed" >> "$LOG_FILE"
fi
# Install Go tooling for linting, debugging, and fuzzing
GO_TOOLS=(
  golang.org/x/tools/cmd/goimports
  honnef.co/go/tools/cmd/staticcheck
  github.com/golangci/golangci-lint/cmd/golangci-lint
  github.com/go-delve/delve/cmd/dlv
  github.com/google/pprof
  github.com/google/gofuzz
)
for tool in "${GO_TOOLS[@]}"; do
  if ! GO111MODULE=on go install "$tool@latest" >/dev/null 2>&1; then
    echo "Go install failed: $tool" >> "$LOG_FILE"
  fi
done


# Python environment for repo
python3 -m venv /opt/go-nfsd-venv
source /opt/go-nfsd-venv/bin/activate
if ! pip install --upgrade pip >/dev/null 2>&1; then
  echo "PIP upgrade failed" >> "$LOG_FILE"
fi
if ! pip install -r requirements.txt >/dev/null 2>&1; then
  echo "PIP install failed: requirements.txt" >> "$LOG_FILE"
fi
for pkg in pytest pytest-xdist pexpect; do
  pip_install "$pkg"
done

deactivate

# clean up
apt-get clean
rm -rf /var/lib/apt/lists/*

if [ -s "$LOG_FILE" ]; then
  echo "Some packages failed to install. See $LOG_FILE for details." >&2
  cat "$LOG_FILE" >&2
fi

exit 0
