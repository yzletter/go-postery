# 数据初始化文件

本目录维护 MySQL、etcd 和 Redis 的初始化数据文件：

- `create_table.sql`：MySQL 建库、建表和种子数据。
- `redis`：Redis 初始化命令。
- `config.example.conf`：可提交到仓库的 etcd 配置模板，共 72 个 Key。
- `config.local.conf`：本地实际配置，已加入 `.gitignore`，可填写密钥等敏感值。
- `../init-data.sh`：初始化 MySQL、etcd 和 Redis 数据；空值也会创建对应 etcd Key。

> 下列账号和密码仅用于本地开发/测试。部署到共享或生产环境前必须更换。导入会覆盖 etcd 中同名 Key 的现有值。

## 使用方法

```bash
cd cmd
cp script/config.local.conf script/config.local.conf

# 编辑配置，补充本文标记为“空”的账号、密钥和云资源信息
vim script/config.local.conf

# 基础设施启动后，初始化 MySQL、etcd 和 Redis 数据
./init-data.sh
```

指定其他 etcd 地址或配置文件：

```bash
cd cmd
ETCD_ENDPOINT=http://localhost:2379 \
ETCD_CONFIG_FILE=./script/config.local.conf \
./init-data.sh
```

脚本通过环境变量 `ETCD_ENDPOINT` 和 `ETCD_CONFIG_FILE` 指定 etcd 地址和配置文件。默认 etcd 地址为 `http://localhost:2379`；未指定配置文件时优先使用 `script/config.local.conf`，文件不存在则使用 `script/config.example.conf`。

配置文件中的 Key 位于以下前缀：

- `go_postery/conf/common_conf/`
- `go_postery/conf/service_conf/`

配置格式为 `key=value`，以第一个等号分隔，因此值中可以继续包含等号。空值写成 `key=`，不要给值额外添加引号。

## 公共配置（16 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/common_conf/redis/addr` | `localhost:6379` |
| `go_postery/conf/common_conf/milvus/addr` | `localhost:19530` |
| `go_postery/conf/common_conf/jaeger/addr` | `localhost:4318` |
| `go_postery/conf/common_conf/kafka/addr` | `localhost:9092` |
| `go_postery/conf/common_conf/rabbitmq/addr` | `localhost:5672` |
| `go_postery/conf/common_conf/rabbitmq/user` | `go_postery` |
| `go_postery/conf/common_conf/rabbitmq/password` | `123456` |
| `go_postery/conf/common_conf/rocketmq/addr` | `localhost:8081` |
| `go_postery/conf/common_conf/mysql/addr` | `localhost:3306` |
| `go_postery/conf/common_conf/mysql/user` | `go_postery_tester` |
| `go_postery/conf/common_conf/mysql/password` | `123456` |
| `go_postery/conf/common_conf/mysql/db_name` | `go_postery` |
| `go_postery/conf/common_conf/mysql/log/file_dir` | `./logs` |
| `go_postery/conf/common_conf/mysql/log/filename` | `mysql.log` |
| `go_postery/conf/common_conf/service_hub/heartbeat_frequency` | `10` |
| `go_postery/conf/common_conf/service_hub/register_prefix` | `go_postery/service/` |

Jaeger 使用项目当前的 OTLP HTTP 导出器，因此配置端口为 `4318`；RocketMQ 使用 Go 5.x SDK 的 Proxy gRPC 接入，因此配置端口为 `8081`。

## Auth 服务（3 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/auth_service_conf/log/filepath` | `./logs/auth_service.log` |
| `go_postery/conf/service_conf/auth_service_conf/prometheus/port` | `9102` |
| `go_postery/conf/service_conf/auth_service_conf/grpc/port` | `9002` |

## Code 服务（12 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/code_service_conf/log/filepath` | `./logs/code_service.log` |
| `go_postery/conf/service_conf/code_service_conf/prometheus/port` | `9101` |
| `go_postery/conf/service_conf/code_service_conf/grpc/port` | `9001` |
| `go_postery/conf/service_conf/code_service_conf/email/from` | （空） |
| `go_postery/conf/service_conf/code_service_conf/email/auth_code` | （空） |
| `go_postery/conf/service_conf/code_service_conf/email/subject` | `Go Postery - 验证您的邮箱` |
| `go_postery/conf/service_conf/code_service_conf/email/app_name` | `Go Postery` |
| `go_postery/conf/service_conf/code_service_conf/email/expire_min` | `10` |
| `go_postery/conf/service_conf/code_service_conf/email/year` | `2026` |
| `go_postery/conf/service_conf/code_service_conf/email/address` | （空） |
| `go_postery/conf/service_conf/code_service_conf/sms/access_key_id` | （空） |
| `go_postery/conf/service_conf/code_service_conf/sms/access_key_secret` | （空） |

