# go-postery

<p align="center">
  <a href="https://github.com/w8t-io/WatchAlert"> 
    <img src="imgs/logo.png" alt="cloud native monitoring" width="150" height="auto" /></a>
</p>

<p align="center">
  <b>📖 Go-Postery —— 现代化论坛 Web 项目</b>
</p>


## 简介

- 用 Go 实现一个现代化论坛
- 代码量：5000+ 行

## 项目进度

**Gin + Gorm + Mysql + Redis + Viper + Slog + RabbitMQ + Promethus + Grafana + Crontab + Lua**

- **功能：**通过**雪花算法**生成分布式ID；
  - **用户模块：** 注册、登录、个人主页、修改密码、修改个人资料、在线状态过期；
  - **帖子模块：** 发布、删除、编辑帖子、热门榜单；
  - **点赞模块：** 点赞、取消点赞；
  - **评论模块：** 回复帖子、回复评论；
  - **标签模块：** 带标签发表、修改标签、按标签导航；
  - **关注模块：** 关注、取消关注、获取关注列表、获取粉丝列表；
  - **私信模块：** 支持一对一单聊私信；
- **配置：** 使用 **Viper** 进行配置读取，使用 **Slog** 日志库；
- **限流：** 通过 **Redis + Lua** 实现**滑动窗口限流**；
- **运行：** 通过 **Crontab** 执行**定时任务**，利用信号机制完成**优雅关机**；
- **鉴权：** 结合 **JWT** 使用**双 Token** 机制开发鉴权中间件；
- **监控：** 通过 **Promethus + Grafana** 统计接口 **QPS 和平均耗时**；

## 待开发

- **用户头像**
- **Auth 中间件针对 Websocket 连接的优化**
- **点赞：** 当前版本通过 Kafka 进行改造
- **搜索：** 集成 Go-Searchery 手写分布式搜索引擎；
- **私信：** 对私信前置进行校验（当前为任意皆可私信）、群聊
- **抽奖：** 高并发秒杀，利用 RocketMQ；
- **AI Agent：** 接 OpenAI 开发一个 Agent
- **微服务部署与上线**
- **管理员后台**
- **拉黑功能**

## 项目演示

| ![首页.png](imgs/%E9%A6%96%E9%A1%B5.png) | ![帖子详情.png](imgs/%E5%B8%96%E5%AD%90%E8%AF%A6%E6%83%85.png)|
|:--------------------------:|------------------------------|
|    ![评论区.png](imgs/%E8%AF%84%E8%AE%BA%E5%8C%BA.png)    | ![发布帖子.png](imgs/%E5%8F%91%E5%B8%83%E5%B8%96%E5%AD%90.png) |
|   ![关注页面.png](imgs/%E5%85%B3%E6%B3%A8%E9%A1%B5%E9%9D%A2.png)   | ![修改个人资料.png](imgs/%E4%BF%AE%E6%94%B9%E4%B8%AA%E4%BA%BA%E8%B5%84%E6%96%99.png)         |
| ![个人主页.png](imgs/%E4%B8%AA%E4%BA%BA%E4%B8%BB%E9%A1%B5.png) | ![个人主页.png](imgs/%E4%B8%AA%E4%BA%BA%E4%B8%BB%E9%A1%B5.png)|

## 设计文档
