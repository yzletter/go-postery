#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${ETCD_CONFIG_FILE:-$SCRIPT_DIR/etcd/config.example.conf}"
ENDPOINT="${ETCD_ENDPOINT:-http://localhost:2379}"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "配置文件不存在：$CONFIG_FILE" >&2
    exit 1
fi

while IFS= read -r line || [ -n "$line" ]; do
    # 忽略空行和注释
    [ -z "$line" ] && continue
    [[ "$line" == \#* ]] && continue

    key="${line%%=*}"
    value="${line#*=}"

    etcdctl --endpoints="$ENDPOINT" put "$key" "$value"
done < "$CONFIG_FILE"
