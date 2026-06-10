import { apiGet } from '../../utils/api'
import { FETCH_TIMEOUT_MS } from './constants'

export type TopPost = {
  id: string
  title: string
  score: number
}

export const fetchTopPosts = async (): Promise<TopPost[]> => {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS)

  try {
    const { data } = await apiGet<TopPost[]>('/posts/top', { signal: controller.signal })
    if (!Array.isArray(data)) {
      throw new Error('热门帖子响应数据格式错误')
    }

    return data
      .map((item) => ({
        id: String(item.id ?? ''),
        title: String(item.title ?? ''),
        score: Number.isFinite(Number(item.score)) ? Number(item.score) : 0,
      }))
      .filter((item) => item.id && item.title)
  } catch (error) {
    if ((error as { name?: string })?.name === 'AbortError') {
      console.error('Fetch top posts request timeout')
      throw new Error('请求超时，请检查后端服务状态')
    }
    console.error('Failed to fetch top posts:', error)
    throw error
  } finally {
    clearTimeout(timeoutId)
  }
}
