# API 文档

## 基础信息

- Base URL: http://localhost:8765
- API 前缀: /api/v1
- Content-Type: application/json
- 时间格式: RFC3339
- ID 字段：JSON 输出为 string（字段 tag 为 `json:",string"`）

## 认证

- AccessToken: `Authorization: Bearer <token>`
- AccessToken 也可放在 Cookie `x-jwt-token`（主要用于 WS/HTTP 读取）
- RefreshToken: Cookie `refresh-token`
- 注册/登录成功会返回 `Authorization` Header 并设置 `refresh-token` Cookie
- Auth 中间件失败时直接返回 HTTP 401（无统一响应体），并清空 token

## 统一响应

成功:

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

失败:

```json
{
  "code": 10002,
  "msg": "参数错误"
}
```

说明:

- data 字段为空时可能被省略

## 错误码

| Code  | HTTP | 说明 |
| ----- | ---- | ---- |
| 0     | 200  | 成功 |
| 10001 | 500  | 系统繁忙，请稍后重试 |
| 10002 | 400  | 参数错误 |
| 20001 | 404  | 用户不存在 |
| 20002 | 409  | 用户名或邮箱已存在 |
| 20003 | 400  | 密码强度过低 |
| 20004 | 401  | 账号或密码错误 |
| 20005 | 401  | 用户未登录 |
| 20006 | 403  | 没有权限 |
| 20007 | 500  | 登出失败 |
| 20008 | 401  | 旧密码错误 |
| 30001 | 404  | 帖子不存在 |
| 30002 | 409  | 已经点赞过该帖子 |
| 30003 | 409  | 尚未点赞，无法取消 |
| 40001 | 404  | 评论不存在 |
| 50001 | 409  | 标签重复绑定 |
| 60001 | 409  | 已经关注过该用户 |
| 60002 | 409  | 尚未关注，无法取消 |
| 80001 | 404  | 奖品不存在 |
| 80002 | 404  | 没有抢到该商品，或支付时限已过 |
| 80003 | 404  | 订单不存在 |
| 80004 | 404  | 没有抽到奖品 |

## 数据模型

### UserBrief

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | string | 用户 ID |
| email | string | 邮箱 |
| name | string | 用户名 |
| avatar | string | 头像 URL |

### UserDetail

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | string | 用户 ID |
| name | string | 用户名 |
| email | string | 邮箱 |
| avatar | string | 头像 URL |
| bio | string | 个性签名 |
| gender | int | 0=空，1=男，2=女，3=其它 |
| birthday | string | 生日（RFC3339，可能为空） |
| location | string | 地区 |
| country | string | 国家 |
| last_login_ip | string | 最近登录 IP |

### PostDetail

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | string | 帖子 ID |
| view_count | int | 浏览数 |
| like_count | int | 点赞数 |
| comment_count | int | 评论数 |
| title | string | 标题 |
| content | string | 内容 |
| created_at | string | 创建时间（RFC3339） |
| author | UserBrief | 作者 |
| tags | string[] | 标签 |

### PostBrief

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | string | 帖子 ID |
| title | string | 标题 |
| created_at | string | 创建时间（RFC3339） |
| author | UserBrief | 作者 |

### PostTop

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | string | 帖子 ID |
| title | string | 标题 |
| score | float | 热度得分 |

### Comment

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | string | 评论 ID |
| post_id | string | 帖子 ID |
| parent_id | string | 父评论 ID |
| reply_id | string | 回复目标评论 ID |
| content | string | 内容 |
| created_at | string | 创建时间（RFC3339） |
| author | UserBrief | 作者 |

### Session

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| session_id | string | 会话 ID |
| target_id | string | 对方用户 ID |
| target_name | string | 对方用户名 |
| target_avatar | string | 对方头像 |
| last_message_id | string | 最后一条消息 ID |
| last_message | string | 最后一条消息摘要 |
| last_message_time | string | 最后一条消息时间（RFC3339） |
| unread_count | int | 未读数 |

### Message

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| content | string | 消息内容 |
| message_from | string | 发送方 ID |
| message_to | string | 接收方 ID |
| id | string | 消息 ID |
| session_id | string | 会话 ID |
| session_type | int | 会话类型 |
| created_at | string | 创建时间（RFC3339） |