## Interactive 服务（3 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/interactive_service_conf/log/filepath` | `./logs/interactive_service.log` |
| `go_postery/conf/service_conf/interactive_service_conf/prometheus/port` | `9110` |
| `go_postery/conf/service_conf/interactive_service_conf/grpc/port` | `9010` |

## Interview 服务（13 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/interview_service/log/filepath` | `./logs/interview_service.log` |
| `go_postery/conf/service_conf/interview_service/prometheus/port` | `9111` |
| `go_postery/conf/service_conf/interview_service/grpc/port` | `9011` |
| `go_postery/conf/service_conf/interview_service/ark/embedder_model` | `doubao-embedding-vision-251215` |
| `go_postery/conf/service_conf/interview_service/ark/llm_model` | `doubao-seed-2-1-turbo-260628` |
| `go_postery/conf/service_conf/interview_service/ark/api_key` | （空） |
| `go_postery/conf/service_conf/interview_service/qwen/embedder_model` | `text-embedding-v3` |
| `go_postery/conf/service_conf/interview_service/qwen/llm_base_url` | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| `go_postery/conf/service_conf/interview_service/qwen/embedder_base_url` | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| `go_postery/conf/service_conf/interview_service/qwen/llm_model` | `qwen-plus` |
| `go_postery/conf/service_conf/interview_service/qwen/llm_api_key` | （空） |
| `go_postery/conf/service_conf/interview_service/qwen/embedder_api_key` | （空） |
| `go_postery/conf/service_conf/interview_service/github/token` | （空） |

Qwen 的 API Key 和 Base URL 分别用于 LLM 和 Embedding，不再读取旧的 `qwen/api_key` 和 `qwen/base_url`。更新已有配置文件时需使用上述新键；写入 etcd 后，重启 Interview 服务以加载新配置。

## Lottery 服务（3 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/lottery_service_conf/log/filepath` | `./logs/lottery_service.log` |
| `go_postery/conf/service_conf/lottery_service_conf/prometheus/port` | `9103` |
| `go_postery/conf/service_conf/lottery_service_conf/grpc/port` | `9003` |

## OSS 服务（9 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/oss_service_conf/log/filepath` | `./logs/oss_service.log` |
| `go_postery/conf/service_conf/oss_service_conf/prometheus/port` | `9113` |
| `go_postery/conf/service_conf/oss_service_conf/grpc/port` | `9013` |
| `go_postery/conf/service_conf/oss_service_conf/oss/access_key_id` | （空） |
| `go_postery/conf/service_conf/oss_service_conf/oss/access_key_secret` | （空） |
| `go_postery/conf/service_conf/oss_service_conf/oss/arn` | （空） |
| `go_postery/conf/service_conf/oss_service_conf/oss/region` | （空） |
| `go_postery/conf/service_conf/oss_service_conf/oss/bucket` | （空） |
| `go_postery/conf/service_conf/oss_service_conf/oss/callback_url` | （空） |

## Post 服务（3 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/post_service_conf/log/filepath` | `./logs/post_service.log` |
| `go_postery/conf/service_conf/post_service_conf/prometheus/port` | `9104` |
| `go_postery/conf/service_conf/post_service_conf/grpc/port` | `9004` |

## Rank 服务（3 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/rank_service_conf/log/filepath` | `./logs/rank_service.log` |
| `go_postery/conf/service_conf/rank_service_conf/prometheus/port` | `9109` |
| `go_postery/conf/service_conf/rank_service_conf/grpc/port` | `9009` |

## Search 服务（3 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/search_service_conf/log/filepath` | `./logs/search_service.log` |
| `go_postery/conf/service_conf/search_service_conf/prometheus/port` | `9105` |
| `go_postery/conf/service_conf/search_service_conf/grpc/port` | `9005` |

## Session 服务（3 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/session_service_conf/log/filepath` | `./logs/session_service.log` |
| `go_postery/conf/service_conf/session_service_conf/prometheus/port` | `9108` |
| `go_postery/conf/service_conf/session_service_conf/grpc/port` | `9008` |

## User 服务（3 个）

| Key | 默认值 |
| --- | --- |
| `go_postery/conf/service_conf/user_service_conf/log/filepath` | `./logs/user_service.log` |
| `go_postery/conf/service_conf/user_service_conf/prometheus/port` | `9107` |
| `go_postery/conf/service_conf/user_service_conf/grpc/port` | `9007` |

## 当前需要自行填写的 15 个 Key

- Code：发件邮箱、邮箱授权码、公司地址、短信 AccessKey ID 和 Secret。
- Interview：Ark API Key、Qwen LLM API Key、Qwen Embedding API Key、GitHub Token。
- OSS：AccessKey ID、AccessKey Secret、ARN、Region、Bucket 和 Callback URL。

填写 `config.local.conf` 并启动基础设施后，运行 `./init-data.sh` 导入全部配置和初始化数据。
