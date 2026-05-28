#!/usr/bin/env python3
"""
MySQL youlin_test.max_adv_data → LCA 日志系统批量导入脚本

用法:
    python3 scripts/import_mysql_to_lca.py

依赖:
    pip3 install pymysql requests

MySQL: root/max2024test@120.79.60.98:3306/youlin_test
API:   http://192.168.18.210/api/v1/log/push  (api_key=ak_api_174950a4)
"""

import json
import sys
from datetime import datetime, date
from decimal import Decimal

try:
    import pymysql
except ImportError:
    print("请先安装 pymysql: pip3 install pymysql")
    sys.exit(1)

try:
    import requests
except ImportError:
    print("请先安装 requests: pip3 install requests")
    sys.exit(1)

# ─── 配置 ────────────────────────────────────────────

MYSQL_CONFIG = {
    "host": "120.79.60.98",
    "port": 3306,
    "user": "root",
    "password": "max2024test",
    "database": "youlin_test",
    "charset": "utf8mb4",
}

API_URL = "http://192.168.18.210/api/v1/log/push"
API_KEY = "ak_adv_data_b6ab8e17"
BATCH_SIZE = 500  # 每批最多 500 条（API 上限 5000）


# ─── 工具函数 ────────────────────────────────────────

def serialize_value(v):
    """把 MySQL 返回的类型转为 JSON 可序列化的值"""
    if v is None:
        return None
    if isinstance(v, (datetime, date)):
        return v.isoformat()
    if isinstance(v, Decimal):
        return float(v)
    if isinstance(v, bytes):
        return v.decode("utf-8", errors="replace")
    if isinstance(v, (int, float, bool, str)):
        return v
    # 兜底：转字符串
    return str(v)


def row_to_dict(row: tuple, columns: list) -> dict:
    """把 pymysql 返回的 tuple + columns 转为 dict"""
    result = {}
    for col, val in zip(columns, row):
        result[col] = serialize_value(val)
    return result


# ─── 主流程 ──────────────────────────────────────────

def main():
    # 1. 连接 MySQL
    print(f"连接 MySQL: {MYSQL_CONFIG['host']}:{MYSQL_CONFIG['port']}/{MYSQL_CONFIG['database']}")
    conn = pymysql.connect(**MYSQL_CONFIG)
    cursor = conn.cursor()

    # 2. 查询全表
    table = "max_adv_data"
    cursor.execute(f"SELECT * FROM `{table}`")
    columns = [desc[0] for desc in cursor.description]
    rows = cursor.fetchall()
    total = len(rows)
    print(f"查询完成: {total} 行, {len(columns)} 列: {', '.join(columns[:10])}{'...' if len(columns) > 10 else ''}")

    if total == 0:
        print("表为空，无需导入")
        cursor.close()
        conn.close()
        return

    # 3. 转为 dict 列表
    logs = [row_to_dict(r, columns) for r in rows]

    # 4. 分批推送到 LCA API
    offset = 0
    success = 0
    failed = 0

    while offset < total:
        batch = logs[offset : offset + BATCH_SIZE]
        payload = {"api_key": API_KEY, "logs": batch}

        try:
            resp = requests.post(API_URL, json=payload, timeout=30)
            data = resp.json()

            if resp.status_code == 200 and data.get("code") == 0:
                count = data.get("data", {}).get("count", len(batch))
                trace_id = data.get("data", {}).get("trace_id", "-")
                success += count
                print(f"  ✅ [{offset:5d}:{offset+len(batch):5d}] count={count} trace={trace_id}")
            else:
                failed += len(batch)
                print(f"  ❌ [{offset:5d}:{offset+len(batch):5d}] status={resp.status_code} code={data.get('code')} msg={data.get('message')}")
        except requests.exceptions.RequestException as e:
            failed += len(batch)
            print(f"  ❌ [{offset:5d}:{offset+len(batch):5d}] 网络错误: {e}")
        except json.JSONDecodeError:
            failed += len(batch)
            print(f"  ❌ [{offset:5d}:{offset+len(batch):5d}] 响应非 JSON: {resp.text[:200]}")

        offset += BATCH_SIZE

    # 5. 清理
    cursor.close()
    conn.close()

    print(f"\n{'='*50}")
    print(f"导入完成: 成功 {success} / 失败 {failed} / 总计 {total}")
    if failed > 0:
        print(f"⚠️  有 {failed} 条导入失败，请检查 API 服务状态后重试")


if __name__ == "__main__":
    main()
