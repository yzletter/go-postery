#!/usr/bin/env bash

set -e

# 获取脚本所在目录，确保从任意目录执行时都能找到 Compose 文件。
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/infra/docker-compose.yml"

# 停止并移除基础设施容器和网络，保留持久化数据卷。
docker compose -f "$COMPOSE_FILE" down
