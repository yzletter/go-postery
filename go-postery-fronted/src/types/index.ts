export interface ApiResponse<T = any> {
  code: number // 0 表示成功，非 0 为业务错误码（如 40001、40003、50001）
  msg?: string // 提示信息
  data?: T // 响应数据
}

export interface User {
  id?: number | string
  name: string
  email?: string
}

export interface Post {
  id: number
  title: string
  content: string
  author: {
    id: number
    name: string
  }
  createdAt: string
  views?: number
  likes?: number
  comments?: number
  tags?: string[]
  category?: string
}

export interface Comment {
  id: number | string
  postId?: number
  parentId?: number
  replyId?: number
  content: string
  author: {
    id: number | string
    name: string
  }
  createdAt: string
  likes?: number
  replies?: Comment[]
}
