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
  views: number
  likes: number
  comments: number
  tags?: string[]
  category?: string
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
  likes: number
  replies?: Comment[]
}

