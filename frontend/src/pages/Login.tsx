import { useState, useEffect, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Lock, User, LogIn, Eye, EyeOff, Phone, KeyRound } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { apiPost, AUTH_API_BASE_URL } from '../utils/api'

type ActiveTab = 'code' | 'password'

export default function Login() {
  const navigate = useNavigate()
  const { loginWithPassword, loginWithPhone } = useAuth()
  const [activeTab, setActiveTab] = useState<ActiveTab>('password')
  const [loginPhone, setLoginPhone] = useState('')
  const [loginCode, setLoginCode] = useState('')
  const [loginCodeCountdown, setLoginCodeCountdown] = useState(0)
  const [loginAccount, setLoginAccount] = useState('')
  const [loginPassword, setLoginPassword] = useState('')
  const [showLoginPassword, setShowLoginPassword] = useState(false)
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isSendingCode, setIsSendingCode] = useState(false)

  useEffect(() => {
    if (loginCodeCountdown <= 0) {
      return undefined
    }
    const timer = window.setInterval(() => {
      setLoginCodeCountdown((prev) => (prev > 0 ? prev - 1 : 0))
    }, 1000)
    return () => window.clearInterval(timer)
  }, [loginCodeCountdown])

  const handleSendLoginCode = async () => {
    setError('')
    const trimmedPhone = loginPhone.trim()
    if (!trimmedPhone) {
      setError('请输入手机号')
      return
    }
    setIsSendingCode(true)
    try {
      await apiPost(
        '/auth/sms',
        { phone: trimmedPhone },
        { baseUrl: AUTH_API_BASE_URL, skipAuthToken: true }
      )
      setLoginCodeCountdown(60)
    } catch (err) {
      const message = err instanceof Error ? err.message : '验证码发送失败'
      setError(message)
    } finally {
      setIsSendingCode(false)
    }
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('')

    if (activeTab === 'code') {
      const trimmedPhone = loginPhone.trim()
      const trimmedCode = loginCode.trim()

      if (!trimmedPhone) {
        setError('请输入手机号')
        return
      }
      if (!trimmedCode) {
        setError('请输入验证码')
        return
      }

      setIsLoading(true)
      try {
        const success = await loginWithPhone(trimmedPhone, trimmedCode)

        if (success) {
          navigate('/')
        } else {
          setError('登录失败，请检查手机号或验证码')
        }
      } catch (err) {
        setError('发生错误，请重试')
      } finally {
        setIsLoading(false)
      }
      return
    }

    const trimmedAccount = loginAccount.trim()

    if (!trimmedAccount) {
      setError('请输入手机号或邮箱')
      return
    }
    if (!loginPassword) {
      setError('请输入密码')
      return
    }

    setIsLoading(true)
    try {
      const success = await loginWithPassword(trimmedAccount, loginPassword)

      if (success) {
        navigate('/')
      } else {
        setError('登录失败，请检查账号和密码')
      }
    } catch (err) {
      setError('发生错误，请重试')
    } finally {
      setIsLoading(false)
    }
  }

  const title = '欢迎回来'
  const subtitle = activeTab === 'code' ? '使用手机号验证码登录' : '使用账号密码登录'

  return (
    <div className="min-h-[calc(100vh-8rem)] flex items-center justify-center">
      <div className="w-full max-w-md">
        <Link
          to="/"
          className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors mb-6"
        >
          <ArrowLeft className="h-5 w-5" />
          <span>返回首页</span>
        </Link>

        <div className="card">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-gray-900 mb-2">{title}</h1>
            <p className="text-gray-600">{subtitle}</p>
          </div>

          {/* 切换登录方式 */}
          <div className="flex bg-gray-100 rounded-lg p-1 mb-6">
            <button
              type="button"
              onClick={() => {
                setActiveTab('code')
                setError('')
              }}
              className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'code'
                  ? 'bg-white text-primary-600 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              验证码登录
            </button>
            <button
              type="button"
              onClick={() => {
                setActiveTab('password')
                setError('')
              }}
              className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'password'
                  ? 'bg-white text-primary-600 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              密码登录
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
                {error}
              </div>
            )}

            {activeTab === 'code' && (
              <>
                <div>
                  <label htmlFor="login-phone" className="block text-sm font-medium text-gray-700 mb-2">
                    手机号
                  </label>
                  <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <Phone className="h-5 w-5 text-gray-400" />
                    </div>
                    <input
                      id="login-phone"
                      type="tel"
                      autoComplete="tel"
                      value={loginPhone}
                      onChange={(e) => setLoginPhone(e.target.value)}
                      placeholder="输入手机号"
                      required
                      className="input pl-10"
                    />
                  </div>
                </div>

                <div>
                  <label htmlFor="login-code" className="block text-sm font-medium text-gray-700 mb-2">
                    验证码
                  </label>
                  <div className="flex gap-3">
                    <div className="relative flex-1">
                      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <KeyRound className="h-5 w-5 text-gray-400" />
                      </div>
                      <input
                        id="login-code"
                        type="text"
                        inputMode="numeric"
                        autoComplete="one-time-code"
                        value={loginCode}
                        onChange={(e) => setLoginCode(e.target.value)}
                        placeholder="输入验证码"
                        required
                        className="input pl-10"
                      />
                    </div>
                    <button
                      type="button"
                      onClick={handleSendLoginCode}
                      disabled={loginCodeCountdown > 0 || isLoading || isSendingCode}
                      className="btn-secondary whitespace-nowrap px-4"
                    >
                      {loginCodeCountdown > 0 ? `${loginCodeCountdown}s后重试` : isSendingCode ? '发送中...' : '发送验证码'}
                    </button>
                  </div>
                </div>
              </>
            )}

            {activeTab === 'password' && (
              <>
                <div>
                  <label htmlFor="login-account" className="block text-sm font-medium text-gray-700 mb-2">
                    手机号或邮箱
                  </label>
                  <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <User className="h-5 w-5 text-gray-400" />
                    </div>
                    <input
                      id="login-account"
                      type="text"
                      autoComplete="username"
                      value={loginAccount}
                      onChange={(e) => setLoginAccount(e.target.value)}
                      placeholder="输入手机号或邮箱"
                      required
                      className="input pl-10"
                    />
                  </div>
                </div>

                <div>
                  <label htmlFor="login-password" className="block text-sm font-medium text-gray-700 mb-2">
                    密码
                  </label>
                  <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <Lock className="h-5 w-5 text-gray-400" />
                    </div>
                    <input
                      id="login-password"
                      type={showLoginPassword ? 'text' : 'password'}
                      autoComplete="current-password"
                      value={loginPassword}
                      onChange={(e) => setLoginPassword(e.target.value)}
                      placeholder="输入密码"
                      required
                      minLength={6}
                      className="input pl-10 pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowLoginPassword(!showLoginPassword)}
                      className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                    >
                      {showLoginPassword ? (
                        <EyeOff className="h-5 w-5" />
                      ) : (
                        <Eye className="h-5 w-5" />
                      )}
                    </button>
                  </div>
                </div>
              </>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className="w-full btn-primary flex items-center justify-center space-x-2"
            >
              {isLoading ? (
                <>
                  <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>处理中...</span>
                </>
              ) : (
                <>
                  <LogIn className="h-5 w-5" />
                  <span>{activeTab === 'code' ? '登录 / 注册' : '登录'}</span>
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
