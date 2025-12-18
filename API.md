# Go Postery API 文档（基于 `main.go` 路由）

## 基础信息

- 服务监听：`http://localhost:8765`
- API 前缀：`/api/v1`
- 数据格式：请求/响应均为 `application/json`
- 统一响应结构：

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

- `code`：业务状态码（`0` 表示成功，非 `0` 表示失败）
- `msg`：提示信息
- `data`：响应数据（失败时通常为 `null` 或缺省）

## 认证说明

- AccessToken：响应头/请求头 `x-jwt-token`
- RefreshToken：Cookie `refresh-token`（HTTPOnly）
- 需要登录的接口：
  - 请求时携带 `x-jwt-token: <access_token>`
  - RefreshToken 由浏览器 Cookie 自动携带（如使用跨域 Cookie，请确保前后端允许携带凭证）

## 全局中间件行为（需要注意）

- 鉴权失败（需要登录接口）：直接返回 HTTP `401`，响应体可能为空，并清空 `x-jwt-token` / 删除 `refresh-token` Cookie。
- 触发限流：直接返回 HTTP `429`，响应体为空。
- 限流组件异常：直接返回 HTTP `500`，响应体为空。

## ID 类型说明

后端部分字段使用 `json:",string"`：

- **响应中** `id`/`post_id`/`parent_id`/`reply_id` 等会以 **字符串** 形式返回（避免前端数值精度丢失）。
- **请求中**如需传递 `parent_id`/`reply_id` 等字段，建议同样以 **字符串** 传递（例如：`"0"`、`"123"`）。

## 分页参数

带分页的接口通常支持：

- `pageNo`：页码（默认 `1`）
- `pageSize`：每页大小（默认 `10`，部分接口限制 `<= 100`）

响应通常包含：

- `total`：总数
- `hasMore`：是否还有下一页（后端按 `pageNo * pageSize < total` 计算）

## 接口总览

| 模块 | 方法 | 路径 | 登录 | 说明 |
|---|---|---|---|---|
| 运维 | GET | `/metrics` | 否 | Prometheus 指标 |
| 认证 | POST | `/api/v1/auth/register` | 否 | 注册 |
| 认证 | POST | `/api/v1/auth/login` | 否 | 登录 |
| 认证 | POST | `/api/v1/auth/logout` | 是 | 登出 |
| 用户 | GET | `/api/v1/users/:id` | 否 | 获取用户资料 |
| 用户 | GET | `/api/v1/users/:id/posts` | 否 | 按页获取用户帖子 |
| 用户 | POST | `/api/v1/users/me` | 是 | 修改个人资料 |
| 用户 | POST | `/api/v1/users/me/password` | 是 | 修改密码 |
| 关注 | GET | `/api/v1/users/me/followers` | 是 | 按页获取粉丝 |
| 关注 | GET | `/api/v1/users/me/followees` | 是 | 按页获取关注列表 |
| 关注 | POST | `/api/v1/users/:id/follow` | 是 | 关注用户 |
| 关注 | DELETE | `/api/v1/users/:id/follow` | 是 | 取关用户 |
| 关注 | GET | `/api/v1/users/:id/follow` | 是 | 查询关注关系 |
| 帖子 | GET | `/api/v1/posts` | 否 | 按页获取帖子列表 |
| 帖子 | GET | `/api/v1/posts/tags` | 否 | 按标签分页获取帖子 |
| 帖子 | GET | `/api/v1/posts/:id` | 否 | 获取帖子详情（会 +1 浏览量） |
| 评论 | GET | `/api/v1/posts/:id/comments` | 否 | 按页获取评论列表 |
| 评论 | GET | `/api/v1/posts/:id/comments/:cid` | 否 | 获取某评论的回复列表 |
| 帖子 | POST | `/api/v1/posts` | 是 | 创建帖子 |
| 帖子 | POST | `/api/v1/posts/:id` | 是 | 更新帖子 |
| 帖子 | DELETE | `/api/v1/posts/:id` | 是 | 删除帖子 |
| 评论 | POST | `/api/v1/posts/:id/comments` | 是 | 创建评论/回复 |
| 评论 | DELETE | `/api/v1/posts/:id/comments/:cid` | 是 | 删除评论 |
| 点赞 | GET | `/api/v1/posts/:id/likes` | 是 | 是否点赞该帖子 |
| 点赞 | POST | `/api/v1/posts/:id/likes` | 是 | 点赞帖子 |
| 点赞 | DELETE | `/api/v1/posts/:id/likes` | 是 | 取消点赞 |

---

## 运维接口

### GET `/metrics`

Prometheus 拉取指标使用（返回文本格式，非 JSON）。

## 认证模块

### POST `/api/v1/auth/register`

