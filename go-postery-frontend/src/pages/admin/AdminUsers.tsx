import { useCallback, useMemo, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { ExternalLink, RefreshCw, Search, Trash2, UserRound } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import type { Post, UserDetail } from '../../types'
import { apiDelete, apiGet } from '../../utils/api'
import { normalizeId } from '../../utils/id'
import { normalizePost } from '../../utils/post'
import { normalizeUserDetail } from '../../utils/user'

const formatRelativeTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return formatDistanceToNow(date, { addSuffix: true, locale: zhCN })
}

export default function AdminUsers() {
  const [userIdInput, setUserIdInput] = useState('')
  const [userId, setUserId] = useState<string>('')
  const [profile, setProfile] = useState<UserDetail | null>(null)
  const [posts, setPosts] = useState<Post[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [postsError, setPostsError] = useState<string | null>(null)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const loadUser = useCallback(async (targetUserId: string) => {
    const normalizedUserId = normalizeId(targetUserId)
    if (!normalizedUserId) return

    setIsLoading(true)
    setError(null)
    setPostsError(null)

    const profileTask = apiGet<UserDetail>(`/users/${encodeURIComponent(normalizedUserId)}`)
    const postsTask = apiGet<{
      posts: any[]
      total?: number
      hasMore?: boolean
    }>(`/users/${encodeURIComponent(normalizedUserId)}/posts?pageNo=1&pageSize=50`)

    const [profileResult, postsResult] = await Promise.allSettled([profileTask, postsTask])

    if (profileResult.status === 'fulfilled') {
      setProfile(profileResult.value.data ? normalizeUserDetail(profileResult.value.data) : null)
      setUserId(normalizedUserId)
    } else {
      const message =
        profileResult.reason instanceof Error ? profileResult.reason.message : '获取用户资料失败'
      setError(message)
      setProfile(null)
      setUserId(normalizedUserId)
    }

    if (postsResult.status === 'fulfilled') {
      const data = postsResult.value.data
      const rawList = Array.isArray(data?.posts) ? data.posts : []
      const normalized = rawList.map((item: any) => {
        const post = normalizePost(item)
        return {
          ...post,
          views: post.views ?? 0,
          likes: post.likes ?? 0,
          comments: post.comments ?? 0,
        }
      })
      setPosts(normalized)
    } else {
      const message = postsResult.reason instanceof Error ? postsResult.reason.message : '获取用户帖子失败'
      setPostsError(message)
      setPosts([])
    }

    setIsLoading(false)
  }, [])

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    const normalized = normalizeId(userIdInput)
    if (!normalized) return
    void loadUser(normalized)
  }

  const handleDeletePost = async (postId: string) => {
    const id = normalizeId(postId)
    if (!id) return
    if (!window.confirm(`确认删除帖子 #${id} 吗？此操作不可撤销。`)) return

    setDeletingId(id)
    try {
      await apiDelete(`/posts/${encodeURIComponent(id)}`)
      if (userId) {
        await loadUser(userId)
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '删除失败'
      alert(message)
    } finally {
      setDeletingId(null)
    }
  }

  const avatarUrl = useMemo(() => {
    if (profile?.avatar) return profile.avatar
    const seed = profile?.name || userId || 'user'
    return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(seed)}`
  }, [profile?.avatar, profile?.name, userId])

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-gray-900">用户管理</h2>
          <p className="text-sm text-gray-500">当前仅提供按用户 ID 查询资料与其发帖列表</p>
        </div>

        <div className="flex items-center gap-2">
          {userId && (
            <Link
              to={`/users/${encodeURIComponent(userId)}`}
              className="btn-secondary !py-2"
              title="打开用户主页"
            >
              <ExternalLink className="h-4 w-4" />
              打开主页
            </Link>
          )}
          <button
            type="button"
            onClick={() => userId && void loadUser(userId)}
            disabled={!userId || isLoading}
            className="btn-secondary !py-2"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            刷新
          </button>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="h-4 w-4 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            value={userIdInput}
            onChange={(e) => setUserIdInput(e.target.value)}
            placeholder="输入用户 ID（如 1999760900969463808）"
            className="input pl-9 h-11"
          />
        </div>
        <button type="submit" className="btn-primary h-11 px-6">
          查询用户
        </button>
      </form>

      {error && (
        <div className="p-3 rounded-xl border border-red-200 bg-red-50 text-red-700 text-sm">
          {error}
        </div>
      )}

      {!userId ? (
        <div className="p-6 rounded-2xl border border-gray-100 bg-white/60 text-gray-600 text-sm">
          请输入用户 ID 后查询。
        </div>
      ) : isLoading ? (
        <div className="p-6 rounded-2xl border border-gray-100 bg-white/60 text-gray-600 text-sm">
          加载中...
        </div>
      ) : (
        <div className="space-y-4">
          <div className="border border-gray-100 rounded-2xl bg-white/70 backdrop-blur-sm p-5">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
              <div className="flex items-center gap-3 min-w-0">
                <img src={avatarUrl} alt={profile?.name || userId} className="w-14 h-14 rounded-2xl" />
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="text-lg font-bold text-gray-900 truncate">
                      {profile?.name || '（未知用户）'}
                    </h3>
                    <span className="text-xs text-gray-500">#{userId}</span>
                  </div>
                  <div className="text-sm text-gray-600 truncate">
                    {profile?.email ? profile.email : '未设置邮箱'}
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2 text-sm text-gray-600">
                <span className="inline-flex items-center gap-2 px-3 py-2 rounded-xl bg-gray-50 border border-gray-100">
                  <UserRound className="h-4 w-4 text-primary-600" />
                  {posts.length} 篇帖子
                </span>
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <h3 className="text-base font-semibold text-gray-900">用户帖子</h3>
              {postsError && <span className="text-sm text-red-600">{postsError}</span>}
            </div>

            <div className="overflow-x-auto border border-gray-100 rounded-xl bg-white/60">
              <table className="min-w-full text-sm">
                <thead className="bg-gray-50 text-gray-600">
                  <tr>
                    <th className="text-left font-medium px-4 py-3">ID</th>
                    <th className="text-left font-medium px-4 py-3">标题</th>
                    <th className="text-left font-medium px-4 py-3">创建</th>
                    <th className="text-left font-medium px-4 py-3">数据</th>
                    <th className="text-right font-medium px-4 py-3">操作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {posts.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="px-4 py-10 text-center text-gray-500">
                        暂无帖子
                      </td>
                    </tr>
                  ) : (
                    posts.map(post => {
                      const pid = normalizeId(post.id)
                      const isDeleting = deletingId === pid
                      return (
                        <tr key={pid} className="hover:bg-gray-50/60 transition-colors">
                          <td className="px-4 py-3 text-gray-700 font-medium whitespace-nowrap">{pid}</td>
                          <td className="px-4 py-3">
                            <Link
                              to={`/post/${encodeURIComponent(pid)}`}
                              className="text-gray-900 hover:text-primary-700 font-medium line-clamp-1"
                            >
                              {post.title || '（无标题）'}
                            </Link>
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
                                onClick={() => void handleDeletePost(pid)}
                                disabled={Boolean(deletingId)}
                                className="inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold text-red-700 bg-red-50 hover:bg-red-100 border border-red-100 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
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
        </div>
      )}
    </div>
  )
}
