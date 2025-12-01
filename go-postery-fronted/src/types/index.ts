export interface ApiResponse {
  code: number // 0 表示成功，非 0 为业务错误码（如 40001、40003、50001）
  msg?: string // 提示信息
  data?: any // 响应数据
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
}

export interface Comment {
  id: string
  content: string
  author: {
    id: string
    name: string
  }
  createdAt: string
  replies?: Comment[]
}