### Gift

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | string | 奖品 ID |
| name | string | 奖品名称 |
| avatar | string | 奖品图片 |
| description | string | 奖品描述 |
| prize | string | 奖品说明 |

### LotteryOrder

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | string | 订单 ID |
| user | UserBrief | 用户信息 |
| gift | Gift | 奖品信息 |
| count | int | 购买数量 |
| created_at | string | 创建时间（RFC3339） |

### FollowType

| 值 | 说明 |
| --- | ---- |
| 0 | 互不关注 |
| 1 | 我关注对方 |
| 2 | 对方关注我 |
| 3 | 互相关注 |

## 接口

### 认证 Auth

#### POST /api/v1/auth/register

- Auth: 否
- Body:
  - email (string, 可选)
  - name (string, 必填, 长度 >= 2)
  - password (string, 必填, 长度 = 32)
- Response: UserBrief
- Notes: 返回 `Authorization` Header 并设置 `refresh-token` Cookie

示例请求:

```bash
curl -i -X POST "http://localhost:8765/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "name": "alice",
    "password": "0123456789abcdef0123456789abcdef"
  }'
```

示例响应:

```http
HTTP/1.1 200 OK
Authorization: Bearer <access_token>
Set-Cookie: refresh-token=<refresh_token>; Path=/; HttpOnly
Content-Type: application/json

{
  "code": 0,
  "msg": "注册成功",
  "data": {
    "id": "1001",
    "email": "alice@example.com",
    "name": "alice",
    "avatar": ""
  }
}
```

#### POST /api/v1/auth/login

- Auth: 否
- Body:
  - name (string, 必填, 长度 >= 2)
  - password (string, 必填, 长度 = 32)
- Response: UserBrief
- Notes: 返回 `Authorization` Header 并设置 `refresh-token` Cookie

示例请求:

```bash
curl -i -X POST "http://localhost:8765/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "alice",
    "password": "0123456789abcdef0123456789abcdef"
  }'
```

示例响应:

```http
HTTP/1.1 200 OK
Authorization: Bearer <access_token>
Set-Cookie: refresh-token=<refresh_token>; Path=/; HttpOnly
Content-Type: application/json

{
  "code": 0,
  "msg": "登录成功",
  "data": {
    "id": "1001",
    "email": "alice@example.com",
    "name": "alice",
    "avatar": ""
  }
}
```

#### POST /api/v1/auth/logout

- Auth: 是
- Response: null

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/auth/logout" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "登出成功"
}
```

#### GET /api/v1/auth/status

- Auth: 是
- Response: null

示例请求:

```bash
curl "http://localhost:8765/api/v1/auth/status" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "登录状态检查成功"
}
```

### 用户 Users

#### GET /api/v1/users/:id

- Auth: 否
- Response: UserDetail

示例请求:

```bash
curl "http://localhost:8765/api/v1/users/1001"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取个人资料成功",
  "data": {
    "id": "1001",
    "name": "alice",
    "email": "alice@example.com",
    "avatar": "https://example.com/avatar.png",
    "bio": "hello",
    "gender": 1,
    "birthday": "2024-01-01T00:00:00Z",
    "location": "shanghai",
    "country": "cn",
    "last_login_ip": "127.0.0.1"
  }
}
```

#### GET /api/v1/users/:id/posts

- Auth: 否
- Query:
  - pageNo (int, 默认 1)
  - pageSize (int, 默认 10, 最大 100)
- Response:
  - posts: PostBrief[]
  - total: int
  - hasMore: bool

示例请求:

```bash
curl "http://localhost:8765/api/v1/users/1001/posts?pageNo=1&pageSize=10"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取帖子列表成功",
  "data": {
    "posts": [
      {
        "id": "2001",
        "title": "hello world",
        "created_at": "2024-01-02T15:04:05Z",
        "author": {
          "id": "1001",
          "email": "alice@example.com",
          "name": "alice",
          "avatar": ""
        }
      }
    ],
    "total": 1,
    "hasMore": false
  }
}
```

#### POST /api/v1/users/me

- Auth: 是
- Body: ModifyProfileRequest
  - email (string, 可选)
  - avatar (string, 可选)
  - bio (string, 可选)
  - gender (int, 可选)
  - birthday (string, 可选, 格式 yyyy-mm-dd)
  - location (string, 可选)
  - country (string, 可选)
- Response: null

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/users/me" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "hello",
    "gender": 1,
    "birthday": "2024-01-01",
    "location": "shanghai",
    "country": "cn"
  }'
```

