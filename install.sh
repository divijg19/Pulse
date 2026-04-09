#!/usr/bin/env bash
set -euo pipefail

REPO="divijg19/Pulse"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi

latest_json="$(curl -fsSL "$API_URL")"
version="$(printf '%s' "$latest_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"

if [ -z "$version" ]; then
  echo "error: failed to determine latest release tag from ${API_URL}" >&2
  exit 1
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch_raw="$(uname -m)"

case "$arch_raw" in
  x86_64|amd64)
    arch="amd64"
    ;;
  arm64|aarch64)
    arch="arm64"
    ;;
  *)
    echo "error: unsupported architecture: ${arch_raw}" >&2
    exit 1
    ;;
esac

case "$os" in
  linux)
    if [ "$arch" != "amd64" ]; then
      echo "error: linux ${arch} is not published for Pulse releases" >&2
      exit 1
    fi
    suffix="linux-amd64"
    ;;
  darwin)
    if [ "$arch" = "amd64" ]; then
      suffix="mac-amd64"
    else
      suffix="mac-arm64"
    fi
    ;;
  msys*|mingw*|cygwin*|windows_nt)
    if [ "$arch" != "amd64" ]; then
      echo "error: windows ${arch} is not published for Pulse releases" >&2
      exit 1
    fi
    suffix="windows-amd64.exe"
    ;;
  *)
    echo "error: unsupported operating system: ${os}" >&2
    exit 1
    ;;
esac

file_name="pulse-${version}-${suffix}"
download_url="https://github.com/${REPO}/releases/download/${version}/${file_name}"

tmp_file="$(mktemp)"
cleanup() {
  rm -f "$tmp_file"
}
trap cleanup EXIT

echo "Downloading ${file_name} ..."
curl -fL "$download_url" -o "$tmp_file"

target_dir="/usr/local/bin"
target_path="${target_dir}/pulse"

install_with_sudo=false
if [ ! -w "$target_dir" ]; then
  if command -v sudo >/dev/null 2>&1; then
    install_with_sudo=true
  else
    target_dir="${HOME}/.local/bin"
    target_path="${target_dir}/pulse"
    mkdir -p "$target_dir"
  fi
fi

if [ "$install_with_sudo" = true ]; then
  sudo cp "$tmp_file" "$target_path"
  sudo chmod 0755 "$target_path"
else
  cp "$tmp_file" "$target_path"
  chmod 0755 "$target_path"
fi

echo "Installed pulse to ${target_path}"

case ":${PATH}:" in
  *":${target_dir}:"*) ;;
  *)
    if [ "$target_dir" = "${HOME}/.local/bin" ]; then
      echo "Note: ${target_dir} is not in PATH."
      echo "Add this to your shell profile:"
      echo "  export PATH=\"${HOME}/.local/bin:\$PATH\""
    fi
    ;;
esac

echo "Run: pulse"
