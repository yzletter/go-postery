#!/usr/bin/env bash

set -e

# 获取模式
MODE="${1:-}"

# 本地模式
if [ "$MODE" = "local" ]; then
    HOST_IP="127.0.0.1"
# 生产模式
elif [ "$MODE" = "prod" ]; then
    # 自动获取服务器默认出口 IP
    HOST_IP=$(
        ip route get 1.1.1.1 |
        awk '{
            for(i=1;i<=NF;i++) {
                if($i=="src") {
                    print $(i+1)
                    exit
                }
            }
        }'
    )
# 手动指定 IP 模式
elif [[ "$MODE" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
    HOST_IP="$MODE"
# 用法提示
else
    echo "脚本用法："
    echo "  ./infra-up.sh local"
    echo "  ./infra-up.sh prod"
    echo "  ./infra-up.sh 10.0.0.11"
    exit 1
fi

export HOST_IP

echo "HOST_IP = $HOST_IP"

docker compose up -d