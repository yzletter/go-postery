import { Post } from '../types'
import { normalizeId } from './id'

const normalizeCount = (value: unknown): number | undefined => {
  if (value === null || value === undefined) return undefined
  const num = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(num) ? num : undefined
}

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
    views: normalizeCount(raw?.views ?? raw?.Views ?? raw?.view_count ?? raw?.ViewCount ?? raw?.viewCount),
    likes: normalizeCount(raw?.likes ?? raw?.Likes ?? raw?.like_count ?? raw?.LikeCount ?? raw?.likeCount),
    comments: normalizeCount(raw?.comments ?? raw?.Comments ?? raw?.comment_count ?? raw?.CommentCount ?? raw?.commentCount),
  }
}
