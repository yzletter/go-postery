import type { Comment } from '../types'
import { normalizeId } from './id'
import { toOptionalNumber } from './number'

export function normalizeComment(raw: any): Comment {
  const authorRaw = raw?.author ?? raw?.Author ?? {}
  const repliesRaw = raw?.replies ?? raw?.Replies
  const normalizedReplies = Array.isArray(repliesRaw)
    ? repliesRaw.map((reply: any) => normalizeComment(reply))
    : undefined
  const postIdRaw = raw?.post_id ?? raw?.postId ?? raw?.PostId
  const parentIdRaw = raw?.parent_id ?? raw?.parentId ?? raw?.ParentId
  const replyIdRaw = raw?.reply_id ?? raw?.replyId ?? raw?.ReplyId

  return {
    id: normalizeId(raw?.id ?? raw?.Id ?? ''),
    postId: postIdRaw === undefined ? undefined : normalizeId(postIdRaw),
    parentId: parentIdRaw === undefined ? undefined : normalizeId(parentIdRaw),
    replyId: replyIdRaw === undefined ? undefined : normalizeId(replyIdRaw),
    content: raw?.content ?? raw?.Content ?? '',
    author: {
      id: normalizeId(authorRaw?.id ?? authorRaw?.Id ?? ''),
      name: authorRaw?.name ?? authorRaw?.Name ?? '匿名用户',
    },
    createdAt:
      raw?.createdAt ??
      raw?.CreatedAt ??
      raw?.created_at ??
      raw?.Created_at ??
      raw?.created_time ??
      raw?.CreateTime ??
      new Date().toISOString(),
    likes: toOptionalNumber(raw?.likes ?? raw?.Likes ?? raw?.like_count ?? raw?.LikeCount ?? raw?.likeCount),
    replies: normalizedReplies,
  }
}
