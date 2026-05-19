#!/bin/bash
# =============================================================
# LCA 日志生成测试脚本
# 用途：模拟各种软件/编程语言产生的日志，验证 Agent 采集与 AI 告警
# 使用：bash scripts/gen_test_logs.sh [--error-burst] [--interval N]
#   --error-burst  立即生成大量 ERROR 日志，触发 AI 告警
#   --interval N   日志写入间隔秒数（默认 1 秒）
# =============================================================

LOG_DIR="/tmp/lca-test-logs"
mkdir -p "$LOG_DIR"/{nginx,java,python,golang,mysql,nodejs,php}

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

INTERVAL=1
ERROR_BURST=false

for arg in "$@"; do
  case $arg in
    --error-burst) ERROR_BURST=true ;;
    --interval) INTERVAL="$2"; shift ;;
  esac
done

echo -e "${CYAN}================================================${NC}"
echo -e "${CYAN}  LCA 日志采集 & AI 告警 测试日志生成器${NC}"
echo -e "${CYAN}================================================${NC}"
echo -e "日志目录: ${GREEN}$LOG_DIR${NC}"
echo -e "写入间隔: ${GREEN}${INTERVAL}s${NC}"
echo -e "ERROR触发模式: ${GREEN}$ERROR_BURST${NC}"
echo ""
echo -e "${YELLOW}请在管理后台为每种日志库配置采集任务，路径参考:${NC}"
echo -e "  Nginx 访问日志:  $LOG_DIR/nginx/access.log"
echo -e "  Nginx 错误日志:  $LOG_DIR/nginx/error.log"
echo -e "  Java/Spring:     $LOG_DIR/java/app.log"
echo -e "  Python:          $LOG_DIR/python/app.log"
echo -e "  Go 应用:         $LOG_DIR/golang/app.log"
echo -e "  MySQL 慢查询:    $LOG_DIR/mysql/slow.log"
echo -e "  Node.js:         $LOG_DIR/nodejs/app.log"
echo -e "  PHP:             $LOG_DIR/php/app.log"
echo ""
echo -e "按 ${RED}Ctrl+C${NC} 停止生成"
echo -e "${CYAN}================================================${NC}"
echo ""

# ---- 工具函数 ----
ts()     { date '+%Y-%m-%d %H:%M:%S'; }
ts_iso() { date '+%Y-%m-%dT%H:%M:%S.000+0800'; }
ts_nginx_access() { date '+%d/%b/%Y:%H:%M:%S +0800'; }
ts_nginx_error()  { date '+%Y/%m/%d %H:%M:%S'; }
ts_mysql()        { date '+%Y-%m-%dT%H:%M:%S.%6NZ'; }

METHODS=("GET" "POST" "PUT" "DELETE" "GET" "GET" "GET" "POST")
PATHS=("/api/users" "/api/login" "/api/orders" "/api/products" "/health" "/api/logs" "/api/dashboard" "/api/config" "/static/js/main.js" "/static/css/app.css")
CODES=(200 200 200 201 200 204 301 400 401 403 404 500 502 503)
IPS=("10.0.1.10" "10.0.1.11" "192.168.1.100" "172.16.0.5" "10.10.0.2" "8.8.8.8")
UAS=("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" "curl/7.68.0" "Python-requests/2.28.0" "Go-http-client/1.1")
JAVA_CLASSES=("UserService" "OrderController" "PaymentService" "DatabasePool" "CacheManager" "AuthFilter" "MessageQueue")
PYTHON_MODULES=("app.api.users" "app.services.payment" "app.db.connector" "app.cache.redis" "app.auth.middleware")
GO_PKGS=("handler" "service" "repository" "middleware" "scheduler" "collector")
USERS=("alice" "bob" "charlie" "admin" "system" "api-gateway")
ACTIONS=("login" "logout" "query" "update" "delete" "export" "import")

rand_item() { local arr=("$@"); echo "${arr[$((RANDOM % ${#arr[@]}))]}"; }
rand_int()  { echo $((RANDOM % ($2 - $1) + $1)); }

# ---- Nginx 访问日志 ----
gen_nginx_access() {
  local ip method path code size resp_time ua
  ip=$(rand_item "${IPS[@]}")
  method=$(rand_item "${METHODS[@]}")
  path=$(rand_item "${PATHS[@]}")
  code=$(rand_item "${CODES[@]}")
  size=$(rand_int 100 51200)
  resp_time="0.0$(rand_int 10 999)"
  ua=$(rand_item "${UAS[@]}")
  echo "$ip - - [$(ts_nginx_access)] \"$method $path HTTP/1.1\" $code $size \"http://example.com\" \"$ua\" $resp_time" \
    >> "$LOG_DIR/nginx/access.log"
}

