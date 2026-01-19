# Go-Postery API

## 基本信息
- Base URL: http://localhost:8765
- Base Path: /api/v1
- Content-Type: application/json
- ID 字段: 多数字段使用 JSON 字符串表示（例如 "id":"123"）
- 时间字段: RFC3339（例如 "2024-01-02T15:04:05Z"）；修改资料的 birthday 请求为 YYYY-MM-DD

## 统一响应格式
```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```
- code=0 表示成功，非 0 表示失败
- msg 为空时默认 "success"
- data 失败时通常为 null

## 认证
- AccessToken: 请求头 `Authorization: Bearer <token>`；WS 可使用 Cookie `x-jwt-token`
- RefreshToken: Cookie `refresh-token`（HttpOnly）
- 登录成功会在响应头写入 `Authorization`，并通过 `Set-Cookie` 写入 refresh-token
- AuthRequired 的接口在鉴权失败时可能直接返回 HTTP 401（无 JSON body）

## 错误码
### 统一响应码
- 0: 成功
- 40001: 参数错误
- 40003: 未登录/无权限
- 50001: 服务端错误

### 业务错误码（errno）
- 10001: 系统繁忙，请稍后重试
- 10002: 参数错误
- 20001: 用户不存在
- 20002: 用户已存在
- 30001: 帖子不存在
- 30002: 已经点赞过该帖子
- 30003: 尚未点赞，无法取消
- 40001: 评论不存在
- 50001: 标签重复绑定
- 60001: 已经关注过该用户
- 60002: 尚未关注，无法取消
- 70001: 验证码验证失败
- 70002: 验证码发送过于频繁
- 70003: 验证码不存在
- 70004: 邮箱或验证码错误
- 70005: 手机号或验证码错误
- 70006: 两次密码一致
- 70007: 密码强度过低
- 70008: 账号或密码错误
- 70009: 没有权限
- 70010: 登出失败
- 70011: 旧密码错误
- 70012: 初始化密码失败
- 70013: 用户未登录
- 80001: 奖品不存在
- 80002: 没有抢到该商品，或支付时限已过
- 80003: 订单不存在
- 80004: 没有抽到奖品
- 注意: 40001/50001 在统一响应码与业务错误码中存在重复，请结合 msg 与 HTTP 状态码判断

## 数据结构
### UserBrief
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 用户 ID |
| nickname | string | 昵称 |
| avatar | string | 头像 URL |

### UserDetail
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 用户 ID |
| nickname | string | 用户名 |
| avatar | string | 头像 URL |
| bio | string | 个性签名 |
| gender | int | 性别 |
| birthday | string | 生日（RFC3339） |
| location | string | 地区 |
| country | string | 国家 |
| last_login_ip | string | 最近一次登录 IP |

### UserTop
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 用户 ID |
| nickname | string | 昵称 |
| bio | string | 个性签名 |
| avatar | string | 头像 URL |
| score | number | 推荐分 |

### PostDetail
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 帖子 ID |
| view_count | int | 浏览量 |
| like_count | int | 点赞数 |
| comment_count | int | 评论数 |
| title | string | 标题 |
| content | string | 正文 |
| content_type | int | 正文类型 |
| created_at | string | 创建时间（RFC3339） |
| author | UserBrief | 作者 |
| tags | string[] | 标签列表 |

### PostBrief
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 帖子 ID |
| title | string | 标题 |
| created_at | string | 创建时间（RFC3339） |
| author | UserBrief | 作者 |

### PostTop
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 帖子 ID |
| title | string | 标题 |
| score | number | 热度分 |

### CommentDTO
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 评论 ID |
| post_id | string | 帖子 ID |
| parent_id | string | 父评论 ID |
| reply_id | string | 回复目标评论 ID |
| content | string | 内容 |
| created_at | string | 创建时间（RFC3339） |
| author | UserBrief | 作者 |

