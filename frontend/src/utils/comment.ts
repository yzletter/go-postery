import type { Comment } from '../types'
import { normalizeId } from './id'
import { toOptionalNumber } from './number'

export function normalizeComment(raw: any): Comment {
  const authorRaw = raw?.author ?? raw?.Author ?? {}
  const repliesRaw = raw?.replies ?? raw?.Replies
  const normalizedReplies = Array.isArray(repliesRaw)
    ? repliesRaw.map((reply: any) => normalizeComment(reply))
    : undefined
  const postIdRaw = raw?.post_id ?? raw?.postId ?? raw?.PostId ?? raw?.PostID
  const parentIdRaw = raw?.parent_id ?? raw?.parentId ?? raw?.ParentId ?? raw?.ParentID
  const replyIdRaw = raw?.reply_id ?? raw?.replyId ?? raw?.ReplyId ?? raw?.ReplyID
  const authorIdRaw =
    authorRaw?.id ??
    authorRaw?.Id ??
    authorRaw?.ID ??
    raw?.user_id ??
    raw?.userId ??
    raw?.UserId ??
    raw?.UserID
  const createdAtRaw =
    raw?.createdAt ??
    raw?.CreatedAt ??
    raw?.created_at ??
    raw?.Created_at ??
    raw?.created_time ??
    raw?.CreateTime

  return {
    id: normalizeId(raw?.id ?? raw?.Id ?? raw?.ID ?? ''),
    postId: postIdRaw === undefined ? undefined : normalizeId(postIdRaw),
    parentId: parentIdRaw === undefined ? undefined : normalizeId(parentIdRaw),
    replyId: replyIdRaw === undefined ? undefined : normalizeId(replyIdRaw),
    content: raw?.content ?? raw?.Content ?? '',
    author: {
      id: normalizeId(authorIdRaw ?? ''),
      name:
        authorRaw?.nickname ??
        authorRaw?.Nickname ??
        authorRaw?.name ??
        authorRaw?.Name ??
        '匿名用户',
      avatar:
        authorRaw?.avatar ??
        authorRaw?.Avatar ??
        undefined,
    },
    // The current BFF comment DTO leaves created_at empty. Preserve that
    // absence so the UI can say the timestamp was not provided instead of
    // presenting the current time as if it came from the server.
    createdAt: createdAtRaw === null || createdAtRaw === undefined ? '' : String(createdAtRaw),
    likes: toOptionalNumber(raw?.likes ?? raw?.Likes ?? raw?.like_count ?? raw?.LikeCount ?? raw?.likeCount),
    replies: normalizedReplies,
  }
}