# ---- Nginx 错误日志 ----
gen_nginx_error() {
  local level=$1 msg
  case $level in
    error) msg="connect() failed (111: Connection refused) while connecting to upstream, client: $(rand_item "${IPS[@]}"), server: _, request: \"GET $(rand_item "${PATHS[@]}") HTTP/1.1\", upstream: \"http://127.0.0.1:8080$(rand_item "${PATHS[@]}")\"";;
    warn)  msg="upstream server temporarily disabled while reading response header from upstream";;
    *)     msg="no live upstreams while connecting to upstream";;
  esac
  echo "$(ts_nginx_error) [$level] 1234#0: *$(rand_int 1000 9999) $msg" >> "$LOG_DIR/nginx/error.log"
}

# ---- Java/Spring Boot 日志 ----
gen_java_log() {
  local level=$1
  local cls=$(rand_item "${JAVA_CLASSES[@]}")
  local tid="http-nio-8080-exec-$(rand_int 1 20)"
  local msgs_info=("User login successful: user=$(rand_item "${USERS[@]}")" "Fetched $(rand_int 1 100) records from DB" "Cache hit ratio: $(rand_int 60 99)%" "Request processed in $(rand_int 5 200)ms" "Scheduled task executed successfully" "Connection pool: active=$(rand_int 1 10)/max=20")
  local msgs_warn=("Slow query detected: $(rand_int 500 2000)ms for SELECT * FROM orders" "Connection pool running low: $(rand_int 1 3) available" "Retry attempt $(rand_int 1 3)/3 for downstream service" "JWT token expiring soon for user $(rand_item "${USERS[@]}")")
  local msgs_error=("Failed to connect to database: Connection refused" "NullPointerException at $cls.process(${cls}.java:$(rand_int 50 300))" "Transaction rollback due to constraint violation" "HTTP 503 received from payment-service after 3 retries" "Redis connection pool exhausted" "OutOfMemoryError: Java heap space")

  local msg
  case $level in
    INFO)  msg=$(rand_item "${msgs_info[@]}") ;;
    WARN)  msg=$(rand_item "${msgs_warn[@]}") ;;
    ERROR) msg=$(rand_item "${msgs_error[@]}") ;;
    FATAL) msg="JVM crash detected - system out of memory, shutting down" ;;
  esac

  echo "$(ts_iso)  $level $(rand_int 10000 99999) --- [$tid] $cls : $msg" >> "$LOG_DIR/java/app.log"

  # Java ERROR 附带堆栈
  if [ "$level" = "ERROR" ]; then
    cat >> "$LOG_DIR/java/app.log" << STACK
java.lang.RuntimeException: $msg
	at com.example.$cls.process(${cls}.java:$(rand_int 50 300))
	at com.example.BaseService.execute(BaseService.java:$(rand_int 20 100))
	at com.example.ApiController.handle(ApiController.java:$(rand_int 30 80))
	at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
	at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:119)
STACK
  fi
}

# ---- Python 日志 ----
gen_python_log() {
  local level=$1 mod
  mod=$(rand_item "${PYTHON_MODULES[@]}")
  local msgs_info=("Request GET $(rand_item "${PATHS[@]}") 200 $(rand_int 10 500)ms" "User $(rand_item "${USERS[@]}") authenticated successfully" "Cache refreshed: $(rand_int 100 1000) items" "Celery task completed: send_email" "DB query returned $(rand_int 0 50) rows")
  local msgs_warn=("Deprecated API endpoint called: /api/v1/legacy" "Rate limit approaching: $(rand_int 80 99)% of quota used" "Slow response from Redis: $(rand_int 100 500)ms")
  local msgs_error=("ERROR Database connection failed: psycopg2.OperationalError: could not connect to server" "ERROR Unhandled exception in view $(rand_item "${PATHS[@]}"): KeyError: 'user_id'" "ERROR Celery task failed: ConnectionError max retries exceeded" "ERROR S3 upload failed: botocore.exceptions.ClientError")

  local msg
  case $level in
    INFO)    msg=$(rand_item "${msgs_info[@]}") ;;
    WARNING) msg=$(rand_item "${msgs_warn[@]}") ;;
    ERROR)   msg=$(rand_item "${msgs_error[@]}") ;;
    CRITICAL)msg="CRITICAL System health check failed - service unavailable" ;;
  esac

  echo "$(ts) $level     $mod - $msg" >> "$LOG_DIR/python/app.log"

  # Python ERROR 附带 Traceback
  if [ "$level" = "ERROR" ]; then
    cat >> "$LOG_DIR/python/app.log" << TRACEBACK
