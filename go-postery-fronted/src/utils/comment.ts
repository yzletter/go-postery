import { Comment } from '../types'
import { normalizeId } from './id'

export function normalizeComment(raw: any): Comment {
  const authorRaw = raw?.author || {}
  const normalizedReplies = Array.isArray(raw?.replies)
    ? raw.replies.map((r: any) => normalizeComment(r))
    : raw?.Replies
      ? raw.Replies.map((r: any) => normalizeComment(r))
      : undefined
  const postIdRaw = raw?.post_id ?? raw?.postId ?? raw?.PostId
  const parentIdRaw = raw?.parent_id ?? raw?.parentId ?? raw?.ParentId
  const replyIdRaw = raw?.reply_id ?? raw?.replyId ?? raw?.ReplyId

  return {
    id: normalizeId(raw?.id ?? raw?.Id ?? ''),
    postId: postIdRaw === undefined ? undefined : normalizeId(postIdRaw),
    parentId: parentIdRaw === undefined ? undefined : normalizeId(parentIdRaw),
    replyId: replyIdRaw === undefined ? undefined : normalizeId(replyIdRaw),
    content: raw?.content ?? '',
    author: {
      id: normalizeId(authorRaw?.id ?? authorRaw?.Id ?? ''),
      name: authorRaw?.name ?? authorRaw?.Name ?? '匿名用户',
    },
    createdAt: raw?.createdAt ?? raw?.created_at ?? new Date().toISOString(),
    likes: raw?.likes ?? raw?.Likes,
    replies: normalizedReplies,
  }
}
