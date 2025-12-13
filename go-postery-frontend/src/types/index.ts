export interface ApiResponse<T = unknown> {
  code?: number // 0 表示成功，非 0 为业务错误码（如 40001、40003、50001）
  msg?: string // 提示信息
  data?: T // 响应数据
}

export type Id = string

export interface User {
  id?: Id
  name: string
  email?: string
}

export interface UserDetail {
  id: Id
  name: string
  email?: string
  avatar?: string
  bio?: string
  gender?: number
  birthday?: string
  location?: string
  country?: string
  lastLoginIP?: string
}

export interface ModifyUserProfileRequest {
  email?: string
  avatar?: string
  bio?: string
  gender?: number
  birthday?: string
  location?: string
  country?: string
}

export interface Post {
  id: Id
  title: string
  content: string
  author: {
    id: Id
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
  id: Id
  postId?: Id
  parentId?: Id
  replyId?: Id
  content: string
  author: {
    id: Id
    name: string
  }
  createdAt: string
  likes?: number
  replies?: Comment[]
}

export interface FollowUser {
  id: Id
  name: string
  avatar?: string
}

export type FollowRelation = 0 | 1 | 2 | 3