Traceback (most recent call last):
  File "/app/$mod.py", line $(rand_int 50 200), in handle_request
    result = self.service.process(data)
  File "/app/services/base.py", line $(rand_int 30 100), in process
    return self.db.execute(query)
$(echo "$msg" | sed 's/^ERROR //')
TRACEBACK
  fi
}

# ---- Go 应用日志（JSON 格式）----
gen_go_log() {
  local level=$1 pkg
  pkg=$(rand_item "${GO_PKGS[@]}")
  local msgs_info=("request completed" "health check passed" "task scheduled" "cache populated" "config reloaded")
  local msgs_warn=("slow query detected" "connection pool low" "retry triggered" "rate limit warning")
  local msgs_error=("failed to connect elasticsearch" "context deadline exceeded" "connection refused" "write: broken pipe" "panic recovered")

  local msg latency status
  case $level in
    info)  msg=$(rand_item "${msgs_info[@]}"); latency="$(rand_int 1 100)ms"; status=$(rand_item "200 200 200 201 204") ;;
    warn)  msg=$(rand_item "${msgs_warn[@]}"); latency="$(rand_int 200 999)ms"; status="429" ;;
    error) msg=$(rand_item "${msgs_error[@]}"); latency="$(rand_int 1000 5000)ms"; status="500" ;;
    fatal) msg="service startup failed: port already in use"; latency="0ms"; status="000" ;;
  esac

  printf '{"level":"%s","time":"%s","caller":"%s/handler.go:%d","msg":"%s","latency":"%s","status":%s,"host":"%s"}\n' \
    "$level" "$(ts_iso)" "$pkg" "$(rand_int 20 200)" "$msg" "$latency" "$status" "$(rand_item "${IPS[@]}")" \
    >> "$LOG_DIR/golang/app.log"
}

# ---- MySQL 慢查询日志 ----
gen_mysql_slow() {
  local query_time=$(rand_int 1 30)
  local rows=$(rand_int 1000 1000000)
  local tables=("users" "orders" "products" "logs" "sessions" "audit_logs")
  local table=$(rand_item "${tables[@]}")
  cat >> "$LOG_DIR/mysql/slow.log" << MYSQL
# Time: $(ts_mysql)
# User@Host: app[app] @ localhost []  Id: $(rand_int 100 9999)
# Query_time: $query_time.$(rand_int 100000 999999)  Lock_time: 0.000$(rand_int 100 999)  Rows_sent: $(rand_int 1 1000)  Rows_examined: $rows
SET timestamp=$(date +%s);
SELECT * FROM $table WHERE created_at > DATE_SUB(NOW(), INTERVAL 7 DAY) ORDER BY id DESC LIMIT 1000;
MYSQL
}

# ---- Node.js 日志 ----
gen_nodejs_log() {
  local level=$1
  local msgs_info=("[HTTP] $(rand_item "GET POST PUT DELETE") $(rand_item "${PATHS[@]}") $(rand_item "200 200 201 204") $(rand_int 5 200)ms" "[DB] Query executed in $(rand_int 1 50)ms" "[Cache] HIT for key:user:$(rand_int 1 1000)" "[Auth] Token validated for user $(rand_item "${USERS[@]}")" "[Queue] Message processed: job_$(rand_int 1000 9999)")
  local msgs_warn=("[DB] Connection pool usage: $(rand_int 70 90)%" "[HTTP] Slow response: $(rand_int 500 2000)ms" "[Queue] Message retry: attempt $(rand_int 2 4)")
  local msgs_error=("[DB] Error: connect ECONNREFUSED 127.0.0.1:5432" "[HTTP] Unhandled rejection: TypeError: Cannot read property 'id' of undefined" "[Auth] JWT verification failed: invalid signature" "[Queue] Worker crashed: $(rand_int 1 5) times restarted")

  local msg
  case $level in
    INFO)  msg=$(rand_item "${msgs_info[@]}") ;;
    WARN)  msg=$(rand_item "${msgs_warn[@]}") ;;
    ERROR) msg=$(rand_item "${msgs_error[@]}") ;;
  esac

  echo "[$(ts)] $level $msg" >> "$LOG_DIR/nodejs/app.log"
}

