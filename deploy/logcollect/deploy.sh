#!/usr/bin/env bash
# ==========================================================
# logcollect Agent - 一键部署脚本
#
# 使用方式:
#   bash deploy.sh              # 首次部署
#   bash deploy.sh restart      # 更新配置后重启
# ==========================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# ---------- 1. 检查 .env ----------
if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        cp .env.example .env
        echo ">> 已从 .env.example 创建 .env，请修改后重新执行"
        echo "   AGENT_ID: 每台机器必须不同"
        echo "   ADMIN_URL: 管理后台地址"
        echo "   API_URL:   日志推送地址"
        exit 1
    else
        echo "ERROR: .env 和 .env.example 都不存在"
        exit 1
    fi
fi

# 加载 .env
set -a
source .env
set +a

# 校验必填项
if [ -z "${AGENT_ID:-}" ] || [ -z "${ADMIN_URL:-}" ] || [ -z "${API_URL:-}" ] || [ -z "${HOSTNAME:-}" ]; then
    echo "ERROR: .env 中 AGENT_ID / ADMIN_URL / API_URL / HOSTNAME 为必填项"
    exit 1
fi

echo "======== logcollect Agent 部署 ========"
echo "  Agent ID : ${AGENT_ID}"
echo "  Admin    : ${ADMIN_URL}"
echo "  API      : ${API_URL}"
echo "========================================"

# ---------- 2. 生成运行时配置 ----------
export API_URL ADMIN_URL AGENT_ID BATCH_SIZE FLUSH_SECONDS PUSH_CONCURRENCY HOSTNAME
envsubst < config.template.yaml > config.yaml
echo ">> 已生成 config.yaml"

# ---------- 3. 检查二进制文件（按本机架构） ----------
BINARY="logcollect"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
case "$(uname -m)" in
    x86_64|amd64) GOARCH=amd64 ;;
    aarch64|arm64) GOARCH=arm64 ;;
    *) echo "ERROR: 不支持的架构 $(uname -m)"; exit 1 ;;
esac

if [ ! -f "$BINARY" ]; then
    for candidate in \
        "${ROOT_DIR}/release/bin/${GOARCH}/${BINARY}" \
        "${ROOT_DIR}/release/bin/${BINARY}"; do
        if [ -f "$candidate" ]; then
            cp "$candidate" "$BINARY"
            echo ">> 已从 $candidate 复制二进制 (${GOARCH})"
            break
        fi
    done
fi
if [ ! -f "$BINARY" ]; then
    echo "ERROR: 未找到 logcollect 二进制文件（架构 ${GOARCH}）"
    echo "  请先执行: cd $ROOT_DIR && make release-native"
    exit 1
fi

# ---------- 4. 创建运行时目录/文件 ----------
mkdir -p logs dead_letters
# 确保 offsets.json 存在（Docker挂载文件时不存在会被当成目录）
[ -f offsets.json ] || echo '[]' > offsets.json

# ---------- 5. 启动 ----------
if [ "${1:-}" = "restart" ]; then
    echo ">> 重新部署..."
    docker compose down
    docker compose up -d --build --force-recreate
else
    echo ">> 启动容器..."
    docker compose up -d --build
fi

echo ""
echo "======== 部署完成 ========"
echo "  查看日志: docker compose logs -f --tail=50"
echo "  停止服务: docker compose down"
echo "  重启服务: bash deploy.sh restart"
