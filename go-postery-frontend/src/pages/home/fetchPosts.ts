import type { Post } from '../../types'
import { apiGet } from '../../utils/api'
import { normalizePost } from '../../utils/post'
import { DEFAULT_PAGE_SIZE, FETCH_TIMEOUT_MS, getRequestTag } from './constants'

export type PostListResult = {
  posts: Post[]
  total: number
  hasMore: boolean
}

export const fetchPosts = async (
  page: number,
  pageSize: number = DEFAULT_PAGE_SIZE,
  categoryKey?: string
): Promise<PostListResult> => {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS)

  try {
    const useAllEndpoint = !categoryKey || categoryKey === 'all' || categoryKey === 'follow'
    const tag = categoryKey ? getRequestTag(categoryKey) : ''
    const path = useAllEndpoint
      ? `/posts?pageNo=${page}&pageSize=${pageSize}`
      : `/posts_tag?pageNo=${page}&pageSize=${pageSize}&tag=${encodeURIComponent(tag)}`

    const { data } = await apiGet<{
      posts: any[]
      total?: number
      hasMore?: boolean
    }>(path, { signal: controller.signal })

    const rawPosts = Array.isArray(data?.posts) ? data.posts : data?.posts == null ? [] : null

    if (!data || rawPosts === null) {
      throw new Error('帖子列表响应数据格式错误')
    }

    const postsWithStats: Post[] = rawPosts.map((p: any) => {
      const normalized = normalizePost(p)
      return {
        ...normalized,
        views: normalized.views ?? 0,
        likes: normalized.likes ?? 0,
        comments: normalized.comments ?? 0,
      }
    })

    return {
      posts: postsWithStats,
      total: data.total ?? postsWithStats.length,
      hasMore: typeof data.hasMore === 'boolean' ? data.hasMore : postsWithStats.length >= pageSize,
    }
  } catch (error) {
    if ((error as any)?.name === 'AbortError') {
      console.error('Fetch posts request timeout')
      throw new Error('请求超时，请检查后端服务状态')
    }
    console.error('Failed to fetch posts:', error)
    throw error
  } finally {
    clearTimeout(timeoutId)
  }
}

