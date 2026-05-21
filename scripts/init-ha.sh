#!/bin/bash
# ==========================================================
# LCA HA 基础设施初始化脚本
# 功能：MySQL 主从复制初始化、ES 备份仓库注册
# 使用：docker compose -f docker-compose.ha.yaml up -d 后执行本脚本
# ==========================================================
set -e

MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-lca2024}"
MYSQL_DATABASE="${MYSQL_DATABASE:-lca}"
REPL_USER="${REPL_USER:-repl}"
REPL_PASSWORD="${REPL_PASSWORD:-repl2024}"

echo "=========================================="
echo " LCA HA Infrastructure Init"
echo "=========================================="

# ---------- 1. MySQL 主从复制 ----------
echo ""
echo "[1/3] Configuring MySQL Replication..."

# 等待 master 和 slave 启动完成
echo "  Waiting for mysql-master..."
until docker exec lca-mysql-master mysqladmin ping -h localhost -p"${MYSQL_ROOT_PASSWORD}" --silent 2>/dev/null; do
    sleep 2
done
echo "  Waiting for mysql-slave..."
until docker exec lca-mysql-slave mysqladmin ping -h localhost -p"${MYSQL_ROOT_PASSWORD}" --silent 2>/dev/null; do
    sleep 2
done

# 在 master 上创建复制用户
echo "  Creating replication user on master..."
docker exec lca-mysql-master mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "
  CREATE USER IF NOT EXISTS '${REPL_USER}'@'%' IDENTIFIED WITH mysql_native_password BY '${REPL_PASSWORD}';
  GRANT REPLICATION SLAVE ON *.* TO '${REPL_USER}'@'%';
  FLUSH PRIVILEGES;
"

# 在 slave 上配置复制
echo "  Configuring slave to replicate from master..."
docker exec lca-mysql-slave mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "
  STOP SLAVE;
  CHANGE MASTER TO
    MASTER_HOST='mysql-master',
    MASTER_PORT=3306,
    MASTER_USER='${REPL_USER}',
    MASTER_PASSWORD='${REPL_PASSWORD}',
    MASTER_AUTO_POSITION=1;
  START SLAVE;
"

# 验证复制状态
echo "  Verifying replication status..."
SLAVE_STATUS=$(docker exec lca-mysql-slave mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "SHOW SLAVE STATUS\G" 2>/dev/null)
IO_RUNNING=$(echo "$SLAVE_STATUS" | grep "Slave_IO_Running" | awk '{print $2}')
SQL_RUNNING=$(echo "$SLAVE_STATUS" | grep "Slave_SQL_Running:" | awk '{print $2}')

if [ "$IO_RUNNING" = "Yes" ] && [ "$SQL_RUNNING" = "Yes" ]; then
    echo "  ✅ MySQL Replication: OK (IO=Yes, SQL=Yes)"
else
    echo "  ⚠️  MySQL Replication: IO=$IO_RUNNING, SQL=$SQL_RUNNING (check logs)"
fi

# ---------- 2. Elasticsearch 备份仓库 ----------
echo ""
echo "[2/3] Registering ES backup repository..."

# 等待 ES 集群就绪
echo "  Waiting for ES cluster..."
until curl -sf http://localhost:9200/_cluster/health?wait_for_status=green\&timeout=5s >/dev/null 2>&1; do
    sleep 3
done

# 注册文件系统备份仓库
curl -sf -X PUT "http://localhost:9200/_snapshot/lca_backup" \
  -H 'Content-Type: application/json' \
  -d '{
    "type": "fs",
    "settings": {
      "location": "/data/backups",
      "compress": true
    }
  }' >/dev/null

echo "  ✅ ES backup repository 'lca_backup' registered"

# 设置 ES 索引副本（所有 lca_ 前缀索引设为 1 副本）
echo "  Setting index template with 1 replica..."
curl -sf -X PUT "http://localhost:9200/_index_template/lca_ha_template" \
  -H 'Content-Type: application/json' \
  -d '{
    "index_patterns": ["lca_*", "app-*"],
    "priority": 100,
    "template": {
      "settings": {
        "number_of_replicas": 1,
        "number_of_shards": 3
      }
    }
  }' >/dev/null

echo "  ✅ ES index template configured (replicas=1, shards=3)"

# ---------- 3. Kafka Topic 验证 ----------
echo ""
echo "[3/3] Verifying Kafka cluster..."

# 等待 Kafka broker 就绪
echo "  Waiting for Kafka brokers..."
until docker exec lca-kafka1 kafka-broker-api-versions.sh --bootstrap-server localhost:9092 >/dev/null 2>&1; do
    sleep 3
done

# 列出 topic 验证集群正常
TOPIC_COUNT=$(docker exec lca-kafka1 kafka-topics.sh --bootstrap-server localhost:9092 --list 2>/dev/null | wc -l)
echo "  ✅ Kafka cluster ready (${TOPIC_COUNT} topics)"

echo ""
echo "=========================================="
echo " ✅ HA Infrastructure Init Complete!"
echo "=========================================="
echo ""
echo "Services:"
echo "  MySQL Master:  localhost:3306"
echo "  MySQL Slave:   localhost:3307 (read-only)"
echo "  Redis Sentinel: localhost:26379,26380,26381"
echo "  Kafka Brokers: kafka1:9092, kafka2:9092, kafka3:9092"
echo "  ES Cluster:    localhost:9200 (3 nodes)"
echo "  Nginx LB:      localhost:80"
echo ""
