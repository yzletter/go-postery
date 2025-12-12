# go-postery

## 简介

用 Go 实现一个简单的信息发布系统

![homepage](/img/homepage.png)

## 项目进度

Gin + Gorm + Mysql + Redis + Viper + Slog + Promethus + Grafana + Crontab + Lua

- **功能：**通过**雪花算法**生成分布式ID；
  - **用户模块：** 注册、登录、主页、修改密码、修改个人资料；
  - **帖子模块：** 发布、删除、编辑、点赞、取消点赞；
  - **评论模块：** 回复帖子、回复评论；
  - **标签模块：** 文章发表可带标签
- **配置：** 使用 **Viper** 进行配置读取，使用 **Slog** 日志库；
- **限流：** 通过 **Redis + Lua** 实现**滑动窗口限流**；
- **运行：** 通过 **Crontab** 执行**定时任务**，利用信号机制完成**优雅关机**；
- **鉴权：** 结合 **JWT** 使用**双 Token** 机制开发鉴权中间件；
- **监控：** 通过 **Promethus + Grafana** 统计接口 **QPS 和平均耗时**；

## 待开发

- **热门榜单：** 采用 Reddit 算法，通过 Redis Zset (或 try 本地手写堆) 实现；
- **点赞：** 当前版本通过 Kafka 进行改造
- **搜索：** 集成 Go-Searchery 手写分布式搜索引擎；
- **标签分类导航**
- **关注模块**
- **私信：** 集成 Go-Chatery 即时通讯系统，利用 RabbitMQ；
- **抽奖：** 高并发秒杀，利用 RocketMQ；
- **AI Agent：** 接 OpenAI 开发一个 Agent
- **微服务部署与上线**
- **管理员后台**
- **重构 Repository 层：** 拆分为 DAO 和 Cache 层

## 设计文档
