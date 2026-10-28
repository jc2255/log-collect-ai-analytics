#!/usr/bin/env bash
# 将 uname -m 映射为 Go/Docker 架构名（amd64 | arm64）
# 用法:
#   source scripts/goarch.sh && echo "$GOARCH"
#   ARCH=$(bash scripts/goarch.sh)
set -euo pipefail

detect_goarch() {
  local m="${1:-$(uname -m)}"
  case "$m" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    *)
      echo "不支持的 CPU 架构: $m（仅支持 x86_64/amd64 与 aarch64/arm64）" >&2
      return 1
      ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  detect_goarch "$@"
else
  GOARCH="$(detect_goarch)"
  export GOARCH
fi