### SessionDTO
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| session_id | string | 会话 ID |
| target_id | string | 对方用户 ID |
| target_name | string | 对方昵称 |
| target_avatar | string | 对方头像 |
| last_message_id | string | 最后一条消息 ID |
| last_message | string | 最后一条消息摘要 |
| last_message_time | string | 最后一条消息时间（RFC3339） |
| unread_count | int | 未读数 |

### MessageDTO
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| content | string | 消息内容 |
| message_from | string | 发送方 ID |
| message_to | string | 接收方 ID |
| id | string | 消息 ID |
| session_id | string | 会话 ID |
| session_type | int | 会话类型 |
| created_at | string | 创建时间（RFC3339） |

### GiftDTO
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 奖品 ID |
| name | string | 奖品名称 |
| avatar | string | 奖品图片 |
| description | string | 描述 |
| prize | int | 奖品价格 |

### OrderDTO
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 订单 ID |
| user | UserBrief | 用户信息 |
| gift | GiftDTO | 奖品信息 |
| count | int | 购买数量 |
| created_at | string | 创建时间（RFC3339） |

### PassStatusResponse
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| has_password | bool | 是否已设置密码 |

### AuthIdentityResponse
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| phone | string | 手机号 |
| email | string | 邮箱 |

### AgentDTO
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| session_id | string | 会话 ID |
| content | string | 回复内容 |
| documents | string[] | 参考文档片段 |

## 枚举约定
### content_type
| 值 | 说明 |
| --- | --- |
| 0 | 普通文本 |
| 1 | Markdown 文本 |

### gender
| 值 | 说明 |
| --- | --- |
| 0 | 未设置 |
| 1 | 男 |
| 2 | 女 |
| 3 | 其它 |

### follow_type
| 值 | 说明 |
| --- | --- |
| 0 | 无关注关系 |
| 1 | 我关注对方 |
| 2 | 对方关注我 |
| 3 | 互相关注 |

### session_type
| 值 | 说明 |
| --- | --- |
| 1 | 私聊 |
| 2 | 群聊 |

## 请求体结构
### SendSMSCodeRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| phone | string | 是 | len=11 | 手机号 |

### SendEmailCodeRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| email | string | 是 | email | 邮箱 |

### LoginByPasswordRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| identifier | string | 是 | - | 手机号或邮箱 |
| password | string | 是 | - | 密码 |

### LoginByPhoneRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| phone | string | 是 | len=11 | 手机号 |
| code | string | 是 | - | 验证码 |

### UpdatePassRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| old_password | string | 是 | len=32 | 旧密码 |
| new_password | string | 是 | len=32 | 新密码 |

### SetPassRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| new_password | string | 是 | len=32 | 新密码 |
| code | string | 是 | - | 验证码 |

### ModifyProfileRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| nickname | string | 否 | - | 昵称 |
| avatar | string | 否 | - | 头像 URL |
| bio | string | 否 | - | 个性签名 |
| gender | int | 否 | - | 性别 |
| birthday | string | 否 | YYYY-MM-DD | 生日 |
| location | string | 否 | - | 地区 |
| country | string | 否 | - | 国家 |

### PostCreateRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| title | string | 是 | gte=1 | 标题 |
| content | string | 是 | gte=1 | 正文 |
| content_type | int | 是 | - | 正文类型 |
| tags | string[] | 否 | - | 标签 |

### PostUpdateRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| title | string | 是 | gte=1 | 标题 |
| content | string | 是 | gte=1 | 正文 |
| tags | string[] | 否 | - | 标签 |

### CommentCreateRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| parent_id | string | 否 | >=0 | 父评论 ID（顶级评论建议传 0） |
| reply_id | string | 否 | - | 回复评论 ID |
| content | string | 是 | gte=1 | 内容 |

### LotteryPayRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| user_id | string | 是 | - | 用户 ID |
| gift_id | string | 是 | - | 奖品 ID |

### LotteryGiveUpRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| user_id | string | 是 | - | 用户 ID |
| gift_id | string | 是 | - | 奖品 ID |

### AgentChatRequest
| 字段 | 类型 | 必填 | 约束 | 说明 |
| --- | --- | --- | --- | --- |
| session_id | string | 是 | 可解析为 int64 | 会话 ID |
| query | string | 否 | - | 问题内容 |

## 运维接口
### GET /metrics
- 认证: 否
- 响应: Prometheus 文本格式（非 JSON）

## 认证模块
### POST /api/v1/auth/sms
- 认证: 否
- 请求体: SendSMSCodeRequest
- 响应 data: null

### POST /api/v1/auth/email
- 认证: 否
- 请求体: SendEmailCodeRequest
- 响应 data: null

### POST /api/v1/auth/login/password
- 认证: 否
- 请求体: LoginByPasswordRequest
- 响应 data: UserBrief
- 备注: 成功时响应头 `Authorization` 写入 AccessToken，并设置 Cookie `refresh-token`

### POST /api/v1/auth/login/phone
- 认证: 否
- 请求体: LoginByPhoneRequest
- 响应 data: UserBrief
- 备注: 成功时响应头 `Authorization` 写入 AccessToken，并设置 Cookie `refresh-token`

### POST /api/v1/auth/logout
- 认证: 是
- 请求体: 无
- 响应 data: null
- 备注: 清空 `Authorization` 与 `refresh-token` Cookie

### GET /api/v1/auth/status
- 认证: 是
- 响应 data: null

### POST /api/v1/auth/password/update
- 认证: 是
- 请求体: UpdatePassRequest
- 响应 data: null

### POST /api/v1/auth/password/set
- 认证: 是
- 请求体: SetPassRequest
- 响应 data: null

### GET /api/v1/auth/password/status
- 认证: 是
- 响应 data: PassStatusResponse

### GET /api/v1/auth/auth_identity
- 认证: 是
- 响应 data: AuthIdentityResponse

## 用户模块
### GET /api/v1/users/:id
- 认证: 否
- Path: id (int64)
- 响应 data: UserDetail

### GET /api/v1/users/:id/posts
- 认证: 否
- Path: id (int64)
- Query: pageNo (int, default 1, min 1), pageSize (int, default 10, max 100)
- 响应 data: `{posts: PostBrief[], total: int, hasMore: bool}`

### GET /api/v1/users/top
- 认证: 否
- 响应 data: UserTop[]

### POST /api/v1/users/me
- 认证: 是
- 请求体: ModifyProfileRequest
- 响应 data: null
- 备注: birthday 解析失败会写入 null

### GET /api/v1/users/me/followers
- 认证: 是
- Query: pageNo (int, default 1, min 1), pageSize (int, default 10, max 100)
- 响应 data: `{followers: UserBrief[], total: int, hasMore: bool}`

### GET /api/v1/users/me/followees
- 认证: 是
- Query: pageNo (int, default 1, min 1), pageSize (int, default 10, max 100)
- 响应 data: `{followees: UserBrief[], total: int, hasMore: bool}`

### POST /api/v1/users/:id/follow
- 认证: 是
- Path: id (int64)
- 响应 data: null

### DELETE /api/v1/users/:id/follow
- 认证: 是
- Path: id (int64)
- 响应 data: null

### GET /api/v1/users/:id/follow
- 认证: 是
- Path: id (int64)
- 响应 data: follow_type（0/1/2/3）

### GET /api/v1/users/:id/sessions
- 认证: 是
- Path: id (int64)
- 响应 data: SessionDTO

### GET /api/v1/users/:id/sessions/messages
- 认证: 是
- Path: id (int64)
- Query: pageNo (int, default 1), pageSize (int, default 5)
- 响应 data: `{total: int, has_more: bool, messages: MessageDTO[]}`

