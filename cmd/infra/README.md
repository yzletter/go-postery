# Go Postery 基础设施

`docker-compose.yml` 中的宿主机端口统一绑定到 `${HOST_IP:-127.0.0.1}`：

- 未设置 `HOST_IP` 时，仅允许本机通过 `127.0.0.1` 访问。
- 服务器部署时可设置 `HOST_IP` 为内网地址，例如 `10.0.0.11`。
- 下列端口均使用 TCP；管理页面和集群端口不建议直接暴露到公网。

## etcd 配置初始化

微服务所需的全部 etcd Key、默认值和导入方法见 [`etcd/README.md`](./etcd/README.md)。

## 宿主机端口

| 端口 | 服务 | 用途 | 开放建议 |
| ---: | --- | --- | --- |
| 2379 | etcd | Client API | 仅业务服务器/内网 |
| 2380 | etcd | Peer 通信 | 仅 etcd 集群节点；单节点无需对其他主机开放 |
| 3306 | MySQL | 数据库连接 | 仅业务服务器/内网 |
| 6379 | Redis | 缓存连接 | 仅业务服务器/内网 |
| 9000 | MinIO | S3 API | 仅业务服务器/内网 |
| 9001 | MinIO | 管理控制台 | 仅管理员/内网 |
| 19530 | Milvus | SDK / gRPC API | 仅业务服务器/内网 |
| 9091 | Milvus | WebUI、REST 与健康检查 | 仅管理员/内网 |
| 9092 | Kafka | Broker 外部客户端连接 | 仅业务服务器/内网 |
| 5672 | RabbitMQ | AMQP 客户端连接 | 仅业务服务器/内网 |
| 15672 | RabbitMQ | Management Web UI | 仅管理员/内网 |
| 4317 | Jaeger | OTLP gRPC 接收 | 仅业务服务器/内网 |
| 4318 | Jaeger | OTLP HTTP 接收 | 仅业务服务器/内网 |
| 16686 | Jaeger | 查询 Web UI | 仅管理员/内网 |
| 9876 | RocketMQ | NameServer | 仅 Broker、Proxy 和业务服务器/内网 |
| 10909 | RocketMQ | Broker Fast Remoting / VIP Channel | 仅业务服务器/内网 |
| 10911 | RocketMQ | Broker Remoting | 仅业务服务器/内网 |
| 10912 | RocketMQ | Broker HA 同步 | 仅 Broker 节点/内网 |
| 8080 | RocketMQ Proxy | Remoting 接入 | 仅业务服务器/内网 |
| 8081 | RocketMQ Proxy | gRPC 接入 | 仅业务服务器/内网 |
| 9090 | Prometheus | Web UI、HTTP API | 仅管理员/内网 |
| 3000 | Grafana | Web UI | 仅管理员/内网或经反向代理访问 |

## Docker 内部端口

这些端口没有发布到宿主机，不需要配置服务器防火墙：

| 端口 | 服务 | 用途 |
| ---: | --- | --- |
| 2379 | milvus-etcd | Milvus 元数据存储，仅供 Compose 网络访问 |
| 29092 | Kafka | Broker 内部 Listener |
| 9093 | Kafka | KRaft Controller Listener |
| 13133 | Jaeger | 容器健康检查端点 |

## 端口汇总

需要在服务器安全组或防火墙中按实际调用关系放行的宿主机端口：

```text
2379, 2380, 3000, 3306, 4317, 4318, 5672, 6379,
8080, 8081, 9000, 9001, 9090, 9091, 9092, 9876,
10909, 10911, 10912, 15672, 16686, 19530
```
