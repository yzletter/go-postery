#!/usr/bin/env bash

set -e

# 获取脚本所在目录，确保从任意目录执行时都能正确找到初始化文件。
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MYSQL_SQL_FILE="$SCRIPT_DIR/script/create_table.sql"
REDIS_DATA_FILE="$SCRIPT_DIR/script/redis"
ETCD_ENDPOINT="${ETCD_ENDPOINT:-http://localhost:2379}"

# 优先使用显式指定的配置，其次使用本地配置，最后回退到示例配置。
if [ -n "${ETCD_CONFIG_FILE:-}" ]; then
    CONFIG_FILE="$ETCD_CONFIG_FILE"
elif [ -f "$SCRIPT_DIR/script/config.local.conf" ]; then
    CONFIG_FILE="$SCRIPT_DIR/script/config.local.conf"
else
    CONFIG_FILE="$SCRIPT_DIR/script/config.example.conf"
fi

# 在执行任何导入前检查全部数据文件，避免只完成部分初始化。
if [ ! -f "$MYSQL_SQL_FILE" ]; then
    echo "数据库初始化文件不存在：$MYSQL_SQL_FILE" >&2
    exit 1
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "etcd 配置文件不存在：$CONFIG_FILE" >&2
    exit 1
fi

if [ ! -f "$REDIS_DATA_FILE" ]; then
    echo "Redis 初始化文件不存在：$REDIS_DATA_FILE" >&2
    exit 1
fi

# 初始化 MySQL：创建业务数据库、用户、数据表和种子数据。
docker exec -i go-postery-mysql \
    mysql -uroot -p123456 < "$MYSQL_SQL_FILE"

# 初始化 etcd：按 key=value 格式逐行导入配置，空值也会正常写入。
while IFS= read -r line || [ -n "$line" ]; do
    # 忽略空行和以 # 开头的注释行。
    [ -z "$line" ] && continue
    [[ "$line" == \#* ]] && continue

    key="${line%%=*}"
    value="${line#*=}"

    docker exec go-postery-etcd \
        etcdctl --endpoints="$ETCD_ENDPOINT" put "$key" "$value"
done < "$CONFIG_FILE"

# 初始化 Redis：执行文件中的 Redis 命令并覆盖对应 Key。
docker exec -i go-postery-redis \
    redis-cli < "$REDIS_DATA_FILE"

echo "MySQL、etcd 和 Redis 数据初始化完成"