示例响应:

```json
{
  "code": 0,
  "msg": "修改个人资料成功"
}
```

#### POST /api/v1/users/me/password

- Auth: 是
- Body:
  - old_password (string, 必填, 长度 = 32)
  - new_password (string, 必填, 长度 = 32)
- Response: null

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/users/me/password" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "0123456789abcdef0123456789abcdef",
    "new_password": "abcdef0123456789abcdef0123456789"
  }'
```

示例响应:

```json
{
  "code": 0,
  "msg": "密码修改成功"
}
```

#### GET /api/v1/users/me/followers

- Auth: 是
- Query:
  - pageNo (int, 默认 1)
  - pageSize (int, 默认 10, 最大 100)
- Response:
  - followers: UserBrief[]
  - total: int
  - hasMore: bool

示例请求:

```bash
curl "http://localhost:8765/api/v1/users/me/followers?pageNo=1&pageSize=10" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取粉丝列表成功",
  "data": {
    "followers": [
      {
        "id": "1002",
        "email": "bob@example.com",
        "name": "bob",
        "avatar": ""
      }
    ],
    "total": 1,
    "hasMore": false
  }
}
```

#### GET /api/v1/users/me/followees

- Auth: 是
- Query:
  - pageNo (int, 默认 1)
  - pageSize (int, 默认 10, 最大 100)
- Response:
  - followees: UserBrief[]
  - total: int
  - hasMore: bool

示例请求:

```bash
curl "http://localhost:8765/api/v1/users/me/followees?pageNo=1&pageSize=10" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取关注列表成功",
  "data": {
    "followees": [
      {
        "id": "1002",
        "email": "bob@example.com",
        "name": "bob",
        "avatar": ""
      }
    ],
    "total": 1,
    "hasMore": false
  }
}
```

#### POST /api/v1/users/:id/follow

- Auth: 是
- Response: null

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/users/1002/follow" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "关注成功"
}
```

#### DELETE /api/v1/users/:id/follow

- Auth: 是
- Response: null

示例请求:

```bash
curl -X DELETE "http://localhost:8765/api/v1/users/1002/follow" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "取消关注成功"
}
```

#### GET /api/v1/users/:id/follow

- Auth: 是
- Response: FollowType

示例请求:

```bash
curl "http://localhost:8765/api/v1/users/1002/follow" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取关注关系成功",
  "data": 1
}
```

#### GET /api/v1/users/:id/sessions

- Auth: 是
- Response: Session

示例请求:

```bash
curl "http://localhost:8765/api/v1/users/1002/sessions" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取会话成功",
  "data": {
    "session_id": "4001",
    "target_id": "1002",
    "target_name": "bob",
    "target_avatar": "",
    "last_message_id": "5001",
    "last_message": "hello",
    "last_message_time": "2024-01-02T15:04:05Z",
    "unread_count": 2
  }
}
```

#### GET /api/v1/users/:id/sessions/messages

- Auth: 是
- Query:
  - pageNo (int, 默认 1)
  - pageSize (int, 默认 5)
- Response:
  - total: int
  - has_more: bool
  - messages: Message[]

示例请求:

```bash
curl "http://localhost:8765/api/v1/users/1002/sessions/messages?pageNo=1&pageSize=5" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取聊天记录成功",
  "data": {
    "total": 1,
    "has_more": false,
    "messages": [
      {
        "content": "hello",
        "message_from": "1001",
        "message_to": "1002",
        "id": "5001",
        "session_id": "4001",
        "session_type": 1,
        "created_at": "2024-01-02T15:04:05Z"
      }
    ]
  }
}
```

### 帖子 Posts

#### GET /api/v1/posts

- Auth: 否
- Query:
  - pageNo (int, 默认 1)
  - pageSize (int, 默认 10)
