import { useCallback, useEffect, useState } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  PenSquare,
  Settings,
  HeartHandshake,
  Send,
} from 'lucide-react'
import UserAvatar from '../components/UserAvatar'
import { useAuth } from '../contexts/AuthContext'
import { apiGet } from '../utils/api'
import { formatRelativeTime } from '../utils/date'
import { normalizeUserDetail } from '../utils/user'
import { normalizePost } from '../utils/post'
import type { Post, UserDetail } from '../types'
import type { FollowRelation } from '../types'
import { followUser, getFollowRelation, isFollowing, unfollowUser } from '../utils/follow'

export default function Profile() {
  const navigate = useNavigate()
  const location = useLocation()
  const { userId } = useParams<{ userId?: string }>()
  const { user } = useAuth()
  const locationState = (location.state as { username?: string } | null) || {}
  const [profileInfo, setProfileInfo] = useState<UserDetail | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [recentPosts, setRecentPosts] = useState<Post[]>([])
  const [recentTotal, setRecentTotal] = useState<number | null>(null)
  const [recentHasMore, setRecentHasMore] = useState(false)
  const [isRecentLoading, setIsRecentLoading] = useState(true)
  const [recentError, setRecentError] = useState<string | null>(null)
  const [followRelation, setFollowRelation] = useState<FollowRelation | null>(null)
  const [isFollowLoading, setIsFollowLoading] = useState(false)
  const [followRelationError, setFollowRelationError] = useState(false)

  const resolvedUserId = userId ?? (user?.id != null ? String(user.id) : '')
  const isCurrentUser = userId
    ? Boolean(user && resolvedUserId && String(user.id) === resolvedUserId)
    : Boolean(user)
  const displayName =
    profileInfo?.name ||
    locationState.username ||
    (isCurrentUser ? user?.name : undefined) ||
    (resolvedUserId ? `用户 ${resolvedUserId}` : '个人主页')
  const displayUserId = profileInfo?.id || resolvedUserId
  const subtitle = isCurrentUser
    ? '分享你的想法，构建更好的社区'
    : `正在查看 ${displayName} 的主页`
  const formatDate = (value?: string | Date) => {
    if (!value) return '未设置'
    const date = value instanceof Date ? value : new Date(value)
    return Number.isNaN(date.getTime()) ? '未设置' : date.toLocaleDateString('zh-CN')
  }
  const genderLabel = (() => {
    switch (profileInfo?.gender) {
      case 1:
        return '男'
      case 2:
        return '女'
      case 3:
        return '其他'
      default:
        return '保密'
    }
  })()
  const summaryStats = [
    {
      key: 'posts',
      label: '帖子',
      value: recentTotal == null ? '—' : recentTotal.toLocaleString('zh-CN'),
      icon: <PenSquare className="h-4 w-4 text-primary-600" />,
    },
  ]
  const detailStatus = isLoading ? '资料加载中...' : error ? '加载失败' : '最新资料'
  const actionButtons = (
    <div className="flex items-center space-x-3">
      <Link to="/create" className="btn-primary flex items-center space-x-2 shadow-sm hover:-translate-y-0.5 transition-transform">
        <PenSquare className="h-4 w-4" />
        <span>发帖</span>
      </Link>
      <Link to="/settings" className="btn-secondary flex items-center space-x-2 shadow-sm hover:-translate-y-0.5 transition-transform">
        <Settings className="h-4 w-4" />
        <span>设置</span>
      </Link>
    </div>
  )

  const loadFollowRelation = useCallback(async () => {
    if (!user || isCurrentUser || !resolvedUserId) {
      setFollowRelation(null)
      setFollowRelationError(false)
      return
    }

    setIsFollowLoading(true)
    setFollowRelationError(false)
    try {
      const relation = await getFollowRelation(resolvedUserId)
      setFollowRelation(relation)
    } catch (error) {
      console.warn('Failed to fetch follow relation:', error)
      setFollowRelation(null)
      setFollowRelationError(true)
    } finally {
      setIsFollowLoading(false)
    }
  }, [isCurrentUser, resolvedUserId, user])

  useEffect(() => {
    void loadFollowRelation()
  }, [loadFollowRelation])

  const handleToggleFollow = async () => {
    if (!user || isCurrentUser || !resolvedUserId || isFollowLoading || followRelation === null) return

    const prev = followRelation
    const relation = followRelation

    setIsFollowLoading(true)
    setFollowRelationError(false)
    try {
      if (isFollowing(relation)) {
        await unfollowUser(resolvedUserId)
        setFollowRelation(relation === 3 ? 2 : 0)
      } else {
        await followUser(resolvedUserId)
        setFollowRelation(relation === 2 ? 3 : 1)
      }
    } catch (error) {
      console.error('更新关注关系失败:', error)
      setFollowRelation(prev)
      setFollowRelationError(true)
      alert(error instanceof Error ? error.message : '更新关注关系失败')
    } finally {
      setIsFollowLoading(false)
    }
  }

  useEffect(() => {
    if (!resolvedUserId) {
      setIsLoading(false)
      setError(null)
      setProfileInfo(null)
      return
    }

    let isMounted = true

    const fetchProfile = async () => {
      setIsLoading(true)
      setError(null)
      setProfileInfo(null)

      try {
        const { data } = await apiGet<UserDetail>(`/users/${resolvedUserId}`)
        if (!isMounted) return
        setProfileInfo(data ? normalizeUserDetail(data) : null)
      } catch (err) {
        if (!isMounted) return
        console.error('Failed to fetch profile:', err)
        const message = err instanceof Error ? err.message : '获取个人资料失败'
        setError(message)
        setProfileInfo(null)
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    void fetchProfile()

    return () => {
      isMounted = false
    }
  }, [resolvedUserId])

  useEffect(() => {
    if (!resolvedUserId) {
      setRecentPosts([])
      setRecentTotal(null)
      setRecentHasMore(false)
      setIsRecentLoading(false)
      setRecentError(null)
      return
    }

    let isMounted = true
    const controller = new AbortController()

    const fetchRecentPosts = async () => {
      setIsRecentLoading(true)
      setRecentError(null)
      setRecentPosts([])
      setRecentTotal(null)
      setRecentHasMore(false)

      try {
        const { data } = await apiGet<{
          posts: any[]
          total?: number
          has_more?: boolean
          hasMore?: boolean
        }>(`/users/${resolvedUserId}/posts?pageNo=1&pageSize=20`, { signal: controller.signal })
        if (!isMounted) return

        const rawList = Array.isArray(data?.posts) ? data.posts : []

        const normalized = rawList.map((item: any) => {
          const post = normalizePost(item)
          const createdAt =
            item?.created_at ??
            item?.createdAt ??
            item?.CreatedAt ??
            ''
          return {
            ...post,
            createdAt: typeof createdAt === 'string' ? createdAt.trim() : '',
          }
        })
        const total = Number(data?.total)

        setRecentPosts(normalized)
        setRecentTotal(Number.isFinite(total) && total >= 0 ? total : null)
        setRecentHasMore(Boolean(data?.has_more ?? data?.hasMore))
      } catch (err) {
        if (!isMounted) return
        console.error('Failed to fetch recent posts:', err)
        const message = err instanceof Error ? err.message : '获取最近动态失败'
        setRecentError(message)
        setRecentPosts([])
        setRecentTotal(null)
        setRecentHasMore(false)
      } finally {
        if (isMounted) {
          setIsRecentLoading(false)
        }
      }
    }

    void fetchRecentPosts()

    return () => {
      isMounted = false
      controller.abort()
    }
  }, [resolvedUserId])

  if (!userId && !user) {
    navigate('/login')
    return null
  }

  return (
    <div className="max-w-6xl mx-auto space-y-5">
      <Link
        to="/"
        className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
      >
        <ArrowLeft className="h-5 w-5" />
        <span>返回首页</span>
      </Link>

      <div className="card relative overflow-hidden bg-gradient-to-r from-primary-50 via-white to-blue-50 border-primary-100">
        <div className="absolute inset-0 pointer-events-none bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.1),transparent_45%),radial-gradient(circle_at_bottom_right,rgba(14,165,233,0.12),transparent_40%)]" />
        <div className="relative space-y-5">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div className="flex items-center space-x-4">
              <UserAvatar
                avatar={profileInfo?.avatar}
                name={displayName}
                userId={displayUserId}
                className="w-20 h-20 rounded-full border-4 border-white shadow-sm ring-2 ring-primary-100"
              />
              <div>
                <h1 className="text-2xl md:text-3xl font-bold text-gray-900">{displayName}</h1>
                <p className="text-gray-600 text-sm">{subtitle}</p>
                <div className="flex items-center gap-2 flex-wrap text-xs text-gray-600 mt-2">
                  {displayUserId && (
                    <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white/80 border border-gray-200">
                      ID: {displayUserId}
                    </span>
                  )}
                  <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white/80 border border-gray-200">
                    {profileInfo?.location || '未设置'}
                  </span>
                  <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white/80 border border-gray-200">
                    {profileInfo?.country || '未设置'}
                  </span>
                </div>
              </div>
            </div>
            {isCurrentUser ? (
              actionButtons
            ) : (
              <div className="flex items-center space-x-3">
                {user ? (
                  <button
                    type="button"
                    onClick={() => {
                      if (followRelationError) {
                        void loadFollowRelation()
                        return
                      }
                      void handleToggleFollow()
                    }}
                    disabled={isFollowLoading || (!followRelationError && followRelation === null)}
                    className="btn-secondary flex items-center space-x-2 shadow-sm hover:-translate-y-0.5 transition-transform disabled:opacity-60 disabled:cursor-not-allowed"
                  >
                    <HeartHandshake className="h-4 w-4" />
                    <span>
                      {isFollowLoading
                        ? followRelation === null
                          ? '获取关注状态...'
                          : '处理中...'
                        : followRelationError
                          ? '重试关注状态'
                        : isFollowing(followRelation ?? 0)
                          ? '取消关注'
                          : followRelation === 2
                            ? '回关'
                            : '关注'}
                    </span>
                  </button>
                ) : (
                  <Link
                    to="/login"
                    className="btn-secondary flex items-center space-x-2 shadow-sm hover:-translate-y-0.5 transition-transform"
                  >
                    <HeartHandshake className="h-4 w-4" />
                    <span>登录后关注</span>
                  </Link>
                )}
                <Link
                  to="/messages"
                  state={{ username: displayName, userId: displayUserId }}
                  className="btn-primary flex items-center space-x-2 shadow-sm hover:-translate-y-0.5 transition-transform"
                >
                  <Send className="h-4 w-4" />
                  <span>私信</span>
                </Link>
              </div>
            )}
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            {summaryStats.map((stat) => (
              <div
                key={stat.key}
                className="rounded-xl bg-white/80 border border-white/60 px-3 py-3 shadow-sm flex items-center gap-3 hover:-translate-y-0.5 transition-transform"
              >
                <div className="w-9 h-9 rounded-lg bg-primary-50 flex items-center justify-center">
                  {stat.icon}
                </div>
                <div>
                  <p className="text-xs text-gray-500">{stat.label}</p>
                  <p className="text-lg font-semibold text-gray-900">{stat.value}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="space-y-6">
        {/* 详细信息 */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-primary-500" />
              <h2 className="text-lg font-semibold text-gray-900">详细资料</h2>
            </div>
            <span className="text-xs text-gray-500">{detailStatus}</span>
          </div>
          {error && (
            <div className="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          )}
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <InfoItem label="性别" value={genderLabel} />
            <InfoItem label="生日" value={formatDate(profileInfo?.birthday)} />
            <InfoItem label="国家" value={profileInfo?.country || '—'} />
            <InfoItem label="所在地" value={profileInfo?.location || '—'} />
            <InfoItem label="最近登录 IP" value={profileInfo?.lastLoginIP || '—'} />
          </div>
          <div className="mt-4">
            <div className="border border-primary-50 bg-primary-50/60 rounded-lg p-4">
              <p className="text-xs text-primary-700 font-semibold">个人简介</p>
              <p className="text-sm font-medium text-gray-900 mt-1 break-words">
                {profileInfo?.bio || '这个人很神秘，还没有简介'}
              </p>
            </div>
          </div>
        </div>

        {/* 最近动态 */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">最近动态</h2>
            <Link to="/create" className="text-sm text-primary-600 hover:text-primary-700 font-medium">
              去创作
            </Link>
          </div>
          <div className="divide-y divide-gray-100">
            {isRecentLoading && (
              <div className="py-4 text-sm text-gray-500 px-2">加载最近动态...</div>
            )}
            {recentError && !isRecentLoading && (
              <div className="py-3 text-sm text-red-600 px-2">{recentError}</div>
            )}
            {!isRecentLoading && !recentError && recentPosts.length === 0 && (
              <div className="py-4 text-sm text-gray-500 px-2">暂无最近动态</div>
            )}
            {recentPosts.map((post, index) => (
              <Link
                to={`/post/${post.id}`}
                key={post.id}
                className="flex items-center justify-between py-3 hover:bg-gray-50 px-2 rounded-lg transition-all hover:-translate-y-0.5"
              >
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 rounded-lg bg-primary-50 text-primary-700 font-semibold flex items-center justify-center">
                    {index + 1}
                  </div>
                  <div>
                    <p className="text-gray-900 font-medium line-clamp-1">{post.title}</p>
                    <p className="text-xs text-gray-500">
                      发布时间：{formatRelativeTime(post.createdAt, '时间未提供')}
                    </p>
                  </div>
                </div>
                <div className="text-sm text-primary-600">查看详情</div>
              </Link>
            ))}
            {!isRecentLoading && !recentError && recentPosts.length > 0 && recentTotal != null && (
              <div className="py-3 px-2 text-xs text-gray-500">
                已展示 {recentPosts.length} / {recentTotal} 条动态
                {recentHasMore ? '，还有更多内容' : ''}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function InfoItem({
  label,
  value,
  full,
}: {
  label: string
  value: string | number
  full?: boolean
}) {
  return (
    <div className={`border border-gray-100 rounded-lg p-3 bg-gray-50 hover:-translate-y-0.5 transition-transform ${full ? 'w-full' : ''}`}>
      <p className="text-xs text-gray-500">{label}</p>
      <p className="text-sm font-medium text-gray-900 mt-1 break-words">{value}</p>
    </div>
  )
}
