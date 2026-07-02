#!/usr/bin/env bash
# ==========================================================
# LCA Agent - Linux 一键部署脚本
# 将 logcollect 注册为 systemd 服务，异常退出自动重启
#
# 使用方式:
#   sudo bash install-agent-linux.sh
# ==========================================================
set -euo pipefail

# ---------- 配置区域（请根据实际修改） ----------
AGENT_ID="${AGENT_ID:-agent-001}"
ADMIN_URL="${ADMIN_URL:-http://192.168.1.100:8080}"
API_URL="${API_URL:-http://192.168.1.100:8086}"
BATCH_SIZE="${BATCH_SIZE:-50}"
FLUSH_SECONDS="${FLUSH_SECONDS:-5}"
PUSH_CONCURRENCY="${PUSH_CONCURRENCY:-5}"
HOSTNAME_OVERRIDE="${HOSTNAME_OVERRIDE:-$(hostname)}"

# ---------- 安装路径 ----------
INSTALL_DIR="/opt/lca-agent"
SERVICE_NAME="lca-agent"
BINARY_NAME="logcollect"
CONFIG_FILE="${INSTALL_DIR}/config.yaml"

# ---------- 检查权限 ----------
if [ "$(id -u)" -ne 0 ]; then
    echo "ERROR: 请使用 sudo 或 root 用户执行此脚本"
    exit 1
fi

echo "========================================"
echo "  LCA Agent - Linux 部署"
echo "========================================"
echo "  Agent ID : ${AGENT_ID}"
echo "  Admin    : ${ADMIN_URL}"
echo "  API      : ${API_URL}"
echo "  Hostname : ${HOSTNAME_OVERRIDE}"
echo "  安装目录 : ${INSTALL_DIR}"
echo "========================================"

# ---------- 1. 创建安装目录 ----------
mkdir -p "${INSTALL_DIR}/logs" "${INSTALL_DIR}/dead_letters"

# ---------- 2. 复制二进制文件 ----------
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY_SRC="${SCRIPT_DIR}/../release/bin/${BINARY_NAME}"

if [ -f "${SCRIPT_DIR}/${BINARY_NAME}" ]; then
    BINARY_SRC="${SCRIPT_DIR}/${BINARY_NAME}"
elif [ ! -f "${BINARY_SRC}" ]; then
    echo "ERROR: 未找到 ${BINARY_NAME} 二进制文件"
    echo "  请将编译好的 logcollect 放到当前目录或 release/bin/ 下"
    exit 1
fi

cp "${BINARY_SRC}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
echo ">> 已安装二进制文件到 ${INSTALL_DIR}/${BINARY_NAME}"

# ---------- 3. 生成配置文件 ----------
cat > "${CONFIG_FILE}" <<EOF
# LCA Agent 配置文件（由 install-agent-linux.sh 生成）
api_server: "${API_URL}"
admin_server: "${ADMIN_URL}"
agent_id: "${AGENT_ID}"
batch_size: ${BATCH_SIZE}
flush_seconds: ${FLUSH_SECONDS}
push_concurrency: ${PUSH_CONCURRENCY}
hostname: "${HOSTNAME_OVERRIDE}"
EOF

echo ">> 已生成配置文件 ${CONFIG_FILE}"

# ---------- 4. 创建 offsets.json ----------
[ -f "${INSTALL_DIR}/offsets.json" ] || echo '[]' > "${INSTALL_DIR}/offsets.json"

# ---------- 5. 创建 systemd 服务 ----------
cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=LCA Log Collection Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/${BINARY_NAME} -config ${CONFIG_FILE}
WorkingDirectory=${INSTALL_DIR}
Restart=always
RestartSec=5
StartLimitIntervalSec=60
StartLimitBurst=10
StandardOutput=append:${INSTALL_DIR}/logs/agent.log
StandardError=append:${INSTALL_DIR}/logs/agent.log

# 资源限制
LimitNOFILE=65535
MemoryMax=512M

[Install]
WantedBy=multi-user.target
EOF

echo ">> 已创建 systemd 服务: ${SERVICE_NAME}"

# ---------- 6. 启动服务 ----------
systemctl daemon-reload
systemctl enable ${SERVICE_NAME}
systemctl restart ${SERVICE_NAME}

echo ""
echo "======== 部署完成 ========"
echo "  查看状态: systemctl status ${SERVICE_NAME}"
echo "  查看日志: journalctl -u ${SERVICE_NAME} -f"
echo "  停止服务: systemctl stop ${SERVICE_NAME}"
echo "  卸载服务: systemctl disable ${SERVICE_NAME} && rm /etc/systemd/system/${SERVICE_NAME}.service"
echo ""
echo "  配置文件: ${CONFIG_FILE}"
echo "  运行日志: ${INSTALL_DIR}/logs/agent.log"
echo "==============================="
