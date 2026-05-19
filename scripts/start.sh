#!/bin/bash
# LCA 系统一键启动脚本 (本地开发模式)
# 用法: ./scripts/start.sh [infra|backend|frontend|all|stop]
#   infra   - 仅启动基础设施 (MySQL/Redis/Kafka/ES)
#   backend - 仅启动后端服务 (admin/apiserver/logtransfer)
#   frontend- 仅启动前端开发服务器
#   all     - 启动全部 (默认)
#   stop    - 停止所有服务

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="$ROOT_DIR/logs"
PID_DIR="$ROOT_DIR/logs"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; }

# 创建必要目录
mkdir -p "$LOG_DIR"

# ==================== 基础设施 ====================

start_infra() {
    info "启动基础设施 (MySQL/Redis/Kafka/ZooKeeper/Elasticsearch)..."
    cd "$ROOT_DIR"
    docker compose up -d mysql redis kafka zookeeper elasticsearch

    info "等待基础设施就绪..."
    local retries=0
    local max_retries=30
    while [ $retries -lt $max_retries ]; do
        if docker compose exec mysql mysqladmin ping -h localhost -plca2024 --silent 2>/dev/null && \
           docker compose exec redis redis-cli -a lca2024 ping 2>/dev/null | grep -q PONG && \
           curl -sf http://127.0.0.1:9200/_cluster/health > /dev/null 2>&1; then
            info "基础设施已就绪!"
            return 0
        fi
        retries=$((retries + 1))
        printf "  等待中... (%d/%d)\r" "$retries" "$max_retries"
        sleep 2
    done
    warn "基础设施可能未完全就绪，请检查 docker compose ps"
}

# ==================== 后端服务 ====================

start_backend_service() {
    local name=$1
    local cmd=$2
    local config=$3
    local pidfile="$PID_DIR/${name}.pid"

    if [ -f "$pidfile" ]; then
        local old_pid
        old_pid=$(cat "$pidfile")
        if kill -0 "$old_pid" 2>/dev/null; then
            warn "$name 已在运行 (PID: $old_pid)，跳过"
            return 0
        else
            rm -f "$pidfile"
        fi
    fi

    info "启动 $name ..."
    cd "$ROOT_DIR"
    go run "$cmd" -config "$config" > "$LOG_DIR/${name}.log" 2>&1 &
    local pid=$!
    echo "$pid" > "$pidfile"

    # 等待服务启动
    local retries=0
    local max_retries=15
    while [ $retries -lt $max_retries ]; do
        if ! kill -0 "$pid" 2>/dev/null; then
            error "$name 启动失败! 查看 $LOG_DIR/${name}.log"
            cat "$LOG_DIR/${name}.log"
            return 1
        fi
        # 检查日志中是否有 "Listening" 或 "starting" 关键字
        if grep -qiE "(Listening|starting|Server starting)" "$LOG_DIR/${name}.log" 2>/dev/null; then
            info "$name 已启动 (PID: $pid)"
            return 0
        fi
        retries=$((retries + 1))
        sleep 1
    done
    info "$name 启动中... (PID: $pid)，详情见 $LOG_DIR/${name}.log"
}

start_backend() {
    info "启动后端服务..."
    start_backend_service "admin"      "./cmd/admin"      "configs/admin.yaml"
    start_backend_service "apiserver"  "./cmd/apiserver"  "configs/apiserver.yaml"
    start_backend_service "logtransfer" "./cmd/logtransfer" "configs/logtransfer.yaml"
    info "后端服务启动完成"
}

# ==================== 前端 ====================

start_frontend() {
    local pidfile="$PID_DIR/frontend.pid"

    if [ -f "$pidfile" ]; then
        local old_pid
        old_pid=$(cat "$pidfile")
        if kill -0 "$old_pid" 2>/dev/null; then
            warn "前端已在运行 (PID: $old_pid)，跳过"
            return 0
        else
            rm -f "$pidfile"
        fi
    fi

    info "启动前端开发服务器..."
    cd "$ROOT_DIR/web"

    # 检查 node_modules
    if [ ! -d "node_modules" ]; then
        info "安装前端依赖..."
        npm install
    fi

    npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
    local pid=$!
    echo "$pid" > "$pidfile"

    local retries=0
    local max_retries=15
    while [ $retries -lt $max_retries ]; do
        if ! kill -0 "$pid" 2>/dev/null; then
            error "前端启动失败! 查看 $LOG_DIR/frontend.log"
            cat "$LOG_DIR/frontend.log"
            return 1
        fi
        if grep -q "Local:" "$LOG_DIR/frontend.log" 2>/dev/null; then
            info "前端已启动 (PID: $pid)"
            return 0
        fi
        retries=$((retries + 1))
        sleep 1
    done
    info "前端启动中... (PID: $pid)，详情见 $LOG_DIR/frontend.log"
}

# ==================== 停止服务 ====================

stop_all() {
    info "停止后端和前端服务..."
    for name in admin apiserver logtransfer frontend; do
        local pidfile="$PID_DIR/${name}.pid"
        if [ -f "$pidfile" ]; then
            local pid
            pid=$(cat "$pidfile")
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid" 2>/dev/null && info "$name (PID: $pid) 已停止"
            fi
            rm -f "$pidfile"
        fi
    done

    info "停止基础设施..."
    cd "$ROOT_DIR"
    docker compose down 2>/dev/null || true

    info "所有服务已停止"
}

# ==================== 状态检查 ====================

show_status() {
    echo ""
    echo -e "${CYAN}========== LCA 服务状态 ==========${NC}"
    for name in admin apiserver logtransfer frontend; do
        local pidfile="$PID_DIR/${name}.pid"
        if [ -f "$pidfile" ]; then
            local pid
            pid=$(cat "$pidfile")
            if kill -0 "$pid" 2>/dev/null; then
                echo -e "  $name:  ${GREEN}运行中${NC} (PID: $pid)"
            else
                echo -e "  $name:  ${RED}已停止${NC} (PID文件过期)"
            fi
        else
            echo -e "  $name:  ${YELLOW}未启动${NC}"
        fi
    done

    echo ""
    echo -e "  基础设施容器:"
    cd "$ROOT_DIR"
    docker compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null | sed 's/^/    /' || echo "    无法获取"

    echo ""
    echo -e "${CYAN}===================================${NC}"
    echo ""
    echo -e "  前端访问: ${GREEN}http://localhost:3000${NC}"
    echo -e "  后端API:  ${GREEN}http://localhost:8080${NC}"
    echo ""
}

# ==================== 主入口 ====================

ACTION="${1:-all}"

case "$ACTION" in
    infra)
        start_infra
        ;;
    backend)
        start_backend
        ;;
    frontend)
        start_frontend
        ;;
    all)
        start_infra
        echo ""
        start_backend
        echo ""
        start_frontend
        echo ""
        show_status
        ;;
    stop)
        stop_all
        ;;
    status)
        show_status
        ;;
    *)
        echo "用法: $0 {infra|backend|frontend|all|stop|status}"
        echo ""
        echo "  infra    - 仅启动基础设施 (MySQL/Redis/Kafka/ES)"
        echo "  backend  - 仅启动后端服务 (admin/apiserver/logtransfer)"
        echo "  frontend - 仅启动前端开发服务器"
        echo "  all      - 启动全部 (默认)"
        echo "  stop     - 停止所有服务"
        echo "  status   - 查看服务状态"
        exit 1
        ;;
esac
