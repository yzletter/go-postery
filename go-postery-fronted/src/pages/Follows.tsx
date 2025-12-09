import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Users, HeartHandshake, ShieldAlert } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

const mockFollowing = [
  { id: 201, name: '前端小能手', desc: 'React / TS / 动效' },
  { id: 202, name: 'Go 语言爱好者', desc: 'Go / 微服务 / 云原生' },
  { id: 203, name: '设计灵感库', desc: 'UI/UX 灵感' },
]

const mockFollowers = [
  { id: 301, name: '测试同学', desc: '质量保障 / 自动化' },
  { id: 302, name: '产品小伙伴', desc: '需求讨论 / 拆解' },
  { id: 303, name: '后端老王', desc: '架构 & 性能' },
]

const mockBlocked = [
  { id: 401, name: '噪音用户A', desc: '已屏蔽' },
  { id: 402, name: '广告账号B', desc: '已屏蔽' },
]

type TabKey = 'following' | 'followers' | 'blocked'

export default function Follows() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<TabKey>('following')

  if (!user) {
    navigate('/login')
    return null
  }

  const tabConfig: Record<
    TabKey,
    { label: string; data: typeof mockFollowing; icon: React.ReactNode }
  > = {
    following: { label: '我关注的', data: mockFollowing, icon: <HeartHandshake className="h-4 w-4" /> },
    followers: { label: '关注我的', data: mockFollowers, icon: <Users className="h-4 w-4" /> },
    blocked: { label: '黑名单', data: mockBlocked, icon: <ShieldAlert className="h-4 w-4" /> },
  }

  const current = tabConfig[activeTab]

  return (
    <div className="max-w-6xl mx-auto space-y-6 lg:-ml-1">
      <div className="flex items-center justify-between">
        <Link
          to="/"
          className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
        >
          <ArrowLeft className="h-5 w-5" />
          <span>返回首页</span>
        </Link>
        <div className="text-sm text-gray-500">
          {user.name} · 关注关系
        </div>
      </div>

      <div className="grid md:grid-cols-[220px_1fr] gap-4">
        <div className="card h-fit p-3 space-y-1">
          {Object.entries(tabConfig).map(([key, item]) => {
            const isActive = key === activeTab
            return (
              <button
                key={key}
                onClick={() => setActiveTab(key as TabKey)}
                className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-primary-50 text-primary-700 font-semibold'
                    : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                <span className="flex items-center space-x-2">
                  {item.icon}
                  <span>{item.label}</span>
                </span>
                <span className="text-xs text-gray-500">{item.data.length}</span>
              </button>
            )
          })}
        </div>

        <div className="card space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              {current.icon}
              <h2 className="text-lg font-semibold text-gray-900">{current.label}</h2>
            </div>
            <span className="text-xs text-gray-500">示例数据</span>
          </div>
          {current.data.length === 0 ? (
            <div className="text-sm text-gray-500">暂无数据</div>
          ) : (
            <div className="space-y-3">
              {current.data.map((item) => (
                <div
                  key={item.id}
                  className="flex items-center space-x-3 p-2 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <Link
                    to={`/users/${item.id}`}
                    state={{ username: item.name }}
                    className="flex-shrink-0"
                  >
                    <img
                      src={`https://api.dicebear.com/7.x/avataaars/svg?seed=follow-${item.id}`}
                      alt={item.name}
                      className="w-10 h-10 rounded-full"
                    />
                  </Link>
                  <div className="flex-1 min-w-0">
                    <Link
                      to={`/users/${item.id}`}
                      state={{ username: item.name }}
                      className="font-medium text-gray-900 hover:text-primary-600 transition-colors line-clamp-1"
                    >
                      {item.name}
                    </Link>
                    <p className="text-xs text-gray-500 line-clamp-1">{item.desc}</p>
                  </div>
                  <button className="text-xs text-primary-600 font-medium hover:text-primary-700">
                    查看
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
