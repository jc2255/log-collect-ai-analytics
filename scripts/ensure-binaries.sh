#!/usr/bin/env bash
# 确保 Linux 二进制已在 release/bin/<arch>/
# - 源码仓：缺失时交叉编译
# - 交付包：没有 cmd/，只检查预编译文件，绝不 go build
# 用法:
#   bash scripts/ensure-binaries.sh           # 本机架构
#   bash scripts/ensure-binaries.sh amd64     # 指定架构
#   bash scripts/ensure-binaries.sh amd64 arm64
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
NATIVE="$(bash "$ROOT_DIR/scripts/goarch.sh")"
if [[ $# -gt 0 ]]; then
  ARCHES=("$@")
else
  ARCHES=("$NATIVE")
fi
APPS=(admin apiserver logcollect logtransfer syslog)
HAS_SOURCE=0
if [[ -d "$ROOT_DIR/cmd/admin" && -f "$ROOT_DIR/go.mod" ]]; then
  HAS_SOURCE=1
fi

cd "$ROOT_DIR"
mkdir -p release/bin

for GOARCH in "${ARCHES[@]}"; do
  case "$GOARCH" in
    amd64|arm64) ;;
    *) echo "不支持的架构: $GOARCH" >&2; exit 1 ;;
  esac
  OUT_DIR="release/bin/${GOARCH}"
  mkdir -p "$OUT_DIR"

  missing=()
  for app in "${APPS[@]}"; do
    if [[ ! -f "$OUT_DIR/$app" ]]; then
      missing+=("$app")
    fi
  done

  if [[ ${#missing[@]} -eq 0 ]]; then
    echo "[ensure-binaries] 已存在 $GOARCH 二进制: $OUT_DIR"
    continue
  fi

  if [[ "$HAS_SOURCE" -eq 0 ]]; then
    echo "[ensure-binaries] 交付包缺少 linux/${GOARCH}: ${missing[*]}" >&2
    echo "请使用完整安装包（含 release/bin/amd64 与 release/bin/arm64），不要在客户机编译。" >&2
    exit 1
  fi

  echo "[ensure-binaries] 交叉编译 linux/${GOARCH} ..."
  for app in "${APPS[@]}"; do
    CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" go build -o "$OUT_DIR/$app" "./cmd/$app"
    echo "  built: $OUT_DIR/$app"
  done
done

# 本机架构同步到 release/bin/<app>，兼容旧脚本 / Dockerfile 回退路径
for app in "${APPS[@]}"; do
  if [[ -f "release/bin/${NATIVE}/$app" ]]; then
    cp -f "release/bin/${NATIVE}/$app" "release/bin/$app"
  fi
done
echo "[ensure-binaries] 完成 (arches: ${ARCHES[*]})"
