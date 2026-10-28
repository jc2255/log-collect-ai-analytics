#!/usr/bin/env bash
# 模拟防火墙 / 路由器向 LCA syslog 发送 RFC3164 日志
#
# 用法:
#   bash scripts/simulate-syslog.sh
#   bash scripts/simulate-syslog.sh 127.0.0.1 514 udp 3
#   bash scripts/simulate-syslog.sh 192.168.1.10 514 tcp 1
#
# 参数: [主机] [端口] [udp|tcp] [每台设备发送条数]
# 服务监控里看到的主机名来自报文（fw-edge01 等）。
# Docker 发布 514 时，来源 IP 往往是网桥地址，不是本机 127.0.0.1。
set -euo pipefail

HOST="${1:-127.0.0.1}"
PORT="${2:-514}"
PROTO="${3:-udp}"
COUNT="${4:-3}"

export LC_ALL=C

if ! command -v nc >/dev/null 2>&1; then
  echo "需要 nc（netcat）" >&2
  exit 1
fi

# RFC3164 时间戳固定 15 字符: "Oct 27 16:06:01" 或 "Oct  5 16:06:01"
rfc3164_ts() {
  date "+%b %e %T"
}

send_msg() {
  local msg="$1"
  if [[ "$PROTO" == "tcp" ]]; then
    printf '%s\n' "$msg" | nc -w 2 "$HOST" "$PORT"
  else
    printf '%s\n' "$msg" | nc -u -w 1 "$HOST" "$PORT"
  fi
}

# priority: facility*8+severity
# local4=20 -> 20*8=160; info=6 -> 166; warning=4 -> 164; err=3 -> 163
emit_device() {
  local host="$1"
  local app="$2"
  shift 2
  local i msg ts
  for i in $(seq 1 "$COUNT"); do
    ts="$(rfc3164_ts)"
    for msg in "$@"; do
      send_msg "<${PRI:-166}>${ts} ${host} ${app}: ${msg}"
      sleep 0.2
    done
  done
}

echo "发送到 ${PROTO}://${HOST}:${PORT} ，每台设备 ${COUNT} 轮"

PRI=164
emit_device "fw-edge01" "USG" \
  "%%01SEC/4/ATCKDF(l): AttackType=port-scan, Source=10.1.2.8, Destination=192.168.1.10, Action=deny" \
  "%%01POLICY/6/POLICYPERMIT(l): vsys=public, policy=trust-to-untrust, src=10.1.2.20, dst=1.1.1.1, action=permit"

PRI=163
emit_device "fw-branch02" "ASA" \
  "%ASA-3-710003: TCP access denied by ACL from 172.16.8.9/51234 to outside:203.0.113.10/443" \
  "%ASA-4-106023: Deny udp src inside:10.8.0.15/53 dst outside:8.8.8.8/53 by access-group"

PRI=166
emit_device "rt-core01" "H3C" \
  "OSPF/6/OSPF_NBR_CHG: Neighbor 10.0.0.2 (GigabitEthernet0/0) changed from Loading to Full" \
  "BGP/5/BGP_STATE_CHANGED: Peer 10.255.0.1 state changed from OpenConfirm to Established"

PRI=165
emit_device "rt-wan01" "Router" \
  "LINEPROTO-5-UPDOWN: Line protocol on Interface GigabitEthernet0/1, changed state to down" \
  "SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.168.9.20)"

echo "发送完成。到「服务监控 → lca-syslog → 详情」查看 fw-edge01 / fw-branch02 / rt-core01 / rt-wan01"