- Response:
  - posts: PostDetail[]
  - total: int
  - hasMore: bool

示例请求:

```bash
curl "http://localhost:8765/api/v1/posts?pageNo=1&pageSize=10"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取帖子列表成功",
  "data": {
    "posts": [
      {
        "id": "2001",
        "view_count": 10,
        "like_count": 2,
        "comment_count": 1,
        "title": "hello world",
        "content": "first user",
        "created_at": "2024-01-02T15:04:05Z",
        "author": {
          "id": "1001",
          "email": "alice@example.com",
          "name": "alice",
          "avatar": ""
        },
        "tags": ["go", "gin"]
      }
    ],
    "total": 1,
    "hasMore": false
  }
}
```

#### GET /api/v1/posts/top

- Auth: 否
- Response: PostTop[]

示例请求:

```bash
curl "http://localhost:8765/api/v1/posts/top"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取热门帖子榜单成功",
  "data": [
    {
      "id": "2001",
      "title": "hello world",
      "score": 12.3
    }
  ]
}
```

#### GET /api/v1/posts/tags

- Auth: 否
- Query:
  - tag (string, 必填)
  - pageNo (int, 默认 1)
  - pageSize (int, 默认 10)
- Response:
  - posts: PostDetail[]
  - total: int
  - hasMore: bool

示例请求:

```bash
curl "http://localhost:8765/api/v1/posts/tags?tag=go&pageNo=1&pageSize=10"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取帖子列表成功",
  "data": {
    "posts": [
      {
        "id": "2001",
        "view_count": 10,
        "like_count": 2,
        "comment_count": 1,
        "title": "hello world",
        "content": "first user",
        "created_at": "2024-01-02T15:04:05Z",
        "author": {
          "id": "1001",
          "email": "alice@example.com",
          "name": "alice",
          "avatar": ""
        },
        "tags": ["go"]
      }
    ],
    "total": 1,
    "hasMore": false
  }
}
```

#### GET /api/v1/posts/:id

- Auth: 否
- Response: PostDetail

示例请求:

```bash
curl "http://localhost:8765/api/v1/posts/2001"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取帖子详情成功",
  "data": {
    "id": "2001",
    "view_count": 11,
    "like_count": 2,
    "comment_count": 1,
    "title": "hello world",
    "content": "first user",
    "created_at": "2024-01-02T15:04:05Z",
    "author": {
      "id": "1001",
      "email": "alice@example.com",
      "name": "alice",
      "avatar": ""
    },
    "tags": ["go", "gin"]
  }
}
```

#### GET /api/v1/posts/:id/comments

- Auth: 否
- Query:
  - pageNo (int, 默认 1)
  - pageSize (int, 默认 10, 最大 100)
- Response:
  - comments: Comment[]
  - total: int
  - hasMore: bool

示例请求:

```bash
curl "http://localhost:8765/api/v1/posts/2001/comments?pageNo=1&pageSize=10"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取评论列表成功",
  "data": {
    "comments": [
      {
        "id": "3001",
        "post_id": "2001",
        "parent_id": "0",
        "reply_id": "0",
        "content": "nice",
        "created_at": "2024-01-02T15:04:05Z",
        "author": {
          "id": "1002",
          "email": "bob@example.com",
          "name": "bob",
          "avatar": ""
        }
      }
    ],
    "total": 1,
    "hasMore": false
  }
}
```

#### GET /api/v1/posts/:id/comments/:cid

- Auth: 否
- Query:
  - pageNo (int, 默认 1)
  - pageSize (int, 默认 3, 最大 100)
- Response:
  - comments: Comment[]
  - total: int
  - hasMore: bool

示例请求:

```bash
curl "http://localhost:8765/api/v1/posts/2001/comments/3001?pageNo=1&pageSize=3"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取评论回复列表成功",
  "data": {
    "comments": [
      {
        "id": "3002",
        "post_id": "2001",
        "parent_id": "3001",
        "reply_id": "3001",
        "content": "reply",
        "created_at": "2024-01-02T15:04:05Z",
        "author": {
          "id": "1001",
          "email": "alice@example.com",
          "name": "alice",
          "avatar": ""
        }
      }
    ],
    "total": 1,
    "hasMore": false
  }
}
```