# ---- PHP 错误日志 ----
gen_php_log() {
  local level=$1
  local msgs_notice=("[$(ts)] PHP Notice: Undefined index: user_id in /var/www/app/controllers/UserController.php on line $(rand_int 50 300)" "[$(ts)] PHP Notice: Array to string conversion in /var/www/app/helpers.php on line $(rand_int 10 100)")
  local msgs_warn=("[$(ts)] PHP Warning: mysqli_connect(): (HY000/2002): Connection refused in /var/www/app/db.php on line $(rand_int 10 50)" "[$(ts)] PHP Warning: file_get_contents(https://api.example.com/): failed to open stream: HTTP request failed")
  local msgs_error=("[$(ts)] PHP Fatal error: Uncaught Exception: Database connection failed in /var/www/app/db.php:$(rand_int 20 80)" "[$(ts)] PHP Fatal error: Maximum execution time of 30 seconds exceeded in /var/www/app/export.php:$(rand_int 100 500)" "[$(ts)] PHP Fatal error: Allowed memory size of 134217728 bytes exhausted")

  case $level in
    notice) rand_item "${msgs_notice[@]}" >> "$LOG_DIR/php/app.log" ;;
    warn)   rand_item "${msgs_warn[@]}"   >> "$LOG_DIR/php/app.log" ;;
    error)  rand_item "${msgs_error[@]}"  >> "$LOG_DIR/php/app.log" ;;
  esac
}

# ---- AI 告警触发：大量 ERROR ----
trigger_error_burst() {
  echo -e "\n${RED}>>> 触发 AI 告警测试：写入大量 ERROR 日志...${NC}"
  local count=30
  for i in $(seq 1 $count); do
    gen_java_log   "ERROR"
    gen_python_log "ERROR"
    gen_go_log     "error"
    gen_nginx_error "error"
    gen_nodejs_log "ERROR"
    sleep 0.1
  done
  echo -e "${RED}>>> 已写入 $count 条 ERROR 日志（每种格式），AI 告警应在静默期结束后触发${NC}\n"
}

# ---- 主循环 ----
if $ERROR_BURST; then
  trigger_error_burst
  echo -e "${YELLOW}测试日志路径列表:${NC}"
  ls -lh "$LOG_DIR"/*/*.log 2>/dev/null
  exit 0
fi

COUNT=0
while true; do
  COUNT=$((COUNT + 1))

  # 每次循环按权重随机生成
  ROLL=$((RANDOM % 100))

  # 70% 正常日志
  if [ $ROLL -lt 50 ]; then
    gen_nginx_access
    gen_java_log   "INFO"
    gen_python_log "INFO"
    gen_go_log     "info"
    gen_nodejs_log "INFO"

  elif [ $ROLL -lt 70 ]; then
    gen_nginx_access
    gen_go_log     "info"
    gen_python_log "INFO"

  # 15% 警告
  elif [ $ROLL -lt 85 ]; then
    gen_nginx_error "warn"
    gen_java_log   "WARN"
    gen_python_log "WARNING"
    gen_go_log     "warn"
    gen_nodejs_log "WARN"
    gen_php_log    "warn"

  # 10% 错误
  elif [ $ROLL -lt 95 ]; then
    gen_nginx_error "error"
    gen_java_log   "ERROR"
    gen_python_log "ERROR"
    gen_go_log     "error"
    gen_nodejs_log "ERROR"
    gen_php_log    "error"

  # 5% 慢查询 + PHP + 严重错误
  else
    gen_mysql_slow
    gen_php_log "error"
    gen_java_log "FATAL"
  fi

  # 每 50 条输出一次状态
  if [ $((COUNT % 50)) -eq 0 ]; then
    echo -e "${GREEN}[$(ts)]${NC} 已写入 ${CYAN}$COUNT${NC} 批日志"
    for f in "$LOG_DIR"/*/*.log; do
      lines=$(wc -l < "$f" 2>/dev/null || echo 0)
      size=$(du -sh "$f" 2>/dev/null | cut -f1)
      printf "  %-40s %6d 行  %s\n" "$f" "$lines" "$size"
    done
    echo ""
  fi

  # 每 200 条自动触发一次 ERROR 爆发（模拟真实故障）
  if [ $((COUNT % 200)) -eq 0 ]; then
    echo -e "${RED}[$(ts)] 自动触发 ERROR 爆发（模拟故障）...${NC}"
    for i in $(seq 1 15); do
      gen_java_log   "ERROR"
      gen_go_log     "error"
      gen_python_log "ERROR"
    done
    echo -e "${RED}[$(ts)] ERROR 爆发写入完成${NC}"
  fi

  sleep "$INTERVAL"
done
