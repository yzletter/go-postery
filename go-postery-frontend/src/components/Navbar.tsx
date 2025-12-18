import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { MessageSquare, Plus, LogOut, LogIn, User, Search, Settings, Bot, HeartHandshake, Send, Sparkles, Shield } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { isAdminUser } from '../utils/admin'

export default function Navbar() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [showMobileSearch, setShowMobileSearch] = useState(false)
  const profileLink = user?.id ? `/users/${user.id}` : '/profile'
  const showAdmin = isAdminUser(user)

  return (
    <nav className="sticky top-0 z-50 border-b border-gray-200/60 bg-white/70 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center h-16 gap-4 md:gap-8">
          <Link
            to="/"
            reloadDocument
            className="flex items-center gap-2 group flex-shrink-0 lg:-ml-20"
          >
            <MessageSquare className="h-8 w-8 text-primary-600 group-hover:text-primary-700 transition-colors" />
            <span className="text-xl sm:text-2xl font-bold text-gray-900 group-hover:text-primary-700 transition-colors">
              Go Postery
            </span>
          </Link>

          <form
            className="hidden md:block flex-1 max-w-xl lg:ml-20"
            onSubmit={(e) => {
              e.preventDefault()
              const query = searchTerm.trim()
              setShowMobileSearch(false)
              navigate(query ? `/search?q=${encodeURIComponent(query)}` : '/search')
            }}
          >
            <div className="relative">
              <Search className="h-5 w-5 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="搜索帖子标题、内容或作者..."
                className="input w-full pl-10 pr-4 h-11"
              />
            </div>
          </form>
          
          <div className="flex items-center gap-2 sm:gap-3 flex-shrink-0 ml-auto">
            <button
              type="button"
              onClick={() => setShowMobileSearch(prev => !prev)}
              className="md:hidden btn-secondary !px-3 !py-2"
              aria-label={showMobileSearch ? '关闭搜索' : '打开搜索'}
            >
              <Search className="h-5 w-5" />
            </button>
            <Link
              to="/agent"
              className="relative overflow-hidden group flex items-center gap-2 px-3.5 py-2 rounded-full border border-primary-100 bg-gradient-to-r from-primary-50 via-white to-white text-primary-800 shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all"
            >
              <span className="absolute inset-0 bg-primary-100/40 opacity-0 group-hover:opacity-100 transition-opacity blur-xl" aria-hidden />
              <Bot className="h-5 w-5 relative z-10" />
              <span className="hidden sm:inline font-semibold relative z-10">AI 助手</span>
              <span className="hidden sm:inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full bg-white/70 border border-primary-100 text-primary-700 relative z-10">
                <Sparkles className="h-3 w-3" />
                新
              </span>
            </Link>
            {showAdmin && (
              <Link
                to="/admin"
                className="btn-secondary flex items-center space-x-2"
              >
                <Shield className="h-5 w-5" />
                <span className="hidden sm:inline">后台</span>
              </Link>
            )}
            {user ? (
              <>
                <Link
                  to="/create"
                  className="btn-primary flex items-center space-x-2"
                >
                  <Plus className="h-5 w-5" />
                  <span className="hidden sm:inline">发帖</span>
                </Link>
                {/* 用户菜单 */}
                <div className="relative">
                  <button
                    onClick={() => setShowUserMenu(!showUserMenu)}
                    className="flex items-center space-x-2 px-3 py-2 rounded-xl bg-white/60 ring-1 ring-gray-200/70 shadow-sm hover:bg-white transition-colors"
                  >
                    <img
                      src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${user.name}`}
                      alt={user.name}
                      className="w-8 h-8 rounded-full"
                    />
                    <span className="hidden md:inline text-gray-700 font-medium">
                      {user.name}
                    </span>
                  </button>
                  
                  {showUserMenu && (
                    <>
                      <div
                        className="fixed inset-0 z-10"
                        onClick={() => setShowUserMenu(false)}
                      />
                      <div className="absolute right-0 mt-2 w-52 bg-white/90 backdrop-blur-md rounded-xl shadow-xl ring-1 ring-gray-200/70 py-1 z-20">
                        <div className="px-4 py-2 border-b border-gray-200/60">
                          <p className="text-sm font-medium text-gray-900">{user.name}</p>
                        </div>
                        <Link
                          to={profileLink}
                          onClick={() => setShowUserMenu(false)}
                          className="w-full flex items-center space-x-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        >
                          <User className="h-4 w-4" />
                          <span>个人主页</span>
                        </Link>
                        <Link
                          to="/follows"
                          onClick={() => setShowUserMenu(false)}
                          className="w-full flex items-center space-x-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        >
                          <HeartHandshake className="h-4 w-4" />
                          <span>关注</span>
                        </Link>
                        <Link
                          to="/messages"
                          onClick={() => setShowUserMenu(false)}
                          className="w-full flex items-center space-x-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        >
                          <Send className="h-4 w-4" />
                          <span>私信</span>
                        </Link>
                        <Link
                          to="/settings"
                          onClick={() => setShowUserMenu(false)}
                          className="w-full flex items-center space-x-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        >
                          <Settings className="h-4 w-4" />
                          <span>设置</span>
                        </Link>
                        <button
                          onClick={() => {
                            logout()
                            setShowUserMenu(false)
                            navigate('/')
                          }}
                          className="w-full flex items-center space-x-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        >
                          <LogOut className="h-4 w-4" />
                          <span>退出登录</span>
                        </button>
                      </div>
                    </>
                  )}
                </div>
              </>
            ) : (
              <Link
                to="/login"
                className="btn-secondary flex items-center space-x-2"
              >
                <LogIn className="h-5 w-5" />
                <span className="hidden sm:inline">登录</span>
              </Link>
            )}
          </div>
        </div>

        {showMobileSearch && (
          <div className="md:hidden pb-4">
            <form
              onSubmit={(e) => {
                e.preventDefault()
                const query = searchTerm.trim()
                setShowMobileSearch(false)
                navigate(query ? `/search?q=${encodeURIComponent(query)}` : '/search')
              }}
            >
              <div className="relative">
                <Search className="h-5 w-5 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="搜索帖子标题、内容或作者..."
                  className="input w-full pl-10 pr-4 h-11"
                />
              </div>
            </form>
          </div>
        )}
      </div>
    </nav>
  )
}
