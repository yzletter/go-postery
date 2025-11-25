# Go Postery 后端 API 文档

## 概述

本文档描述了 Go Postery 论坛前端所需的后端 API 接口规范。基于前端代码分析，后端需要提供用户认证、帖子管理等功能。

## 接口规范

### 1. 认证相关接口

#### 1.1 用户登录
**接口地址**: `POST /login` 或 `POST /api/auth/login`

**请求参数**:
```json
{
  "name": "string, required, 用户名",
  "password": "string, required, 密码（MD5哈希）"
}
```

**成功响应** (200 OK):
```json
{
  "user": {
    "id": "string, 用户唯一标识",
    "name": "string, 用户名",
    "email": "string, optional, 用户邮箱（可选，前端已删除显示）",
    "avatar": "string, optional, 头像URL"
  },
  "token": "string, JWT访问令牌"
}
```

**错误响应** (400/401):
```json
{
  "message": "string, 错误信息"
}
```

#### 1.2 用户注册
**接口地址**: `POST /api/auth/register`

**请求参数**:
```json
{
  "name": "string, required, 用户名",
  "password": "string, required, 密码（MD5哈希）"
}
```

**成功响应** (201 Created):
```json
{
  "user": {
    "id": "string, 用户唯一标识",
    "name": "string, 用户名",
    "email": "string, optional, 用户邮箱",
    "avatar": "string, optional, 头像URL"
  },
  "token": "string, JWT访问令牌"
}
```

#### 1.3 修改密码
**接口地址**: `POST /api/auth/change-password`

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "oldPassword": "string, required, 旧密码（MD5哈希）",
  "newPassword": "string, required, 新密码（MD5哈希）"
}
```

**成功响应** (200 OK):
```json
{
  "message": "密码修改成功"
}
```

### 2. 帖子相关接口

#### 2.1 获取帖子列表
**接口地址**: `GET /api/posts?page={页码}&pageSize={每页数量}`

**查询参数**:
- `page`: integer, optional, 页码，默认1
- `pageSize`: integer, optional, 每页数量，默认10

**成功响应** (200 OK):
```json
{
  "posts": [
    {
      "id": "string, 帖子唯一标识",
      "title": "string, 帖子标题",
      "content": "string, 帖子内容",
      "author": {
        "id": "string, 作者ID",
        "name": "string, 作者名称",
        "avatar": "string, optional, 作者头像URL"
      },
      "createdAt": "string, ISO 8601格式时间",
      "updatedAt": "string, optional, 更新时间"
    }
  ],
  "total": "integer, 总帖子数",
  "hasMore": "boolean, 是否还有更多数据"
}
```

#### 2.2 获取单个帖子详情
**接口地址**: `GET /api/posts/{postId}`

**路径参数**:
- `postId`: string, required, 帖子ID

**成功响应** (200 OK):
```json
{
  "id": "string, 帖子唯一标识",
  "title": "string, 帖子标题",
  "content": "string, 帖子内容",
  "author": {
    "id": "string, 作者ID",
    "name": "string, 作者名称",
    "avatar": "string, optional, 作者头像URL"
  },
  "createdAt": "string, ISO 8601格式时间",
  "updatedAt": "string, optional, 更新时间"
}
```

#### 2.3 创建新帖子
**接口地址**: `POST /api/posts`

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "title": "string, required, 帖子标题",
  "content": "string, required, 帖子内容"
}
```

**成功响应** (201 Created):
```json
{
  "id": "string, 新创建帖子的ID",
  "title": "string, 帖子标题",
  "content": "string, 帖子内容",
  "author": {
    "id": "string, 作者ID",
    "name": "string, 作者名称",
    "avatar": "string, optional, 作者头像URL"
  },
  "createdAt": "string, ISO 8601格式时间"
}
```

### 3. 数据类型定义

#### User 类型
```typescript
interface User {
  id: string        // 用户唯一标识
  name: string      // 用户名
  email?: string    // 用户邮箱（可选，前端已删除显示）
  avatar?: string   // 头像URL（可选）
}
```

#### Post 类型
```typescript
interface Post {
  id: string           // 帖子唯一标识
  title: string        // 帖子标题
  content: string      // 帖子内容
  author: {
    id: string         // 作者ID
    name: string       // 作者名称
    avatar?: string    // 作者头像URL（可选）
  }
  createdAt: string    // 创建时间（ISO 8601格式）
  updatedAt?: string   // 更新时间（可选）
}
```

## 实现要求

### 1. 认证机制
- 使用 JWT (JSON Web Token) 进行用户认证
- Token 有效期建议设置为 24 小时
- 在请求头中通过 `Authorization: Bearer {token}` 传递

### 2. 密码处理
- 前端传输的密码已经是 MD5 哈希值
- 后端需要能够处理 MD5 格式的密码
- 建议在后端再进行一次加密存储

### 3. 头像处理
- 如果用户没有上传头像，前端会使用 DiceBear API 生成默认头像
- 默认头像格式：`https://api.dicebear.com/7.x/avataaars/svg?seed={用户ID或用户名}`
- 后端可以存储用户自定义头像 URL

### 4. 时间格式
- 所有时间字段必须使用 ISO 8601 格式字符串
- 示例：`2024-01-15T10:30:00.000Z`

### 5. 分页处理
- 帖子列表接口支持分页查询
- 默认每页 10 条记录
- 需要返回总记录数和是否还有更多数据

### 6. 错误处理
- 所有错误响应应包含明确的错误信息
- HTTP 状态码应符合 RESTful 规范
- 错误信息格式：`{ "message": "错误描述" }`

## 环境配置

前端通过环境变量配置 API 地址：
- `VITE_API_BASE_URL`: 主要 API 地址（默认：http://localhost:8080/api）
- `VITE_AUTH_API_URL`: 认证 API 地址（默认：http://localhost:8080）

## 重要变更说明

1. **邮箱显示已删除**：根据最近的修改，邮箱显示已从 Profile 页面和 Navbar 用户菜单中删除，因此 `email` 字段变为可选。

2. **搜索功能已删除**：前端搜索功能已被移除，后端不需要提供搜索相关的接口。

3. **评论功能已删除**：评论功能已从详情页面移除，后端暂时不需要提供评论相关的接口。

## 示例代码

### 后端 Go 语言示例结构

```go
// 用户模型
type User struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email,omitempty"` // omitempty 使其可选
    Avatar   string `json:"avatar,omitempty"` // omitempty 使其可选
    Password string `json:"-"` // 不返回给前端
}

// 帖子模型
type Post struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Author    Author    `json:"author"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt,omitempty"` // omitempty 使其可选
}

type Author struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Avatar string `json:"avatar,omitempty"` // omitempty 使其可选
}
```

这个文档详细描述了后端需要实现的所有接口，包括请求参数、响应格式和实现要求。你可以根据这个文档来开发后端 API。