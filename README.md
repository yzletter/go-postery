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
- 后端纯代码量：7300 + 行

## 项目进度

**Gin + Gorm + Mysql + Redis + Viper + Slog + Kafka + RabbitMQ + RocketMQ + Promethus + Grafana + Crontab + Lua**

- **功能：** 通过 **SnowFlake** 生成**分布式ID**；
  - **认证功能：**
    - 验证码功能：**阿里云 SMS** 发送短信验证码；**QQ 邮箱 SMTP** 发送邮件验证码；
    - 登录功能：手机号码 + 验证码进行登录（未注册的手机号自动注册）；手机号码/邮箱 + 密码进行登录；退出登录
    - 密码功能：设置初始密码（首次登录的手机号自动注册的用户）；修改密码；
  - **个人功能：** 个人主页；修改个人资料；
  - **关注功能：** 关注、取消关注用户；关注列表、粉丝列表；推荐关注用户榜单；
  - **私信功能：** 一对一实时单聊私信；
  - **帖子功能：** 发布、删除、编辑帖子；热门帖子榜单；
    - 点赞功能： 点赞、取消点赞帖子；
    - 评论功能： 回复帖子、回复评论；
    - 标签功能： 带标签发表帖子、修改帖子标签；按标签导航；
  - **抽奖功能：** 高并发抽奖，单机压测 QPS 3160, avg time 73ms；
- **配置：** 使用 **Viper** 进行配置读取，使用 **Slog** 日志库；
- **限流：** 通过 **Redis + Lua** 实现**滑动窗口限流**；
- **运行：** 通过 **Crontab** 执行**定时任务**，利用信号机制完成**优雅关机**；
- **鉴权：** 结合 **JWT** 使用**双 Token** 机制开发鉴权中间件；
- **监控：** 通过 **Promethus + Grafana** 统计接口 **QPS 和平均耗时**；

## 待开发

- **用户头像**
- mysql 表的抢占设计
- 抽象出 Kafka Consumer
- 添加 RAG metadata
  1. **biz**：文本类型（plain / md）
  2. **source_id**：原始文档 id（你索引的 document_id）
  3. **chunk_id**：chunk 唯一 id（建议：`source_id + ":" + chunk_index` 或 snowflake）
  4. **chunk_index**：chunk 序号
  5. **chunk_total**：该 source 一共几个 chunk
  6. **start_offset / end_offset**：chunk 在原文中的字符区间（方便回溯定位、拼接上下文）
  7. **title / headers**：如果是 markdown，记录 header 路径（`# A / ## B`）
  8. **hash**：chunk 内容 hash（去重、幂等、更新对比非常好用）
- 对每个 email、phone 设置单日验证码上限防刷
- 注册 Session 前进行 DB 唯一约束避免打爆 RabbitMQ
- Outbox 对消息抢占避免（或乐观）多实例部署重复发送消息
  - next_retry_at 指数退避
  - 批量发送 Kafka：减少网络调用，提高吞吐
  - 运行控制：支持 ctx cancel + ticker.Stop()

- 密码校验 identifier
- 找回密码功能
- **关注用户发表的文章**
- **Auth 中间件针对 Websocket 连接的优化**
- **点赞：** 当前版本通过 Kafka 进行改造
- **搜索：** 集成 Go-Searchery 手写分布式搜索引擎；
- **私信：** 对私信前置进行校验（当前为任意皆可私信）、群聊
- **抽奖：** 中间状态保持
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
