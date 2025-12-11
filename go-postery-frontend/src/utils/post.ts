import { Post } from '../types'
import { normalizeId } from './id'

// 将后端的 PostDTO / 含 Title/Id 格式转为前端使用的 Post 结构
export function normalizePost(raw: any): Post {
  const authorRaw = raw?.author || {}
  return {
    id: normalizeId(raw?.id ?? raw?.Id),
    title: raw?.title ?? raw?.Title ?? '',
    content: raw?.content ?? raw?.Content ?? '',
    author: {
      id: normalizeId(authorRaw?.id ?? authorRaw?.Id),
      name: authorRaw?.name ?? authorRaw?.Name ?? '匿名用户',
    },
    createdAt: raw?.createdAt ?? raw?.CreatedAt ?? new Date().toISOString(),
    views: raw?.views ?? raw?.Views,
    likes: raw?.likes ?? raw?.Likes,
    comments: raw?.comments ?? raw?.Comments,
  }
}
