import { Link, NavLink, Outlet } from 'react-router-dom'
import { ArrowLeft, FileText, MessageSquare, Shield, Users } from 'lucide-react'
import type { ReactNode } from 'react'
import { useAuth } from '../../contexts/AuthContext'

const NavItem = ({
  to,
  icon,
  label,
}: {
  to: string
  icon: ReactNode
  label: string
}) => (
  <NavLink
    to={to}
    className={({ isActive }) =>
      `w-full flex items-center space-x-2 px-3 py-2 rounded-lg text-sm transition-colors ${
        isActive ? 'bg-primary-50 text-primary-700 font-semibold' : 'text-gray-700 hover:bg-gray-50'
      }`
    }
    end
  >
    <span className="flex items-center space-x-2">
      {icon}
      <span>{label}</span>
    </span>
  </NavLink>
)

export default function AdminLayout() {
  const { user } = useAuth()

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <Link
          to="/"
          className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
        >
          <ArrowLeft className="h-5 w-5" />
          <span>返回首页</span>
        </Link>
        <div className="text-sm text-gray-600">
          当前管理员：<span className="font-semibold text-gray-900">{user?.name || '未知用户'}</span>
        </div>
      </div>

      <div className="card">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">管理员后台</h1>
            <p className="text-gray-600">管理用户、帖子与评论（能力以服务端接口为准）</p>
          </div>
          <div className="w-12 h-12 rounded-full bg-primary-50 text-primary-700 flex items-center justify-center">
            <Shield className="h-6 w-6" />
          </div>
        </div>

        <div className="grid md:grid-cols-[220px_1fr] gap-4">
          <aside className="border border-gray-100 rounded-xl p-3 space-y-2 bg-gray-50">
            <NavItem to="posts" icon={<FileText className="h-4 w-4" />} label="帖子管理" />
            <NavItem to="comments" icon={<MessageSquare className="h-4 w-4" />} label="评论管理" />
            <NavItem to="users" icon={<Users className="h-4 w-4" />} label="用户管理" />
          </aside>

          <section className="space-y-6">
            <Outlet />
          </section>
        </div>
      </div>
    </div>
  )
}
