import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { RefreshCw, Search, Trash2 } from 'lucide-react'
import type { Post } from '../../types'
import { apiDelete } from '../../utils/api'
import { formatRelativeTime } from '../../utils/date'
import { normalizeId } from '../../utils/id'
import { fetchPosts } from '../home/fetchPosts'

export default function AdminPosts() {
  const pageSize = 20
  const [page, setPage] = useState(1)
  const [posts, setPosts] = useState<Post[]>([])
  const [total, setTotal] = useState<number>(0)
  const [hasMore, setHasMore] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const result = await fetchPosts(page, pageSize, 'all')
      setPosts(result.posts)
      setTotal(result.total)
      setHasMore(result.hasMore)
    } catch (err) {
      const message = err instanceof Error ? err.message : '加载帖子失败'
      setError(message)
      setPosts([])
      setTotal(0)
      setHasMore(false)
    } finally {
      setIsLoading(false)
    }
  }, [page])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = useMemo(() => {
    const q = keyword.trim().toLowerCase()
    if (!q) return posts
    return posts.filter(post => {
      const pool = `${post.id} ${post.title} ${post.content} ${post.author?.name ?? ''}`
      return pool.toLowerCase().includes(q)
    })
  }, [keyword, posts])

  const handleDelete = async (postId: string) => {
    const id = normalizeId(postId)
    if (!id) return

    if (!window.confirm(`确认删除帖子 #${id} 吗？此操作不可撤销。`)) {
      return
    }

    setDeletingId(id)
    try {
      await apiDelete(`/posts/${encodeURIComponent(id)}`)
      await load()
    } catch (err) {
      const message = err instanceof Error ? err.message : '删除失败'
      alert(message)
    } finally {
      setDeletingId(null)
    }
  }

  const totalPages = total > 0 ? Math.ceil(total / pageSize) : 0
  const canPrev = page > 1 && !isLoading
  const canNext = (hasMore || (totalPages > 0 && page < totalPages)) && !isLoading

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-gray-900">帖子管理</h2>
          <p className="text-sm text-gray-500">
            当前第 {page} 页{totalPages > 0 ? ` / 共 ${totalPages} 页` : ''}，共 {total} 条
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => void load()}
            disabled={isLoading}
            className="btn-secondary !py-2"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            刷新
          </button>
        </div>
      </div>

      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="h-4 w-4 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            placeholder="筛选（当前页）：ID / 标题 / 内容 / 作者"
            className="input pl-9 h-11"
          />
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setPage(prev => Math.max(1, prev - 1))}
            disabled={!canPrev}
            className="btn-secondary !py-2"
          >
            上一页
          </button>
          <button
            type="button"
            onClick={() => setPage(prev => prev + 1)}
            disabled={!canNext}
            className="btn-secondary !py-2"
          >
            下一页
          </button>
        </div>
      </div>

      {error && (
        <div className="p-3 rounded-xl border border-red-200 bg-red-50 text-red-700 text-sm">
          {error}
        </div>
      )}

      <div className="overflow-x-auto border border-gray-100 rounded-xl bg-white/60">
        <table className="min-w-full text-sm">
          <thead className="bg-gray-50 text-gray-600">
            <tr>
              <th className="text-left font-medium px-4 py-3">ID</th>
              <th className="text-left font-medium px-4 py-3">标题</th>
              <th className="text-left font-medium px-4 py-3">作者</th>
              <th className="text-left font-medium px-4 py-3">创建</th>
              <th className="text-left font-medium px-4 py-3">数据</th>
              <th className="text-right font-medium px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {isLoading ? (
              <tr>
                <td colSpan={6} className="px-4 py-10 text-center text-gray-500">
                  加载中...
                </td>
              </tr>
            ) : filtered.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-10 text-center text-gray-500">
                  暂无数据
                </td>
              </tr>
            ) : (
              filtered.map(post => {
                const id = normalizeId(post.id)
                const authorId = normalizeId(post.author?.id)
                const isDeleting = deletingId === id
                return (
                  <tr key={id} className="hover:bg-gray-50/60 transition-colors">
                    <td className="px-4 py-3 text-gray-700 font-medium whitespace-nowrap">{id}</td>
                    <td className="px-4 py-3">
                      <div className="flex flex-col">
                        <Link
                          to={`/post/${encodeURIComponent(id)}`}
                          className="text-gray-900 hover:text-primary-700 font-medium line-clamp-1"
                        >
                          {post.title || '（无标题）'}
                        </Link>
                        <span className="text-xs text-gray-500 line-clamp-1">
                          {post.content ? post.content.replace(/\s+/g, ' ').trim() : ''}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
                      {authorId ? (
                        <Link
                          to={`/users/${encodeURIComponent(authorId)}`}
                          className="hover:text-primary-700"
                        >
                          {post.author?.name || '匿名用户'}
                        </Link>
                      ) : (
                        <span>{post.author?.name || '匿名用户'}</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600 whitespace-nowrap">
                      {formatRelativeTime(post.createdAt)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 whitespace-nowrap">
                      {post.views ?? 0} 浏览 · {post.likes ?? 0} 赞 · {post.comments ?? 0} 评论
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => void handleDelete(id)}
                          disabled={Boolean(deletingId)}
                          className="inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold text-red-700 bg-red-50 hover:bg-red-100 border border-red-100 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          title="删除"
                        >
                          <Trash2 className={`h-4 w-4 ${isDeleting ? 'animate-pulse' : ''}`} />
                          <span className="hidden sm:inline">{isDeleting ? '删除中' : '删除'}</span>
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
