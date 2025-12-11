import { useEffect, useState, type ReactNode } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  Heart,
  MessageSquare,
  PenSquare,
  Settings,
  Share2,
  Users,
  HeartHandshake,
  Send,
} from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { apiGet } from '../utils/api'
import { normalizeUserDetail } from '../utils/user'
import type { UserDetail } from '../types'

const mockRecentPosts = [
  { id: 1, title: '如何快速搭建一套前后端同构的论坛？', time: '2 小时前', views: 320 },
  { id: 2, title: 'Go 服务的日志与链路追踪最佳实践', time: '1 天前', views: 210 },
  { id: 3, title: 'Tailwind 设计系统落地经验分享', time: '3 天前', views: 180 },
]

export default function Profile() {
  const navigate = useNavigate()
  const location = useLocation()
  const { userId } = useParams<{ userId?: string }>()
  const { user } = useAuth()
  const locationState = (location.state as { username?: string } | null) || {}
  const [profileInfo, setProfileInfo] = useState<UserDetail | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const resolvedUserId = userId ?? (user?.id != null ? String(user.id) : '')
  const isCurrentUser = userId
    ? Boolean(user && resolvedUserId && String(user.id) === resolvedUserId)
    : Boolean(user)
  const displayName =
    profileInfo?.name ||
    locationState.username ||
    (isCurrentUser ? user?.name : undefined) ||
    user?.name ||
    (resolvedUserId ? `用户 ${resolvedUserId}` : '个人主页')
  const displayUserId = profileInfo?.id || resolvedUserId
  const subtitle = isCurrentUser
    ? '分享你的想法，构建更好的社区'
    : `正在查看 ${displayName} 的主页`
  const avatarUrl = profileInfo?.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${displayName}`
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
    { key: 'posts', label: '帖子', value: '12', icon: <PenSquare className="h-4 w-4 text-primary-600" /> },
    { key: 'followers', label: '关注者', value: '89', icon: <Users className="h-4 w-4 text-primary-600" /> },
    { key: 'likes', label: '获赞', value: '326', icon: <Heart className="h-4 w-4 text-primary-600" /> },
    { key: 'share', label: '分享', value: '34', icon: <Share2 className="h-4 w-4 text-primary-600" /> },
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

      try {
        const { data } = await apiGet<UserDetail>(`/profile/${resolvedUserId}`)
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
              <img
                src={avatarUrl}
                alt={displayName}
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
                <Link
                  to="/follows"
                  className="btn-secondary flex items-center space-x-2 shadow-sm hover:-translate-y-0.5 transition-transform"
                >
                  <HeartHandshake className="h-4 w-4" />
                  <span>关注</span>
                </Link>
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
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
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

      <div className="grid lg:grid-cols-[1.15fr_0.85fr] gap-6">
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
              <InfoItem label="邮箱" value={profileInfo?.email || '—'} />
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
              {mockRecentPosts.map((post) => (
                <Link
                  to={`/post/${post.id}`}
                  key={post.id}
                  className="flex items-center justify-between py-3 hover:bg-gray-50 px-2 rounded-lg transition-all hover:-translate-y-0.5"
                >
                  <div className="flex items-center space-x-3">
                    <div className="w-10 h-10 rounded-lg bg-primary-50 text-primary-700 font-semibold flex items-center justify-center">
                      {post.id}
                    </div>
                    <div>
                      <p className="text-gray-900 font-medium line-clamp-1">{post.title}</p>
                      <p className="text-xs text-gray-500">更新于 {post.time}</p>
                    </div>
                  </div>
                  <div className="text-sm text-gray-500">浏览 {post.views}</div>
                </Link>
              ))}
            </div>
          </div>
        </div>

        <div className="space-y-6">
          {/* 统计卡片 */}
          <div className="card space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900">创作概览</h2>
              <span className="text-xs text-gray-500">本周</span>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-2 lg:grid-cols-2 gap-3">
              <StatItem icon={<PenSquare className="h-5 w-5 text-primary-600" />} label="帖子" value="12" />
              <StatItem icon={<MessageSquare className="h-5 w-5 text-primary-600" />} label="回复" value="48" />
              <StatItem icon={<Heart className="h-5 w-5 text-primary-600" />} label="获赞" value="326" />
              <StatItem icon={<Share2 className="h-5 w-5 text-primary-600" />} label="分享" value="34" />
            </div>
          </div>

          <div className="card space-y-3 text-sm text-gray-700">
            <div className="flex items-center gap-2">
              <HeartHandshake className="h-4 w-4 text-primary-600" />
              <p className="font-semibold text-gray-900">互动空间</p>
            </div>
            <p className="text-gray-600">
              维护好个人资料、主动发布或私信交流，可以让更多同好看到你的分享。
            </p>
            <div className="flex flex-wrap gap-2">
              <Link to="/create" className="btn-primary text-sm px-3 py-1.5">去发帖</Link>
              <Link to="/messages" className="btn-secondary text-sm px-3 py-1.5">打开私信</Link>
              <Link to="/settings" className="btn-secondary text-sm px-3 py-1.5">完善资料</Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function StatItem({
  icon,
  label,
  value,
}: {
  icon: ReactNode
  label: string
  value: string
}) {
  return (
    <div className="rounded-xl border border-gray-100 bg-white px-4 py-3 shadow-sm hover:-translate-y-0.5 transition-transform">
      <div className="flex items-center space-x-2 mb-2">
        <div className="w-8 h-8 rounded-lg bg-primary-50 text-primary-700 flex items-center justify-center">
          {icon}
        </div>
        <span className="text-xs text-gray-500">{label}</span>
      </div>
      <div className="text-xl font-bold text-gray-900">{value}</div>
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
