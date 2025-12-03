import type { ReactNode } from 'react'
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
  
  if (!userId && !user) {
    navigate('/login')
    return null
  }

  const resolvedUserId = userId ?? (user?.id != null ? String(user.id) : '')
  const isCurrentUser = userId
    ? Boolean(user && resolvedUserId && String(user.id) === resolvedUserId)
    : Boolean(user)
  const displayName =
    locationState.username ||
    (isCurrentUser ? user?.name : undefined) ||
    user?.name ||
    (resolvedUserId ? `用户 ${resolvedUserId}` : '个人主页')
  const subtitle = isCurrentUser
    ? '分享你的想法，构建更好的社区'
    : `正在查看 ${displayName} 的主页`
  const profileInfo = {
    userId: Number(resolvedUserId) || Number(user?.id) || 0,
    gender: 1,
    signature: '热爱分享与探索新技术',
    country: '中国',
    location: '上海',
    birthDay: '1995-05-20',
    createdAt: '2024-01-05T00:00:00Z',
  }
  const formatDate = (value: string | Date) => new Date(value).toLocaleDateString('zh-CN')
  const genderLabel = profileInfo.gender === 1 ? '男' : profileInfo.gender === 2 ? '女' : '保密'
  const actionButtons = (
    <div className="flex items-center space-x-3">
      <Link to="/create" className="btn-primary flex items-center space-x-2">
        <PenSquare className="h-4 w-4" />
        <span>发帖</span>
      </Link>
      <Link to="/settings" className="btn-secondary flex items-center space-x-2">
        <Settings className="h-4 w-4" />
        <span>设置</span>
      </Link>
    </div>
  )

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      <Link
        to="/"
        className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
      >
        <ArrowLeft className="h-5 w-5" />
        <span>返回首页</span>
      </Link>

      <div className="grid lg:grid-cols-[1.2fr_1fr] gap-6">
        {/* 个人信息 */}
        <div className="card lg:col-span-2">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div className="flex items-center space-x-4">
              <img
                src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${displayName}`}
                alt={displayName}
                className="w-16 h-16 rounded-full border-4 border-primary-100"
              />
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{displayName}</h1>
                <p className="text-gray-500 text-sm">{subtitle}</p>
                {resolvedUserId && (
                  <p className="text-xs text-gray-400 mt-1">ID: {resolvedUserId}</p>
                )}
              </div>
            </div>
            {isCurrentUser ? (
              actionButtons
            ) : (
              <div className="flex items-center space-x-3">
                <Link
                  to="/follows"
                  className="btn-secondary flex items-center space-x-2"
                >
                  <HeartHandshake className="h-4 w-4" />
                  <span>关注</span>
                </Link>
                <Link
                  to="/messages"
                  state={{ username: displayName, userId: resolvedUserId }}
                  className="btn-primary flex items-center space-x-2"
                >
                  <Send className="h-4 w-4" />
                  <span>私信</span>
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* 详细信息 */}
        <div className="card lg:col-span-2">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">详细资料</h2>
            <span className="text-xs text-gray-500">示例数据</span>
          </div>
          <div className="grid sm:grid-cols-2 lg:grid-cols-5 gap-4">
            <InfoItem label="性别" value={genderLabel} />
            <InfoItem label="生日" value={formatDate(profileInfo.birthDay)} />
            <InfoItem label="国家" value={profileInfo.country || '—'} />
            <InfoItem label="所在地" value={profileInfo.location || '—'} />
            <InfoItem label="加入时间" value={formatDate(profileInfo.createdAt)} />
          </div>
          <div className="mt-4">
            <InfoItem
              label="个人简介"
              value={profileInfo.signature || '这个人很神秘，还没有简介'}
              full
            />
          </div>
        </div>

        {/* 统计卡片 */}
        <div className="card space-y-4 lg:col-span-2">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">创作概览</h2>
            <span className="text-xs text-gray-500">本周</span>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            <StatItem icon={<PenSquare className="h-5 w-5 text-primary-600" />} label="帖子" value="12" />
            <StatItem icon={<MessageSquare className="h-5 w-5 text-primary-600" />} label="回复" value="48" />
            <StatItem icon={<Heart className="h-5 w-5 text-primary-600" />} label="获赞" value="326" />
            <StatItem icon={<Users className="h-5 w-5 text-primary-600" />} label="关注者" value="89" />
            <StatItem icon={<Share2 className="h-5 w-5 text-primary-600" />} label="分享" value="34" />
          </div>
        </div>

        {/* 最近动态 */}
        <div className="card lg:col-span-2">
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
                className="flex items-center justify-between py-3 hover:bg-gray-50 px-2 rounded-lg transition-colors"
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
    <div className="rounded-xl border border-gray-100 bg-gray-50 px-4 py-3">
      <div className="flex items-center space-x-2 mb-2">
        {icon}
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
    <div className={`border border-gray-100 rounded-lg p-3 bg-gray-50 ${full ? 'w-full' : ''}`}>
      <p className="text-xs text-gray-500">{label}</p>
      <p className="text-sm font-medium text-gray-900 mt-1 break-words">{value}</p>
    </div>
  )
}
