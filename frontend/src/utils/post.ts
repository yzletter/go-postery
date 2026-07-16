import type { Post } from '../types'
import { normalizeId } from './id'
import { toOptionalNumber } from './number'

const parseTags = (raw: unknown): string[] | undefined => {
  if (!raw) return undefined

  const list = Array.isArray(raw)
    ? raw
    : typeof raw === 'string'
      ? raw.split(/[,，\s]+/)
      : []

  const tags = list
    .map(tag => (typeof tag === 'string' ? tag.trim() : String(tag ?? '')).trim())
    .filter(Boolean)

  const unique = Array.from(new Set(tags))
  return unique.length > 0 ? unique : undefined
}

const parseContentType = (raw: unknown): 0 | 1 | undefined => {
  const normalized = toOptionalNumber(raw)
  if (normalized === 0) return 0
  if (normalized === 1) return 1
  return undefined
}

// 将后端的 PostDTO / 含 Title/Id 格式转为前端使用的 Post 结构
export function normalizePost(raw: any): Post {
  const authorRaw = raw?.author || {}
  const tags = parseTags(
    raw?.tags ??
    raw?.Tags ??
    raw?.tagList ??
    raw?.tag_list ??
    raw?.tag_names ??
    raw?.TagNames
  )

  return {
    id: normalizeId(raw?.id ?? raw?.Id),
    title: raw?.title ?? raw?.Title ?? '',
    content: raw?.content ?? raw?.Content ?? '',
    contentType: parseContentType(
      raw?.content_type ??
      raw?.ContentType ??
      raw?.contentType ??
      raw?.contentTypeId ??
      raw?.Content_Type
    ),
    author: {
      id: normalizeId(authorRaw?.id ?? authorRaw?.Id),
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
    createdAt:
      raw?.createdAt ??
      raw?.CreatedAt ??
      raw?.created_at ??
      raw?.Created_at ??
      raw?.create_time ??
      raw?.CreateTime ??
      '',
    views: toOptionalNumber(raw?.views ?? raw?.Views ?? raw?.view_count ?? raw?.ViewCount ?? raw?.viewCount),
    likes: toOptionalNumber(raw?.likes ?? raw?.Likes ?? raw?.like_count ?? raw?.LikeCount ?? raw?.likeCount),
    comments: toOptionalNumber(
      raw?.comments ?? raw?.Comments ?? raw?.comment_count ?? raw?.CommentCount ?? raw?.commentCount
    ),
    tags,
    category:
      typeof raw?.category === 'string'
        ? raw.category.trim()
        : typeof raw?.Category === 'string'
          ? raw.Category.trim()
          : undefined,
  }
}
