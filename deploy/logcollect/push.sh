#!/usr/bin/env bash
# ==========================================================
# logcollect Agent - 推送到远端服务器
#
# 使用方式:
#   bash push.sh 192.168.1.10                    # 默认 root 用户，/opt/lca-agent 目录
#   bash push.sh ubuntu@192.168.1.10 /app/agent  # 指定用户和目录
#   bash push.sh hosts.txt                       # 从文件批量推送
#
# hosts.txt 格式（每行一台）:
#   root@192.168.1.10
#   root@192.168.1.11 /data/lca-agent
#   ubuntu@192.168.1.12
# ==========================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"
BINARY_NAME="logcollect"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# ---------- 颜色输出 ----------
RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'

# ---------- 架构检测 ----------
map_goarch() {
    case "$1" in
        x86_64|amd64) echo amd64 ;;
        aarch64|arm64) echo arm64 ;;
        *) return 1 ;;
    esac
}

# ---------- 确保目标架构二进制已编译 ----------
ensure_binary() {
    local arch="${1:-}"
    if [ -z "$arch" ]; then
        arch="$(bash "$ROOT_DIR/scripts/goarch.sh")"
    fi
    local src="$ROOT_DIR/release/bin/${arch}/${BINARY_NAME}"
    if [ ! -f "$src" ]; then
        echo -e "${CYAN}>> 编译 Linux/${arch} ...${NC}"
        mkdir -p "$ROOT_DIR/release/bin/${arch}"
        (cd "$ROOT_DIR" && CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -o "release/bin/${arch}/${BINARY_NAME}" ./cmd/logcollect)
        echo -e "${GREEN}>> 编译完成${NC}"
    fi
    cp -f "$src" "$SCRIPT_DIR/$BINARY_NAME"
}

# ---------- SSH 参数 ----------
SSH_OPTS="-o StrictHostKeyChecking=no -o ConnectTimeout=5"

# ---------- 推送到单台 ----------
push_single() {
    local target="$1"
    local remote_dir="${2:-/opt/lca-agent}"

    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  推送至: $target → $remote_dir${NC}"
    echo -e "${CYAN}========================================${NC}"

    # 探测远端架构并准备对应二进制
    local remote_m
    remote_m="$(ssh $SSH_OPTS "$target" "uname -m")"
    local remote_arch
    if ! remote_arch="$(map_goarch "$remote_m")"; then
        echo -e "${RED}>> 远端架构不支持: $remote_m${NC}" >&2
        return 1
    fi
    echo -e "${CYAN}>> 远端架构: $remote_m → $remote_arch${NC}"
    ensure_binary "$remote_arch"

    # 1. 创建远端目录
    ssh $SSH_OPTS "$target" "mkdir -p $remote_dir"

    # 2. 同步部署文件（排除运行时生成的文件）
    rsync -avz --delete \
        --exclude='logs/' \
        --exclude='dead_letters/' \
        --exclude='.env' \
        --exclude='config.yaml' \
        -e "ssh $SSH_OPTS" \
        ./ "$target:$remote_dir/"

    # 3. 复制二进制
    scp $SSH_OPTS "$BINARY_NAME" "$target:$remote_dir/"

    # 4. 生成远端 .env（如果不存在）
    ssh $SSH_OPTS "$target" "cd $remote_dir && [ -f .env ] || cp .env.example .env"

    echo -e "${GREEN}>> 推送完成: $target${NC}"
}

# ---------- 推送到多台 ----------
push_multi() {
    local file="$1"
    local fail=0

    while IFS= read -r line || [ -n "$line" ]; do
        # 跳过空行和注释
        [[ -z "$line" || "$line" =~ ^# ]] && continue
        read -r host dir <<< "$line"
        if push_single "$host" "${dir:-/opt/lca-agent}"; then
            echo -e "${GREEN}✅ $host ${NC}"
        else
            echo -e "${RED}❌ $host 推送失败${NC}"
            fail=1
        fi
    done < "$file"
    return $fail
}

# ---------- 主流程 ----------
if [ $# -eq 0 ]; then
    echo "用法:"
    echo "  bash push.sh root@192.168.1.10              # 推送到单台"
    echo "  bash push.sh root@192.168.1.10 /app/agent   # 指定远端目录"
    echo "  bash push.sh hosts.txt                       # 从文件批量推送"
    exit 1
fi

ARG="$1"

if [ -f "$ARG" ]; then
    # 从文件读取
    push_multi "$ARG"
    echo ""
    echo -e "${GREEN}======== 批量推送完成 ========${NC}"
    echo "  在每台服务器上执行远端部署:"
    echo "  ssh <host> 'cd <dir> && bash deploy.sh'"
else
    # 单台推送
    TARGET="$ARG"
    REMOTE_DIR="${2:-/opt/lca-agent}"
    push_single "$TARGET" "$REMOTE_DIR"
    echo ""
    echo -e "${GREEN}======== 推送完成 ========${NC}"
    echo "  远端部署:"
    echo "  ssh $TARGET 'cd $REMOTE_DIR && vim .env && bash deploy.sh'"
fi
