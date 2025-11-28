# Go Postery 后端接口约定

本文件总结了前端当前代码中的所有后端调用需求，便于后端按约定实现或联调。

## 通用
- 响应统一：`ApiResponse { code: number; msg?: string; data?: any }`，`code === 0` 视为成功。
- 默认基址：帖子接口多处写死 `http://localhost:8080/...`，可用环境变量覆盖：
  - `VITE_API_BASE_URL`（帖子相关）
  - `VITE_AUTH_API_URL`（认证相关）
- 所有请求都带 `credentials: 'include'`，后端需允许跨域并下发/校验 Cookie（建议 JWT 放 HttpOnly Cookie）。
- 登录/注册/修改密码前端会先对密码做 MD5，再将哈希值作为参数提交。
- 前端要求 `Content-Type: application/json`，否则会报“响应不是JSON格式”。

## 认证接口
- `POST /login/submit`
  - Body: `{ name, password }`（password 已 MD5）
  - 返回: `data.user`；会话凭证应通过 Cookie 下发
- `POST /register/submit`
  - Body: `{ name, password }`（已 MD5）
  - 返回: `data.user`
- `POST /modify_pass/submit`
  - Body: `{ old_pass, new_pass }`（均 MD5）
  - 返回: `code = 0` 即视为成功
- `GET /logout`
  - 凭 Cookie 退出登录；`code = 0` 即视为成功

## 帖子接口
- `GET /posts?pageNo={page}&pageSize={pageSize}`
  - 返回: `data.posts: Post[]`
- `POST /posts/new`
  - Body: `{ title, content }`
  - 返回: `data` 内含新帖 id 即可
- `GET /posts/{id}`
  - 返回: `data: Post`
- `POST /posts/update`
  - Body: `{ id: number, title, content }`
  - 返回: `code = 0` 视为成功
- `GET /posts/delete/{id}`
  - 删除帖子；`code = 0` 视为成功
- `GET /posts/belong?id={postId}`
  - 返回: `data` 为 `"true"`/`"false"`（或布尔值），表示当前用户是否为作者

## 数据模型（前端期待）
- `Post`: `{ id, title, content, author: { id, name }, createdAt }`
- `User`: `{ id, name, email? }`

## 需要后端注意的点
- 所有接口需支持 Cookie 方式的身份凭证，并配置跨域允许携带 Cookie。
- 帖子接口目前路径有的包含 `/api` 有的没有，如需统一可在前端调整，但当前实现按上述路径调用。