注册用户；成功后：

- 响应头写入 `x-jwt-token: <access_token>`
- 下发 Cookie：`refresh-token=<refresh_token>`

请求体：

```json
{
  "email": "user@example.com",
  "name": "alice",
  "password": "32位字符串（例如前端MD5后传输）"
}
```

响应 `data`：用户简要信息

```json
{
  "code": 0,
  "msg": "注册成功",
  "data": {
    "id": "1999760900969463808",
    "email": "user@example.com",
    "name": "alice",
    "avatar": ""
  }
}
```

常见错误：

- `10002` 参数错误（HTTP 400）
- `20002` 用户名或邮箱已存在（HTTP 409）

### POST `/api/v1/auth/login`

登录；成功后同样会返回 `x-jwt-token` 并下发 `refresh-token` Cookie。

请求体：

```json
{
  "name": "alice",
  "password": "32位字符串（例如前端MD5后传输）"
}
```

响应 `data`：用户简要信息

```json
{
  "code": 0,
  "msg": "登录成功",
  "data": {
    "id": "1999760900969463808",
    "email": "user@example.com",
    "name": "alice",
    "avatar": ""
  }
}
```

常见错误：

- `10002` 参数错误（HTTP 400）
- `20004` 账号或密码错误（HTTP 401）

### POST `/api/v1/auth/logout`（需要登录）

请求头：

- `x-jwt-token: <access_token>`

请求体：无

响应：

```json
{
  "code": 0,
  "msg": "登出成功"
}
```

说明：

- 服务端会清理 token，并将 `x-jwt-token` 置空、`refresh-token` Cookie 置为过期。

## 用户模块

### GET `/api/v1/users/:id`

获取用户资料。

路径参数：

- `id`：用户 ID

响应 `data`：用户详情

```json
{
  "code": 0,
  "msg": "获取个人资料成功",
  "data": {
    "id": "1999760900969463808",
    "name": "alice",
    "email": "user@example.com",
    "avatar": "https://example.com/avatar.png",
    "bio": "个人简介",
    "gender": 1,
    "birthday": "2024-01-01T00:00:00Z",
    "location": "Shanghai",
    "country": "China",
    "last_login_ip": "127.0.0.1"
  }
}
```

常见错误：

- `10002` 参数错误（HTTP 400）
- `20001` 用户不存在（HTTP 404）

### GET `/api/v1/users/:id/posts`

按页获取某个用户发布的帖子（简要信息）。

路径参数：

- `id`：用户 ID

Query 参数：

- `pageNo`（默认 `1`）
- `pageSize`（默认 `10`，限制 `<= 100`）

响应 `data`：

```json
{
  "code": 0,
  "msg": "获取帖子列表成功",
  "data": {
    "posts": [
      {
        "id": "1999760900969463808",
        "title": "标题",
        "created_at": "2024-01-01T00:00:00Z",
        "author": { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" }
      }
    ],
    "total": 100,
    "hasMore": true
  }
}
```

### POST `/api/v1/users/me`（需要登录）

修改个人资料。

请求头：

- `x-jwt-token: <access_token>`

请求体（字段均可选）：

```json
{
  "email": "user@example.com",
  "avatar": "https://example.com/avatar.png",
  "bio": "个人简介",
  "gender": 1,
  "birthday": "2006-01-02",
  "location": "Shanghai",
  "country": "China"
}
```

响应：

```json
{
  "code": 0,
  "msg": "修改个人资料成功"
}
```

说明：

- `birthday` 入参格式为 `YYYY-MM-DD`（后端按 `2006-01-02` 解析）。

### POST `/api/v1/users/me/password`（需要登录）

修改密码。

请求头：

- `x-jwt-token: <access_token>`

请求体：

```json
{
  "old_password": "32位字符串",
  "new_password": "32位字符串"
}
```

响应：

```json
{
  "code": 0,
  "msg": "密码修改成功"
}
```

常见错误：

- `20008` 旧密码错误（HTTP 401）

## 关注模块

### GET `/api/v1/users/me/followers`（需要登录）

按页获取“关注我的人”。

请求头：

- `x-jwt-token: <access_token>`

Query 参数：

- `pageNo`（默认 `1`）
- `pageSize`（默认 `10`，限制 `<= 100`）

响应 `data`：

```json
{
  "code": 0,
  "msg": "获取粉丝列表成功",
  "data": {
    "followers": [
      { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" }
    ],
    "total": 10,
    "hasMore": false
  }
}
```

### GET `/api/v1/users/me/followees`（需要登录）

按页获取“我关注的人”。

请求头：

- `x-jwt-token: <access_token>`

Query 参数：

