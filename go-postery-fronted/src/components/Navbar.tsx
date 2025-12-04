import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { MessageSquare, Plus, LogOut, LogIn, User, Search, Settings, Bot, HeartHandshake, Send } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

export default function Navbar() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const profileLink = user?.id ? `/users/${user.id}` : '/profile'

  return (
    <nav className="bg-white border-b border-gray-200 shadow-sm sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-2 sm:px-4 lg:px-5">
        <div className="flex items-center h-16 gap-4">
          <Link
            to="/"
            reloadDocument
            className="flex items-center space-x-2 group flex-shrink-0 -ml-3 sm:-ml-20"
          >
            <MessageSquare className="h-8 w-8 text-primary-600 group-hover:text-primary-700 transition-colors" />
            <span className="text-2xl font-bold text-gray-900 group-hover:text-primary-600 transition-colors">
              Go Postery
            </span>
          </Link>

          <form
            className="flex-1 max-w-xl hidden sm:block sm:ml-20 md:ml-20"
            onSubmit={(e) => {
              e.preventDefault()
            }}
          >
            <div className="relative">
              <Search className="h-5 w-5 text-gray-800 absolute left-3 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="搜索帖子标题、内容或作者..."
                className="input w-full pl-10 pr-4 h-10 bg-gray-50 border-gray-200 focus:border-primary-400 focus:ring-primary-200"
              />
            </div>
          </form>
          
          <div className="flex items-center space-x-3 sm:space-x-4 flex-shrink-0 ml-auto">
            <Link
              to="/agent"
              className="flex items-center space-x-1 text-gray-900 hover:text-primary-600 transition-colors px-3 py-2 rounded-lg hover:bg-gray-50"
            >
              <Bot className="h-5 w-5" />
              <span className="hidden sm:inline">Go Agentery</span>
            </Link>
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
                    className="flex items-center space-x-2 px-3 py-2 rounded-lg hover:bg-gray-50 transition-colors"
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
                      <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-20">
                        <div className="px-4 py-2 border-b border-gray-200">
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
      </div>
    </nav>
  )
}
