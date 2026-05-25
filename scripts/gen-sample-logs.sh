#!/bin/bash
# 生成 LCA 演示用样本日志
# 输出目录: /tmp/lca-samples/{nginx,backend,security}
set -e

OUT_DIR="${1:-/tmp/lca-samples}"
mkdir -p "$OUT_DIR/nginx" "$OUT_DIR/backend" "$OUT_DIR/security"

NGINX_LOG="$OUT_DIR/nginx/access.log"
BACKEND_LOG="$OUT_DIR/backend/app.log"
SECURITY_LOG="$OUT_DIR/security/audit.log"

> "$NGINX_LOG"
> "$BACKEND_LOG"
> "$SECURITY_LOG"

# 24 小时的数据(过去24h, 每分钟若干条)
TOTAL_MINUTES=1440
NOW_TS=$(date +%s)

# IP 池 / UA / URL
IPS=("203.0.113.10" "203.0.113.22" "203.0.113.45" "198.51.100.7" "198.51.100.88" "192.0.2.5" "192.0.2.99" "10.0.0.1" "10.0.0.42" "172.16.5.30")
UAS=("Mozilla/5.0 (Windows NT 10.0) Chrome/124.0" "Mozilla/5.0 (Macintosh) Safari/17.4" "Mozilla/5.0 (iPhone) Mobile Safari" "curl/7.88.1" "PostmanRuntime/7.36" "python-requests/2.31" "Googlebot/2.1")
PATHS=("/api/v1/login" "/api/v1/users" "/api/v1/orders" "/api/v1/orders/123" "/api/v1/products" "/api/v1/payments" "/api/v1/logout" "/static/app.js" "/static/main.css" "/health")
METHODS=("GET" "GET" "GET" "POST" "POST" "PUT" "DELETE")
STATUSES=("200" "200" "200" "200" "201" "302" "400" "401" "403" "404" "500" "502" "504")
LEVELS=("INFO" "INFO" "INFO" "INFO" "WARN" "ERROR" "DEBUG")
USERS=("admin" "alice" "bob" "charlie" "diana" "eve" "frank")
SERVICES=("auth" "user" "order" "payment" "product" "report")
ACTIONS=("login_success" "login_failed" "logout" "create_user" "delete_user" "update_role" "export_data" "view_secret" "config_change" "failed_attempt")

rand_pick() {
  local arr=("$@")
  echo "${arr[$RANDOM % ${#arr[@]}]}"
}

for ((m=TOTAL_MINUTES; m>=0; m--)); do
  TS=$(( NOW_TS - m * 60 ))
  TIME_STR=$(date -r "$TS" "+%Y-%m-%dT%H:%M:%S+08:00" 2>/dev/null || date -d "@$TS" "+%Y-%m-%dT%H:%M:%S+08:00")
  NGINX_TIME=$(date -r "$TS" "+%d/%b/%Y:%H:%M:%S +0800" 2>/dev/null || date -d "@$TS" "+%d/%b/%Y:%H:%M:%S +0800")

  # 每分钟 3-8 条 nginx 日志
  REQS=$((3 + RANDOM % 6))
  for ((i=0; i<REQS; i++)); do
    IP=$(rand_pick "${IPS[@]}")
    UA=$(rand_pick "${UAS[@]}")
    PATH_=$(rand_pick "${PATHS[@]}")
    METHOD=$(rand_pick "${METHODS[@]}")
    STATUS=$(rand_pick "${STATUSES[@]}")
    SIZE=$((100 + RANDOM % 8000))
    RT=$(awk -v r=$RANDOM 'BEGIN{printf "%.3f", (r%2000)/1000}')
    echo "$IP - - [$NGINX_TIME] \"$METHOD $PATH_ HTTP/1.1\" $STATUS $SIZE \"-\" \"$UA\" $RT" >> "$NGINX_LOG"
  done

  # 每分钟 2-5 条 backend JSON 日志
  REQS=$((2 + RANDOM % 4))
  for ((i=0; i<REQS; i++)); do
    LEVEL=$(rand_pick "${LEVELS[@]}")
    USER=$(rand_pick "${USERS[@]}")
    SVC=$(rand_pick "${SERVICES[@]}")
    LATENCY=$((10 + RANDOM % 1500))
    TRACE=$(printf '%08x%08x' $RANDOM$RANDOM $RANDOM$RANDOM)
    if [ "$LEVEL" = "ERROR" ]; then
      MSG="database query failed: connection timeout"
    elif [ "$LEVEL" = "WARN" ]; then
      MSG="slow query detected, latency=${LATENCY}ms"
    else
      MSG="request handled successfully"
    fi
    echo "{\"time\":\"$TIME_STR\",\"level\":\"$LEVEL\",\"service\":\"$SVC\",\"user\":\"$USER\",\"latency_ms\":$LATENCY,\"trace_id\":\"$TRACE\",\"msg\":\"$MSG\"}" >> "$BACKEND_LOG"
  done

  # 每分钟 1-2 条 security 审计日志
  REQS=$((1 + RANDOM % 2))
  for ((i=0; i<REQS; i++)); do
    USER=$(rand_pick "${USERS[@]}")
    ACTION=$(rand_pick "${ACTIONS[@]}")
    IP=$(rand_pick "${IPS[@]}")
    if [[ "$ACTION" == *"failed"* ]] || [[ "$ACTION" == "delete_user" ]] || [[ "$ACTION" == "view_secret" ]]; then
      RESULT="failure"
      RISK="high"
    else
      RESULT="success"
      RISK="low"
    fi
    echo "{\"time\":\"$TIME_STR\",\"user\":\"$USER\",\"action\":\"$ACTION\",\"client_ip\":\"$IP\",\"result\":\"$RESULT\",\"risk\":\"$RISK\"}" >> "$SECURITY_LOG"
  done
done

echo "✅ 样本日志已生成:"
wc -l "$NGINX_LOG" "$BACKEND_LOG" "$SECURITY_LOG"
echo ""
echo "📁 路径:"
echo "  Nginx:    $NGINX_LOG"
echo "  Backend:  $BACKEND_LOG"
echo "  Security: $SECURITY_LOG"