#### POST /api/v1/posts

- Auth: 是
- Body:
  - title (string, 必填, 长度 >= 1)
  - content (string, 必填, 长度 >= 1)
  - tags (string[], 可选)
- Response: PostDetail

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/posts" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "hello world",
    "content": "first user",
    "tags": ["go", "gin"]
  }'
```

示例响应:

```json
{
  "code": 0,
  "msg": "帖子创建成功",
  "data": {
    "id": "2001",
    "view_count": 0,
    "like_count": 0,
    "comment_count": 0,
    "title": "hello world",
    "content": "first user",
    "created_at": "2024-01-02T15:04:05Z",
    "author": {
      "id": "1001",
      "email": "alice@example.com",
      "name": "alice",
      "avatar": ""
    },
    "tags": ["go", "gin"]
  }
}
```

#### POST /api/v1/posts/:id

- Auth: 是
- Body:
  - title (string, 必填, 长度 >= 1)
  - content (string, 必填, 长度 >= 1)
  - tags (string[], 可选)
- Response: null

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/posts/2001" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "hello world v2",
    "content": "updated",
    "tags": ["go"]
  }'
```

示例响应:

```json
{
  "code": 0,
  "msg": "帖子更新成功"
}
```

#### DELETE /api/v1/posts/:id

- Auth: 是
- Response: null

示例请求:

```bash
curl -X DELETE "http://localhost:8765/api/v1/posts/2001" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "帖子删除成功"
}
```

#### POST /api/v1/posts/:id/comments

- Auth: 是
- Body:
  - parent_id (string, 可选)
  - reply_id (string, 可选)
  - content (string, 必填, 长度 >= 1)
- Response: Comment

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/posts/2001/comments" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "0",
    "reply_id": "0",
    "content": "nice"
  }'
```

示例响应:

```json
{
  "code": 0,
  "msg": "评论成功",
  "data": {
    "id": "3001",
    "post_id": "2001",
    "parent_id": "0",
    "reply_id": "0",
    "content": "nice",
    "created_at": "2024-01-02T15:04:05Z",
    "author": {
      "id": "1001",
      "email": "alice@example.com",
      "name": "alice",
      "avatar": ""
    }
  }
}
```

#### DELETE /api/v1/posts/:id/comments/:cid

- Auth: 是
- Response: null

示例请求:

```bash
curl -X DELETE "http://localhost:8765/api/v1/posts/2001/comments/3001" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "评论删除成功"
}
```

#### GET /api/v1/posts/:id/likes

- Auth: 是
- Response: bool

示例请求:

```bash
curl "http://localhost:8765/api/v1/posts/2001/likes" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "success",
  "data": true
}
```

#### POST /api/v1/posts/:id/likes

- Auth: 是
- Response: null

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/posts/2001/likes" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "success"
}
```

#### DELETE /api/v1/posts/:id/likes

- Auth: 是
- Response: null

示例请求:

```bash
curl -X DELETE "http://localhost:8765/api/v1/posts/2001/likes" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "success"
}
```

### 抽奖 Lottery

#### GET /api/v1/gifts

- Auth: 否
- Response: Gift[]
- Notes: 仅返回奖品基础信息，不包含库存或概率

示例请求:

```bash
curl "http://localhost:8765/api/v1/gifts"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取全部奖品成功",
  "data": [
    {
    "id": "3",
    "name": "论坛定制书",
    "avatar": "https://example.com/gifts/mug.png",
    "description": "限量 500 份的定制书",
    "prize": 200
    },
    {
    "id": "3",
    "name": "论坛定制书",
    "avatar": "https://example.com/gifts/mug.png",
    "description": "限量 500 份的定制书",
    "prize": 200
    }
  ]
}
```

#### GET /api/v1/lottery/lucky

- Auth: 是
- Response: Gift
- Notes:
  - `name` 为 `谢谢参与` 或 `id` 为 `0` 时表示未中奖/奖品已抽完，不生成临时订单
  - 中奖后需在支付时限内调用 `/api/v1/lottery/pay` 或 `/api/v1/lottery/giveup`（默认 600 秒）

