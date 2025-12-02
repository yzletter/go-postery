import { useState, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Lock, Key, Save, Eye, EyeOff, Settings as SettingsIcon } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

export default function Settings() {
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
  const profilePath = user?.id ? `/users/${user.id}` : '/profile'

  if (!user) {
    navigate('/login')
    return null
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('')
    setSuccess('')

    if (newPassword.length < 6) {
      setError('新密码长度至少为 6 位')
      return
    }

    if (newPassword !== confirmPassword) {
      setError('两次输入的新密码不一致')
      return
    }

    if (oldPassword === newPassword) {
      setError('新密码不能与旧密码相同')
      return
    }

    setIsLoading(true)

    try {
      const ok = await changePassword(oldPassword, newPassword)
      if (ok) {
        setOldPassword('')
        setNewPassword('')
        setConfirmPassword('')
        setSuccess('密码修改成功')
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
      <div className="flex items-center justify-between">
        <Link
          to="/"
          className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
        >
          <ArrowLeft className="h-5 w-5" />
          <span>返回首页</span>
        </Link>
        <Link
          to={profilePath}
          className="text-sm text-primary-600 hover:text-primary-700 font-medium"
        >
          返回个人主页
        </Link>
      </div>

      <div className="card">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">设置</h1>
            <p className="text-gray-600">管理账户与安全选项</p>
          </div>
          <div className="w-12 h-12 rounded-full bg-primary-50 text-primary-700 flex items-center justify-center">
            <SettingsIcon className="h-6 w-6" />
          </div>
        </div>

        <div className="space-y-6">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 mb-2">账户信息</h2>
            <div className="grid sm:grid-cols-2 gap-4 bg-gray-50 border border-gray-100 rounded-xl p-4">
              <div>
                <p className="text-xs text-gray-500">用户名</p>
                <p className="text-gray-900 font-medium mt-1">{user.name}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500">用户ID</p>
                <p className="text-gray-900 font-medium mt-1">{user.id}</p>
              </div>
            </div>
          </div>

          <div className="border-t border-gray-100 pt-4">
            <h2 className="text-lg font-semibold text-gray-900 mb-2">修改密码</h2>
            <p className="text-sm text-gray-500 mb-4">建议定期更换密码，确保账户安全</p>

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
              <PasswordField
                id="oldPassword"
                label="当前密码"
                value={oldPassword}
                onChange={setOldPassword}
                show={showOldPassword}
                setShow={setShowOldPassword}
              />
              <PasswordField
                id="newPassword"
                label="新密码"
                value={newPassword}
                onChange={setNewPassword}
                show={showNewPassword}
                setShow={setShowNewPassword}
                helper="密码长度至少为 6 位"
              />
              <PasswordField
                id="confirmPassword"
                label="确认新密码"
                value={confirmPassword}
                onChange={setConfirmPassword}
                show={showConfirmPassword}
                setShow={setShowConfirmPassword}
              />

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
          </div>
        </div>
      </div>
    </div>
  )
}

function PasswordField({
  id,
  label,
  value,
  onChange,
  show,
  setShow,
  helper,
}: {
  id: string
  label: string
  value: string
  onChange: (v: string) => void
  show: boolean
  setShow: (v: boolean) => void
  helper?: string
}) {
  return (
    <div>
      <label htmlFor={id} className="block text-sm font-medium text-gray-700 mb-2">
        {label}
      </label>
      <div className="relative">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          {id === 'oldPassword' ? <Lock className="h-5 w-5 text-gray-400" /> : <Key className="h-5 w-5 text-gray-400" />}
        </div>
        <input
          id={id}
          type={show ? 'text' : 'password'}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={label}
          required
          minLength={6}
          className="input pl-10 pr-10"
        />
        <button
          type="button"
          onClick={() => setShow(!show)}
          className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
        >
          {show ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
        </button>
      </div>
      {helper && <p className="mt-1 text-xs text-gray-500">{helper}</p>}
    </div>
  )
}
