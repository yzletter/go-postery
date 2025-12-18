import { Link } from 'react-router-dom'
import { ShieldAlert } from 'lucide-react'
import { useAuth } from '../../contexts/AuthContext'
import { normalizeId } from '../../utils/id'

export default function AdminForbidden() {
  const { user } = useAuth()
  const userId = normalizeId(user?.id)

  return (
    <div className="card p-6">
      <div className="flex items-start gap-3">
        <div className="w-10 h-10 rounded-xl bg-red-50 text-red-700 flex items-center justify-center flex-shrink-0">
          <ShieldAlert className="h-5 w-5" />
        </div>
        <div className="space-y-2 min-w-0">
          <h2 className="text-lg font-semibold text-gray-900">无权限访问管理员后台</h2>
          <p className="text-sm text-gray-600">
            当前账号：<span className="font-semibold text-gray-900">{user?.name || '未登录'}</span>
            {userId ? <span className="text-gray-500">（#{userId}）</span> : null}
          </p>
          <div className="text-sm text-gray-600 space-y-1">
            <p>解决方法：</p>
            <p>1）使用用户名为 <span className="font-semibold text-gray-900">admin</span> 的账号登录</p>
            <p>
              2）或在项目根目录创建 <span className="font-semibold text-gray-900">.env.local</span>，配置
              <span className="font-semibold text-gray-900"> VITE_ADMIN_NAMES</span> /
              <span className="font-semibold text-gray-900"> VITE_ADMIN_IDS</span>，然后重启前端
            </p>
          </div>
          <div className="flex flex-wrap gap-2 pt-2">
            <Link to="/login" className="btn-secondary !py-2">
              去登录
            </Link>
            <Link to="/" className="btn-secondary !py-2">
              返回首页
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}

