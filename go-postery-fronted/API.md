# Go Postery 前端可见接口约定

本文档根据前端现有调用整理，便于联调与排查。

## 基础信息
- `API_BASE_URL`：帖子业务接口前缀，默认 `http://localhost:8080`（可通过 `VITE_API_BASE_URL` 配置）。
- `AUTH_API_BASE_URL`：认证接口前缀，默认同上（可通过 `VITE_AUTH_API_URL` 覆盖）。
- 所有请求在需要时携带 `credentials: include`，依赖服务端设置的 Cookie；如有 `token`（本地存储），会放在 `Authorization: Bearer <token>`。
- 请求/响应默认使用 `application/json`。
- 通用响应结构：
  ```json
  {
    "code": 0,              // 0 表示成功，其他为业务错误码
    "msg": "提示信息",
    "data": {}
  }
  ```

## 认证相关

### 登录
- `POST /login/submit`
- Body：
  ```json
  { "name": "用户名", "password": "32位MD5哈希" }
  ```
- 成功时 `data` 直接返回当前用户信息并写入认证 Cookie：
  ```json
  { "Id": 1, "Name": "用户名" }
  ```

### 注册
- `POST /register/submit`
- Body：
  ```json
  { "name": "用户名", "password": "32位MD5哈希" }
  ```
- 成功时 `data` 直接返回新用户信息并写入认证 Cookie：
  ```json
  { "Id": 1, "Name": "用户名" }
  ```

### 修改密码
- `POST /modify_pass/submit`
- Headers：`Authorization: Bearer <token>`（如本地存储存在），同时发送 Cookie。
- Body：
  ```json
  { "old_pass": "旧密码MD5", "new_pass": "新密码MD5" }
  ```
- 成功返回 `code: 0`。

### 登出
- `GET /logout`
- 仅依赖 Cookie，清理服务器端会话。

## 帖子相关

### 获取帖子列表
- `GET /posts?pageNo=<number>&pageSize=<number>`
- 成功时 `data` 示例：
  ```json
  {
    "posts": [
      {
        "Id": 1,
        "Title": "标题",
        "content": "内容",
        "author": { "id": 1, "name": "作者名" },
        "createdAt": "2024-01-01T00:00:00Z",
        "views": 120,
        "likes": 12,
        "comments": 3
      }
    ],
    "total": 100,
    "hasMore": true
  }
  ```

### 获取帖子详情
- `GET /posts/{id}`
- 成功时 `data` 为单个帖子对象：
  ```json
  {
    "Id": 1,
    "Title": "标题",
    "content": "内容",
    "author": { "id": 1, "name": "作者名" },
    "createdAt": "2024-01-01T00:00:00Z",
    "views": 120,
    "likes": 12,
    "comments": 3
  }
  ```

### 判断帖子归属
- `GET /posts/belong?id=<postId>`
- 用于检查当前登录用户是否为帖子作者，`code: 0` 表示属于当前用户。

### 创建帖子
- `POST /posts/new`
- Body：
  ```json
  { "Title": "标题", "content": "正文" }
  ```
- 成功时 `data` 返回新建的帖子对象（PostDTO）。

### 更新帖子
- `POST /posts/update`
- Body：
  ```json
  { "id": 1, "title": "新标题", "content": "新正文" }
  ```
- 成功返回 `code: 0`。

### 删除帖子
- `GET /posts/delete/{id}`
- 成功返回 `code: 0`。

## 数据模型（前端使用）
```ts
interface Post {
  id: number
  title: string
  content: string
  author: { id: number; name: string }
  createdAt: string
  views?: number
  likes?: number
  comments?: number
}

interface ApiResponse {
  code: number
  msg?: string
  data?: any
}
```