示例请求:

```bash
curl "http://localhost:8765/api/v1/lottery/lucky" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "抽奖成功",
  "data": {
    "id": "3",
    "name": "论坛定制书",
    "avatar": "https://example.com/gifts/mug.png",
    "description": "限量 500 份的定制书",
    "prize": 200
  }
}
```

#### POST /api/v1/lottery/pay

- Auth: 是
- Body: PayRequest
  - user_id (string, 必填)
  - gift_id (string, 必填)
- Response: null
- Notes: user_id 必须与当前登录用户一致，且需命中临时订单，否则返回 80002

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/lottery/pay" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "1001",
    "gift_id": "3"
  }'
```

示例响应:

```json
{
  "code": 0,
  "msg": "支付成功"
}
```

#### POST /api/v1/lottery/giveup

- Auth: 是
- Body: GiveUpRequest
  - user_id (string, 必填)
  - gift_id (string, 必填)
- Response: null
- Notes: user_id 必须与当前登录用户一致，且需命中临时订单，否则返回 80002

示例请求:

```bash
curl -X POST "http://localhost:8765/api/v1/lottery/giveup" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "1001",
    "gift_id": "3"
  }'
```

示例响应:

```json
{
  "code": 0,
  "msg": "放弃支付成功"
}
```

#### GET /api/v1/lottery/result

- Auth: 是
- Response: LotteryOrder
- Notes: 返回当前用户最近一次支付成功的订单，若不存在返回 80003

示例请求:

```bash
curl "http://localhost:8765/api/v1/lottery/result" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取结果成功",
  "data": {
    "id": "9001",
    "user": {
      "id": "1001",
      "email": "alice@example.com",
      "name": "alice",
      "avatar": ""
    },
    "gift": {
    "id": "3",
    "name": "论坛定制书",
    "avatar": "https://example.com/gifts/mug.png",
    "description": "限量 500 份的定制书",
    "prize": 200
    },
    "count": 1,
    "created_at": "2024-01-02T15:04:05Z"
  }
}
```

### 会话 Sessions

#### GET /api/v1/sessions

- Auth: 是
- Response: Session[]

示例请求:

```bash
curl "http://localhost:8765/api/v1/sessions" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "获取会话列表成功",
  "data": [
    {
      "session_id": "4001",
      "target_id": "1002",
      "target_name": "bob",
      "target_avatar": "",
      "last_message_id": "5001",
      "last_message": "hello",
      "last_message_time": "2024-01-02T15:04:05Z",
      "unread_count": 2
    }
  ]
}
```

#### DELETE /api/v1/sessions/:id

- Auth: 是
- Response: null

示例请求:

```bash
curl -X DELETE "http://localhost:8765/api/v1/sessions/4001" \
  -H "Authorization: Bearer <access_token>"
```

示例响应:

```json
{
  "code": 0,
  "msg": "删除会话成功"
}
```

### WebSocket

#### GET /api/v1/ws

- Auth: 是
- 协议: ws://localhost:8765/api/v1/ws
- Client -> Server (JSON):
  - type = "message"
    - session_id (string)
    - session_type (int)
    - message_from (string)
    - message_to (string)
    - content (string)
    - user_id (string, 可选)
  - type = "read_ack"
    - session_id (string)
    - user_id (string, 可选)
- Server -> Client:
  - Message DTO（见「数据模型」）

示例请求:

```bash
wscat -c "ws://localhost:8765/api/v1/ws" -H "Authorization: Bearer <access_token>"
```

发送消息示例:

```json
{
  "type": "message",
  "session_id": "4001",
  "session_type": 1,
  "message_from": "1001",
  "message_to": "1002",
  "content": "hello"
}
```

服务端推送示例:

```json
{
  "content": "hello",
  "message_from": "1001",
  "message_to": "1002",
  "id": "5001",
  "session_id": "4001",
  "session_type": 1,
  "created_at": "2024-01-02T15:04:05Z"
}
```

### 运维

#### GET /metrics

- Auth: 否
- Description: Prometheus metrics

示例请求:

```bash
curl "http://localhost:8765/metrics"
```

示例响应:

```text
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
```
