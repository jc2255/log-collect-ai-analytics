#!/bin/bash
# 确保本地快照目录对 elasticsearch 用户可写（Docker 命名卷默认属主是 root）
# 并尽量把容器/系统时区落到 Asia/Shanghai（SLM cron 跟 JVM 时区走）
set -e
export TZ="${TZ:-Asia/Shanghai}"
if [ -f /usr/share/zoneinfo/Asia/Shanghai ]; then
  ln -sfn /usr/share/zoneinfo/Asia/Shanghai /etc/localtime 2>/dev/null || true
  echo Asia/Shanghai > /etc/timezone 2>/dev/null || true
fi
# 官方镜像的 ES_JAVA_OPTS 可能被 .env 覆盖掉时区参数，这里兜底追加
if [[ "${ES_JAVA_OPTS:-}" != *"-Duser.timezone="* ]]; then
  export ES_JAVA_OPTS="${ES_JAVA_OPTS:-} -Duser.timezone=Asia/Shanghai"
fi
mkdir -p /data/backups
if [ "$(id -u)" = "0" ]; then
  chown -R elasticsearch:elasticsearch /data/backups 2>/dev/null || chown -R 1000:0 /data/backups || true
  chmod -R ug+rwX /data/backups || true
fi
exec /usr/local/bin/docker-entrypoint.sh "$@"
