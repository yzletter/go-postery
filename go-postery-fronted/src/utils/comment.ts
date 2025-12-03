import { Comment } from '../types'

export function normalizeComment(raw: any): Comment {
  const authorRaw = raw?.author || {}
  const normalizedReplies = Array.isArray(raw?.replies)
    ? raw.replies.map((r: any) => normalizeComment(r))
    : raw?.Replies
      ? raw.Replies.map((r: any) => normalizeComment(r))
      : undefined

  return {
    id: raw?.id ?? raw?.Id ?? '',
    postId: raw?.post_id ?? raw?.postId ?? raw?.PostId,
    parentId: raw?.parent_id ?? raw?.parentId ?? raw?.ParentId,
    content: raw?.content ?? '',
    author: {
      id: authorRaw?.id ?? authorRaw?.Id ?? '',
      name: authorRaw?.name ?? authorRaw?.Name ?? '匿名用户',
    },
    createdAt: raw?.createdAt ?? raw?.created_at ?? new Date().toISOString(),
    likes: raw?.likes ?? raw?.Likes,
    replies: normalizedReplies,
  }
}