## 帖子模块
### GET /api/v1/posts
- 认证: 否
- Query: pageNo (int, default 1), pageSize (int, default 10)
- 响应 data: `{posts: PostDetail[], total: int, hasMore: bool}`

### GET /api/v1/posts/top
- 认证: 否
- 响应 data: PostTop[]

### GET /api/v1/posts/tags
- 认证: 否
- Query: pageNo (int, default 1), pageSize (int, default 10), tag (string)
- 响应 data: `{posts: PostDetail[], total: int, hasMore: bool}`

### GET /api/v1/posts/:id
- 认证: 否
- Path: id (int64)
- 响应 data: PostDetail

### GET /api/v1/posts/:id/comments
- 认证: 否
- Path: id (int64)
- Query: pageNo (int, default 1, min 1), pageSize (int, default 10, max 100)
- 响应 data: `{comments: CommentDTO[], total: int, hasMore: bool}`

### GET /api/v1/posts/:id/comments/:cid
- 认证: 否
- Path: id (int64), cid (int64)
- Query: pageNo (int, default 1, min 1), pageSize (int, default 3, max 100)
- 响应 data: `{comments: CommentDTO[], total: int, hasMore: bool}`

### POST /api/v1/posts
- 认证: 是
- 请求体: PostCreateRequest
- 响应 data: PostDetail
- 备注: tags 在创建响应中可能为空

### POST /api/v1/posts/:id
- 认证: 是
- Path: id (int64)
- 请求体: PostUpdateRequest
- 响应 data: null

### DELETE /api/v1/posts/:id
- 认证: 是
- Path: id (int64)
- 响应 data: null

### POST /api/v1/posts/:id/comments
- 认证: 是
- Path: id (int64)
- 请求体: CommentCreateRequest
- 响应 data: CommentDTO

### DELETE /api/v1/posts/:id/comments/:cid
- 认证: 是
- Path: id (int64), cid (int64)
- 响应 data: null

### GET /api/v1/posts/:id/likes
- 认证: 是
- Path: id (int64)
- 响应 data: bool

### POST /api/v1/posts/:id/likes
- 认证: 是
- Path: id (int64)
- 响应 data: null

### DELETE /api/v1/posts/:id/likes
- 认证: 是
- Path: id (int64)
- 响应 data: null

## 会话模块
### GET /api/v1/sessions
- 认证: 是
- 响应 data: SessionDTO[]

### DELETE /api/v1/sessions/:id
- 认证: 是
- Path: id (int64)
- 响应 data: null

## WebSocket 即时聊天
### GET /api/v1/ws
- 认证: 是
- 备注: 使用 HTTP 升级为 WebSocket，Origin 白名单包含 `http://localhost:5173`

### 客户端 -> 服务端消息
- type=message:
```json
{"type":"message","session_id":"...","session_type":1,"message_from":"...","message_to":"...","content":"..."}
```
- type=read_ack:
```json
{"type":"read_ack","session_id":"..."}
```

### 服务端 -> 客户端消息
- MessageDTO JSON 结构（见数据结构）

## 抽奖模块
### GET /api/v1/gifts
- 认证: 否
- 响应 data: GiftDTO[]

### GET /api/v1/lottery/lucky
- 认证: 是
- 响应 data: GiftDTO
- 备注: 未抽中时可能返回 name 为 "谢谢参与" 或 "奖品已抽完" 的奖品

### POST /api/v1/lottery/giveup
- 认证: 是
- 请求体: LotteryGiveUpRequest
- 响应 data: null
- 备注: user_id 需与当前登录用户一致

### POST /api/v1/lottery/pay
- 认证: 是
- 请求体: LotteryPayRequest
- 响应 data: null
- 备注: user_id 需与当前登录用户一致

### GET /api/v1/lottery/result
- 认证: 是
- 响应 data: OrderDTO

## Agent 模块
### POST /api/v1/agent/chat
- 认证: 是
- 请求体: AgentChatRequest
- 响应 data: AgentDTO