- `pageNo`（默认 `1`）
- `pageSize`（默认 `10`，限制 `<= 100`）

响应 `data`：

```json
{
  "code": 0,
  "msg": "获取关注列表成功",
  "data": {
    "followees": [
      { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" }
    ],
    "total": 10,
    "hasMore": false
  }
}
```

### POST `/api/v1/users/:id/follow`（需要登录）

关注某个用户。

路径参数：

- `id`：对方用户 ID

请求头：

- `x-jwt-token: <access_token>`

响应：

```json
{
  "code": 0,
  "msg": "关注成功"
}
```

常见错误：

- `60001` 已经关注过该用户（HTTP 409）

### DELETE `/api/v1/users/:id/follow`（需要登录）

取消关注。

路径参数：

- `id`：对方用户 ID

请求头：

- `x-jwt-token: <access_token>`

响应：

```json
{
  "code": 0,
  "msg": "取消关注成功"
}
```

常见错误：

- `60002` 尚未关注，无法取消（HTTP 409）

### GET `/api/v1/users/:id/follow`（需要登录）

查询关注关系。

路径参数：

- `id`：对方用户 ID

请求头：

- `x-jwt-token: <access_token>`

响应 `data`：关注关系（数字枚举）

- `0`：互不关注
- `1`：我关注了对方
- `2`：对方关注了我
- `3`：互相关注

```json
{
  "code": 0,
  "msg": "获取关注关系成功",
  "data": 3
}
```

## 帖子模块

### GET `/api/v1/posts`

按页获取帖子列表（包含正文、浏览/点赞/评论计数等）。

Query 参数：

- `pageNo`（默认 `1`）
- `pageSize`（默认 `10`）

响应 `data`：

```json
{
  "code": 0,
  "msg": "获取帖子列表成功",
  "data": {
    "posts": [
      {
        "id": "1999760900969463808",
        "view_count": 10,
        "like_count": 3,
        "comment_count": 2,
        "title": "标题",
        "content": "正文",
        "created_at": "2024-01-01T00:00:00Z",
        "author": { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" },
        "tags": ["go", "gin"]
      }
    ],
    "total": 100,
    "hasMore": true
  }
}
```

### GET `/api/v1/posts/tags`

按标签分页获取帖子列表。

Query 参数：

- `tag`：标签名
- `pageNo`（默认 `1`）
- `pageSize`（默认 `10`）

响应结构同 `GET /api/v1/posts`。

### GET `/api/v1/posts/:id`

获取帖子详情（后端会对该帖 `view_count + 1`）。

路径参数：

- `id`：帖子 ID

响应 `data`：帖子详情（结构同上单个 `post`）

```json
{
  "code": 0,
  "msg": "获取帖子详情成功",
  "data": {
    "id": "1999760900969463808",
    "view_count": 11,
    "like_count": 3,
    "comment_count": 2,
    "title": "标题",
    "content": "正文",
    "created_at": "2024-01-01T00:00:00Z",
    "author": { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" },
    "tags": ["go", "gin"]
  }
}
```

### POST `/api/v1/posts`（需要登录）

创建帖子。

请求头：

- `x-jwt-token: <access_token>`

请求体：

```json
{
  "title": "标题",
  "content": "正文",
  "tags": ["go", "gin"]
}
```

响应 `data`：新建帖子详情（注意：当前实现可能不会在该响应里回填 `tags` 字段）

```json
{
  "code": 0,
  "msg": "帖子创建成功",
  "data": {
    "id": "1999760900969463808",
    "view_count": 0,
    "like_count": 0,
    "comment_count": 0,
    "title": "标题",
    "content": "正文",
    "created_at": "2024-01-01T00:00:00Z",
    "author": { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" },
    "tags": null
  }
}
```

### POST `/api/v1/posts/:id`（需要登录）

更新帖子（同时会进行标签绑定/解绑差异更新）。

路径参数：

- `id`：帖子 ID

请求头：

- `x-jwt-token: <access_token>`

请求体：

```json
{
  "title": "新标题",
  "content": "新正文",
  "tags": ["go", "gin"]
}
```

响应：

```json
{
  "code": 0,
  "msg": "帖子更新成功"
}
```

常见错误：

- `20006` 没有权限（HTTP 403）
- `30001` 帖子不存在（HTTP 404）

### DELETE `/api/v1/posts/:id`（需要登录）

删除帖子。

路径参数：

- `id`：帖子 ID

请求头：

- `x-jwt-token: <access_token>`

响应：

```json
{
  "code": 0,
  "msg": "帖子删除成功"
}
```

常见错误：

- `20006` 没有权限（HTTP 403）

## 评论模块

### GET `/api/v1/posts/:id/comments`

按页获取帖子评论列表。

路径参数：

