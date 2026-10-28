#!/usr/bin/env bash
# 部署前准备：安装 Docker（已有则跳过），并把 vm.max_map_count 调到 ES 要求的 262144。
# 由 make docker-up / make docker-up-ha 自动调用。
set -euo pipefail

MIN_MAP_COUNT=262144

say() { printf '%s\n' "$*"; }

run_root() {
  if [[ "$(id -u)" -eq 0 ]]; then
    "$@"
  else
    sudo "$@"
  fi
}

need_root() {
  if [[ "$(id -u)" -eq 0 ]]; then
    return 0
  fi
  if command -v sudo >/dev/null 2>&1; then
    return 0
  fi
  say "需要 root 或 sudo 才能安装 Docker / 修改 vm.max_map_count" >&2
  exit 1
}

ensure_docker() {
  if command -v docker >/dev/null 2>&1; then
    say "[prepare] Docker 已安装，跳过安装"
    docker --version
  else
    need_root
    if ! command -v curl >/dev/null 2>&1; then
      say "[prepare] 安装 curl"
      if command -v apt-get >/dev/null 2>&1; then
        run_root apt-get update
        run_root apt-get install -y curl ca-certificates
      elif command -v dnf >/dev/null 2>&1; then
        run_root dnf install -y curl ca-certificates
      elif command -v yum >/dev/null 2>&1; then
        run_root yum install -y curl ca-certificates
      else
        say "未找到 curl，请先安装 curl 后再执行" >&2
        exit 1
      fi
    fi
    say "[prepare] 安装 Docker"
    curl -fsSL https://get.docker.com | run_root bash
    docker --version
  fi

  if command -v systemctl >/dev/null 2>&1; then
    need_root
    run_root systemctl enable --now docker
  fi

  if ! docker info >/dev/null 2>&1; then
    if command -v sudo >/dev/null 2>&1 && sudo docker info >/dev/null 2>&1; then
      say "Docker 已安装并启动，但当前用户还不能访问 Docker。" >&2
      say "请重新登录，或执行: sudo usermod -aG docker \"\$USER\" 后重新登录，再执行 make。" >&2
      exit 1
    fi
    say "Docker 未能启动，请检查 systemctl status docker" >&2
    exit 1
  fi

  if ! docker compose version >/dev/null 2>&1; then
    say "未找到 docker compose 插件。请重新执行官方安装脚本，或安装 docker-compose-plugin。" >&2
    exit 1
  fi
}

ensure_map_count() {
  local current
  current="$(sysctl -n vm.max_map_count 2>/dev/null || echo 0)"
  if [[ "$current" =~ ^[0-9]+$ ]] && [[ "$current" -ge "$MIN_MAP_COUNT" ]]; then
    say "[prepare] vm.max_map_count 已是 ${current}，跳过"
    sysctl vm.max_map_count
    return 0
  fi

  need_root
  say "[prepare] 设置 vm.max_map_count=${MIN_MAP_COUNT}（Elasticsearch 必需）"
  run_root touch /etc/sysctl.conf
  run_root bash -c "
    tmp=\$(mktemp)
    if [[ -f /etc/sysctl.conf ]]; then
      grep -v '^vm.max_map_count=' /etc/sysctl.conf > \"\$tmp\" || true
    fi
    echo 'vm.max_map_count=${MIN_MAP_COUNT}' >> \"\$tmp\"
    cat \"\$tmp\" > /etc/sysctl.conf
    rm -f \"\$tmp\"
  "
  run_root sysctl -w "vm.max_map_count=${MIN_MAP_COUNT}" >/dev/null
  run_root sysctl -p >/dev/null || true
  current="$(sysctl -n vm.max_map_count)"
  sysctl vm.max_map_count
  if [[ "$current" -lt "$MIN_MAP_COUNT" ]]; then
    say "vm.max_map_count=${current}，仍低于 ${MIN_MAP_COUNT}。Elasticsearch 会反复重启。" >&2
    exit 1
  fi
}

os="$(uname -s)"
case "$os" in
  Linux)
    ensure_docker
    ensure_map_count
    ;;
  Darwin)
    if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
      say "macOS 请先安装并启动 Docker Desktop，然后再执行 make。" >&2
      say "https://docs.docker.com/desktop/setup/install/mac-install/" >&2
      exit 1
    fi
    say "[prepare] macOS 使用 Docker Desktop，跳过 get.docker.com 与 vm.max_map_count"
    docker --version
    if ! docker compose version >/dev/null 2>&1; then
      say "未找到 docker compose。请在 Docker Desktop 中启用 Compose。" >&2
      exit 1
    fi
    ;;
  *)
    say "未支持的系统: $os。请在 Linux 服务器上执行 make docker-up / make docker-up-ha。" >&2
    exit 1
    ;;
esac
