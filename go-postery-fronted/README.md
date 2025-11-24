# Go Postery - 现代化论坛前端

一个使用 React + TypeScript + Tailwind CSS 构建的现代化论坛前端应用。

## 特性

- 🎨 现代化的用户界面设计
- ⚡ 基于 Vite 的快速开发体验
- 📱 完全响应式设计，支持移动端
- 🔍 分类筛选和标签系统
- 💬 帖子详情和评论功能
- ✍️ 发帖功能

## 技术栈

- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Tailwind CSS** - 样式框架
- **React Router** - 路由管理
- **Vite** - 构建工具
- **Lucide React** - 图标库
- **date-fns** - 日期处理

## 快速开始

### 安装依赖

```bash
npm install
```

### 配置后端 API（可选）

创建 `.env` 文件配置后端 API 地址：

```env
VITE_AUTH_API_URL=http://localhost:8080
VITE_API_BASE_URL=http://localhost:8080/api
```

如果不配置，登录会默认请求 `http://localhost:8080/login`，其他接口会退回到模拟模式（仅用于开发演示）。

### 启动开发服务器

```bash
npm run dev
```

应用将在 `http://localhost:5173` 启动

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 项目结构

```
src/
├── components/      # 可复用组件
│   └── Navbar.tsx  # 导航栏组件
├── pages/          # 页面组件
│   ├── Home.tsx    # 首页（帖子列表）
│   ├── PostDetail.tsx  # 帖子详情页
│   └── CreatePost.tsx  # 发帖页面
├── types/          # TypeScript 类型定义
│   └── index.ts
├── App.tsx         # 主应用组件
├── main.tsx        # 应用入口
└── index.css       # 全局样式
```

## 功能说明

### 首页
- 显示所有帖子列表
- 支持按分类筛选
- 显示帖子预览、作者、统计信息等

### 帖子详情页
- 显示完整帖子内容
- 点赞功能
- 评论列表和发表评论

### 发帖页面
- 创建新帖子
- 选择分类和添加标签
- Markdown 格式支持

## 后端 API 集成

### 认证接口

- 应用需要以下后端 API 端点：

- `POST http://localhost:8080/login` - 用户登录（默认映射端口）
  ```json
  // 请求
  { "name": "用户名", "password": "password123" }
  
  // 响应
  { "user": { "id": "...", "name": "...", "email": "..." }, "token": "jwt_token" }
  ```

- `POST /api/auth/register` - 用户注册
  ```json
  // 请求
  { "name": "用户名", "password": "password123" }
  
  // 响应
  { "user": { "id": "...", "name": "...", "email": "..." }, "token": "jwt_token" }
  ```

- `POST /api/auth/logout` - 用户登出（可选）
  - 需要 Authorization header: `Bearer {token}`

### 模拟模式

如果后端 API 不可用，应用会自动降级到模拟模式：
- 任何非空的用户名和密码都可以"登录"
- 仅用于前端开发和演示
- **生产环境必须配置真实的后端 API**

## 开发计划

- [x] 用户认证系统
- [x] 帖子搜索功能
- [ ] 图片上传
- [ ] 富文本编辑器
- [ ] 实时通知
- [ ] 暗色模式

## 许可证

MIT