- `id`：帖子 ID

Query 参数：

- `pageNo`（默认 `1`）
- `pageSize`（默认 `10`，限制 `<= 100`）

响应 `data`：

```json
{
  "code": 0,
  "msg": "获取评论列表成功",
  "data": {
    "comments": [
      {
        "id": "1999760900969463808",
        "post_id": "1999760900969463808",
        "parent_id": "0",
        "reply_id": "0",
        "content": "评论内容",
        "created_at": "2024-01-01T00:00:00Z",
        "author": { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" }
      }
    ],
    "total": 10,
    "hasMore": false
  }
}
```

### GET `/api/v1/posts/:id/comments/:cid`

获取某条评论的回复列表。

路径参数：

- `id`：帖子 ID
- `cid`：评论 ID

响应 `data`：回复数组（结构同 Comment DTO）

```json
{
  "code": 0,
  "msg": "获取评论回复列表成功",
  "data": [
    {
      "id": "1999760900969463808",
      "post_id": "1999760900969463808",
      "parent_id": "1999760900969463808",
      "reply_id": "0",
      "content": "回复内容",
      "created_at": "2024-01-01T00:00:00Z",
      "author": { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" }
    }
  ]
}
```

### POST `/api/v1/posts/:id/comments`（需要登录）

创建评论/回复。

路径参数：

- `id`：帖子 ID

请求头：

- `x-jwt-token: <access_token>`

请求体：

```json
{
  "parent_id": "0",
  "reply_id": "0",
  "content": "评论内容"
}
```

说明：

- `parent_id=0` 表示一级评论；`parent_id!=0` 表示对某条评论的回复（回复列表通过 `GET /api/v1/posts/:id/comments/:cid` 获取）。

响应 `data`：新建评论

```json
{
  "code": 0,
  "msg": "评论成功",
  "data": {
    "id": "1999760900969463808",
    "post_id": "1999760900969463808",
    "parent_id": "0",
    "reply_id": "0",
    "content": "评论内容",
    "created_at": "2024-01-01T00:00:00Z",
    "author": { "id": "1999760900969463808", "email": "user@example.com", "name": "alice", "avatar": "" }
  }
}
```

### DELETE `/api/v1/posts/:id/comments/:cid`（需要登录）

删除评论（帖子作者可删除该帖下任意评论；评论作者可删除自己的评论）。

路径参数：

- `id`：帖子 ID
- `cid`：评论 ID

请求头：

- `x-jwt-token: <access_token>`

响应：

```json
{
  "code": 0,
  "msg": "评论删除成功"
}
```

常见错误：

- `20006` 没有权限（HTTP 403）
- `40001` 评论不存在（HTTP 404）

## 点赞模块

### GET `/api/v1/posts/:id/likes`（需要登录）

查询是否点赞了该帖子。

路径参数：

- `id`：帖子 ID

请求头：

- `x-jwt-token: <access_token>`

响应 `data`：`true/false`

```json
{
  "code": 0,
  "msg": "success",
  "data": true
}
```

### POST `/api/v1/posts/:id/likes`（需要登录）

点赞帖子。

路径参数：

- `id`：帖子 ID

请求头：

- `x-jwt-token: <access_token>`

响应：

```json
{
  "code": 0,
  "msg": "success"
}
```

常见错误：

- `30001` 帖子不存在（HTTP 404）
- `30002` 已经点赞过该帖子（HTTP 409）

### DELETE `/api/v1/posts/:id/likes`（需要登录）

取消点赞。

路径参数：

- `id`：帖子 ID

请求头：

- `x-jwt-token: <access_token>`

响应：

```json
{
  "code": 0,
  "msg": "success"
}
```

常见错误：

- `30001` 帖子不存在（HTTP 404）
- `30003` 尚未点赞，无法取消（HTTP 409）

---

## 统一错误码（`errno/errors.go`）

通用：

- `10001` 系统繁忙，请稍后重试（HTTP 500）
- `10002` 参数错误（HTTP 400）

用户：

- `20001` 用户不存在（HTTP 404）
- `20002` 用户名或邮箱已存在（HTTP 409）
- `20004` 账号或密码错误（HTTP 401）
- `20005` 用户未登录（HTTP 401）
- `20006` 没有权限（HTTP 403）
- `20008` 旧密码错误（HTTP 401）

帖子：

- `30001` 帖子不存在（HTTP 404）
- `30002` 已经点赞过该帖子（HTTP 409）
- `30003` 尚未点赞，无法取消（HTTP 409）

评论：

- `40001` 评论不存在（HTTP 404）

关注：

- `60001` 已经关注过该用户（HTTP 409）
- `60002` 尚未关注，无法取消（HTTP 409）
