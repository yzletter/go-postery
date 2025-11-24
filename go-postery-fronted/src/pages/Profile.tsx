import { useState, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Lock, Key, Save, Eye, EyeOff } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

export default function Profile() {
  const navigate = useNavigate()
  const { user, changePassword } = useAuth()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [showOldPassword, setShowOldPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

  if (!user) {
    navigate('/login')
    return null
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('')
    setSuccess('')

    // 验证新密码长度
    if (newPassword.length < 6) {
      setError('新密码长度至少为 6 位')
      return
    }

    // 验证两次输入的新密码是否一致
    if (newPassword !== confirmPassword) {
      setError('两次输入的新密码不一致')
      return
    }

    // 验证新旧密码不能相同
    if (oldPassword === newPassword) {
      setError('新密码不能与旧密码相同')
      return
    }

    setIsLoading(true)

    try {
      const success = await changePassword(oldPassword, newPassword)
      if (success) {
        setSuccess('密码修改成功！')
        setOldPassword('')
        setNewPassword('')
        setConfirmPassword('')
        // 3秒后清除成功消息
        setTimeout(() => setSuccess(''), 3000)
      } else {
        setError('修改密码失败，请检查旧密码是否正确')
      }
    } catch (err) {
      setError('发生错误，请重试')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* 返回按钮 */}
      <Link
        to="/"
        className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
      >
        <ArrowLeft className="h-5 w-5" />
        <span>返回首页</span>
      </Link>

      <div className="grid md:grid-cols-3 gap-6">
        {/* 左侧：用户信息卡片 */}
        <div className="md:col-span-1">
          <div className="card text-center">
            <div className="mb-4">
              <img
                src={user.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${user.name}`}
                alt={user.name}
                className="w-24 h-24 rounded-full mx-auto border-4 border-primary-100"
              />
            </div>
            <h2 className="text-2xl font-bold text-gray-900 mb-2">{user.name}</h2>
          </div>
        </div>

        {/* 右侧：修改密码表单 */}
        <div className="md:col-span-2">
          <div className="card">
            <div className="mb-6">
              <h1 className="text-2xl font-bold text-gray-900 mb-2">修改密码</h1>
              <p className="text-gray-600">为了账户安全，请定期更换密码</p>
            </div>

            {error && (
              <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
                {error}
              </div>
            )}

            {success && (
              <div className="mb-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg text-sm">
                {success}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              {/* 旧密码 */}
              <div>
                <label htmlFor="oldPassword" className="block text-sm font-medium text-gray-700 mb-2">
                  当前密码
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Lock className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="oldPassword"
                    type={showOldPassword ? 'text' : 'password'}
                    value={oldPassword}
                    onChange={(e) => setOldPassword(e.target.value)}
                    placeholder="输入当前密码"
                    required
                    className="input pl-10 pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowOldPassword(!showOldPassword)}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                  >
                    {showOldPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                  </button>
                </div>
              </div>

              {/* 新密码 */}
              <div>
                <label htmlFor="newPassword" className="block text-sm font-medium text-gray-700 mb-2">
                  新密码
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Key className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="newPassword"
                    type={showNewPassword ? 'text' : 'password'}
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    placeholder="输入新密码（至少 6 位）"
                    required
                    minLength={6}
                    className="input pl-10 pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowNewPassword(!showNewPassword)}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                  >
                    {showNewPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                  </button>
                </div>
                <p className="mt-1 text-xs text-gray-500">密码长度至少为 6 位</p>
              </div>

              {/* 确认新密码 */}
              <div>
                <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-2">
                  确认新密码
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Key className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="confirmPassword"
                    type={showConfirmPassword ? 'text' : 'password'}
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder="再次输入新密码"
                    required
                    minLength={6}
                    className="input pl-10 pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                  >
                    {showConfirmPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                  </button>
                </div>
              </div>

              {/* 提交按钮 */}
              <div className="pt-4 border-t border-gray-200">
                <button
                  type="submit"
                  disabled={isLoading}
                  className="btn-primary flex items-center justify-center space-x-2 w-full"
                >
                  {isLoading ? (
                    <>
                      <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                      <span>修改中...</span>
                    </>
                  ) : (
                    <>
                      <Save className="h-5 w-5" />
                      <span>保存新密码</span>
                    </>
                  )}
                </button>
              </div>
            </form>

            {/* 安全提示 */}
            <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <h3 className="text-sm font-medium text-blue-900 mb-2">安全提示</h3>
              <ul className="text-xs text-blue-700 space-y-1">
                <li>• 密码长度至少为 6 位</li>
                <li>• 建议使用字母、数字和特殊字符的组合</li>
                <li>• 不要使用过于简单的密码（如 123456）</li>
                <li>• 定期更换密码以提高账户安全性</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

