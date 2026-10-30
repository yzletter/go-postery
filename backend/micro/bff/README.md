# BFF（Backend For Frontend)

## 目录

```
├── README.md
├── dto                     // 定义各个模块的请求和响应结构体
│   ├── auth
│   ├── interview
│   ├── lottery
│   ├── post
│   ├── search
│   ├── session
│   └── user
├── errno                   // 返回给前端的错误
│   └── errors.go
├── handler                 // 各模块 Handler
│   ├── auth_handler.go
│   ├── interactive_handler.go
│   ├── interview_handler.go
│   ├── lottery_handler.go
│   ├── post_handler.go
│   ├── search_handler.go
│   ├── session_handler.go
│   ├── user_handler.go
│   └── websocket_handler.go
├── main.go                 // 程序入口
├── middleware              // 中间件层
│   ├── auth.go                 // 鉴权中间件
│   ├── metric.go               // Prometheus 观测中间件
│   ├── ratelimit.go            // 限流中间件
│   └── tracing.go              // Trace 链路追踪中间件
└── service                 // Service 层
    ├── metric_service.go
    └── websocket_service.go
```
