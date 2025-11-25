export interface ApiResponse {
  code: number // 0表示成功，1表示失败
  msg?: string // 错误信息
  data?: any // 响应数据
}

export interface User {
  id: string
  name: string
  email?: string // 邮箱变为可选
  avatar?: string
}

export interface Post {
  id: string
  title: string
  content: string
  author: {
    id: string
    name: string
    avatar?: string
  }
  createdAt: string
  updatedAt?: string


}

export interface Comment {
  id: string
  content: string
  author: {
    id: string
    name: string
    avatar?: string
  }
  createdAt: string
  replies?: Comment[]
}

