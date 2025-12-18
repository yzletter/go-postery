import { useState, useEffect, useCallback, FormEvent, ReactNode } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Lock,
  Key,
  Save,
  Eye,
  EyeOff,
  Settings as SettingsIcon,
  User as UserIcon,
  Mail,
  Image,
  MapPin,
  Globe,
  Calendar,
} from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { apiGet, apiPost } from '../utils/api'
import { normalizeUserDetail } from '../utils/user'
import type { ModifyUserProfileRequest, UserDetail } from '../types'

export default function Settings() {
  const navigate = useNavigate()
  const { user, changePassword } = useAuth()
  const [activeTab, setActiveTab] = useState<'profile' | 'password'>('profile')
  const [profileEmail, setProfileEmail] = useState(user?.email || '')
  const [avatarUrl, setAvatarUrl] = useState('')
  const [bio, setBio] = useState('')
  const [gender, setGender] = useState<number>(0)
  const [birthday, setBirthday] = useState('')
  const [location, setLocation] = useState('')
  const [country, setCountry] = useState('')
  const [profileSuccess, setProfileSuccess] = useState('')
  const [profileError, setProfileError] = useState('')
  const [isProfileLoading, setIsProfileLoading] = useState(false)
  const [isSavingProfile, setIsSavingProfile] = useState(false)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordError, setPasswordError] = useState('')
  const [passwordSuccess, setPasswordSuccess] = useState('')
  const [isPasswordLoading, setIsPasswordLoading] = useState(false)
  const [showOldPassword, setShowOldPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const profilePath = user?.id ? `/users/${user.id}` : '/profile'

  const normalizeBirthdayInput = useCallback((value?: string) => {
    if (!value) return ''
    const simple = value.split('T')[0]
    if (/^\d{4}-\d{2}-\d{2}$/.test(simple)) return simple
    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? '' : date.toISOString().slice(0, 10)
  }, [])

  const fetchProfile = useCallback(async () => {
    if (!user?.id) return

    setIsProfileLoading(true)
    setProfileError('')
    try {
      const { data } = await apiGet<UserDetail>(`/users/${user.id}`)
      const detail = data ? normalizeUserDetail(data) : null

      if (detail) {
        setProfileEmail(detail.email || '')
        setAvatarUrl(detail.avatar || '')
        setBio(detail.bio || '')
        setGender(detail.gender ?? 0)
        setBirthday(normalizeBirthdayInput(detail.birthday))
        setLocation(detail.location || '')
        setCountry(detail.country || '')
      } else {
        setProfileEmail('')
        setAvatarUrl('')
        setBio('')
        setGender(0)
        setBirthday('')
        setLocation('')
        setCountry('')
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '获取个人资料失败'
      setProfileError(message)
    } finally {
      setIsProfileLoading(false)
    }
  }, [normalizeBirthdayInput, user?.id])

  useEffect(() => {
    void fetchProfile()
  }, [fetchProfile])

  if (!user) {
    navigate('/login')
    return null
  }

  const avatarPreview =
    (avatarUrl && avatarUrl.trim()) ||
    `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(user.name)}`
  const disableProfileForm = isProfileLoading || isSavingProfile

  const handleProfileSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setProfileSuccess('')
    setProfileError('')

    const payload: ModifyUserProfileRequest = {
      email: profileEmail.trim(),
      avatar: avatarUrl.trim(),
      bio: bio.trim(),
      gender,
      birthday,
      location: location.trim(),
      country: country.trim(),
    }

    setIsSavingProfile(true)

    try {
      await apiPost('/users/me', payload as Record<string, unknown>)
      setProfileSuccess('个人资料已更新')
      await fetchProfile()
    } catch (err) {
      const message = err instanceof Error ? err.message : '更新个人资料失败'
      setProfileError(message)
    } finally {
      setIsSavingProfile(false)
    }
  }

  const handlePasswordSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setPasswordError('')
    setPasswordSuccess('')

    if (newPassword.length < 6) {
      setPasswordError('新密码长度至少为 6 位')
      return
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('两次输入的新密码不一致')
      return
    }

    if (oldPassword === newPassword) {
      setPasswordError('新密码不能与旧密码相同')
      return
    }

    setIsPasswordLoading(true)

    try {
      const ok = await changePassword(oldPassword, newPassword)
      if (ok) {
        setOldPassword('')
        setNewPassword('')
        setConfirmPassword('')
        setPasswordSuccess('密码修改成功')
      } else {
        setPasswordError('修改密码失败，请检查旧密码是否正确')
      }
    } catch (err) {
      setPasswordError('发生错误，请重试')
    } finally {
      setIsPasswordLoading(false)
    }
  }

  return (
    <div className="max-w-6xl mx-auto space-y-6">
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

        <div className="grid md:grid-cols-[220px_1fr] gap-4">
          <div className="border border-gray-100 rounded-xl p-3 space-y-2 bg-gray-50">
            <NavButton
              active={activeTab === 'profile'}
              onClick={() => {
                setActiveTab('profile')
                setPasswordError('')
                setPasswordSuccess('')
              }}
              icon={<UserIcon className="h-4 w-4" />}
              label="个人信息"
            />
            <NavButton
              active={activeTab === 'password'}
              onClick={() => {
                setActiveTab('password')
                setProfileSuccess('')
                setProfileError('')
              }}
              icon={<Lock className="h-4 w-4" />}
              label="修改密码"
            />
          </div>

          <div className="space-y-6">
            {activeTab === 'profile' && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900">个人信息</h2>
                    <p className="text-sm text-gray-500">
                      按 ModifyUserProfileRequest 更新邮箱、头像、个性签名、性别、生日、地区与国家
                    </p>
                  </div>
                  <span className="text-xs text-gray-500">
                    {isProfileLoading ? '资料加载中...' : '修改后将立即生效'}
                  </span>
                </div>

                {profileError && (
                  <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
                    {profileError}
                  </div>
                )}

                {profileSuccess && (
                  <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg text-sm">
                    {profileSuccess}
                  </div>
                )}

                <div className="flex items-center space-x-3 rounded-lg border border-gray-100 bg-gray-50 p-3">
                  <img
                    src={avatarPreview}
                    alt={user.name}
                    className="w-14 h-14 rounded-full border border-white shadow-sm"
                  />
                  <div>
                    <p className="text-sm font-semibold text-gray-900">{user.name}</p>
                    <p className="text-xs text-gray-500">用户 ID：{user.id ?? '—'}</p>
                    <p className="text-xs text-gray-500">头像预览基于填写的 URL</p>
                  </div>
                </div>

                <form onSubmit={handleProfileSubmit} className="space-y-4">
                  <div className="grid md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">邮箱</label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <Mail className="h-5 w-5 text-gray-400" />
                        </div>
                        <input
                          type="email"
                          value={profileEmail}
                          onChange={(e) => setProfileEmail(e.target.value)}
                          className="input pl-10"
                          placeholder="邮箱将作为 email 字段"
                          disabled={disableProfileForm}
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">头像 URL</label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <Image className="h-5 w-5 text-gray-400" />
                        </div>
                        <input
                          type="text"
                          value={avatarUrl}
                          onChange={(e) => setAvatarUrl(e.target.value)}
                          className="input pl-10"
                          placeholder="https://example.com/avatar.png"
                          disabled={disableProfileForm}
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">国家</label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <Globe className="h-5 w-5 text-gray-400" />
                        </div>
                        <input
                          type="text"
                          value={country}
                          onChange={(e) => setCountry(e.target.value)}
                          className="input pl-10"
                          placeholder="如 中国 / 美国"
                          disabled={disableProfileForm}
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">地区</label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <MapPin className="h-5 w-5 text-gray-400" />
                        </div>
                        <input
                          type="text"
                          value={location}
                          onChange={(e) => setLocation(e.target.value)}
                          className="input pl-10"
                          placeholder="如 上海 / 纽约"
                          disabled={disableProfileForm}
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">生日</label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <Calendar className="h-5 w-5 text-gray-400" />
                        </div>
                        <input
                          type="date"
                          value={birthday}
                          onChange={(e) => setBirthday(e.target.value)}
                          className="input pl-10"
                          placeholder="选择生日"
                          disabled={disableProfileForm}
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">性别</label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <UserIcon className="h-5 w-5 text-gray-400" />
                        </div>
                        <select
                          value={gender}
                          onChange={(e) => setGender(Number(e.target.value))}
                          className="input pl-10"
                          disabled={disableProfileForm}
                        >
                          <option value={0}>保密 / 未设置</option>
                          <option value={1}>男</option>
                          <option value={2}>女</option>
                          <option value={3}>其他</option>
                        </select>
                      </div>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">个性签名</label>
                    <textarea
                      value={bio}
                      onChange={(e) => setBio(e.target.value)}
                      className="input h-24 resize-none"
                      placeholder="用一段话介绍自己，这会落入 bio 字段"
                      disabled={disableProfileForm}
                      maxLength={160}
                    />
                    <p className="mt-1 text-xs text-gray-500">支持 160 字以内，提交时带上 bio 字段</p>
                  </div>

                  <div className="flex items-center justify-between pt-2">
                    <div className="text-xs text-gray-500">
                      {isSavingProfile ? '正在保存到服务器...' : '保存后刷新个人主页即可查看变更'}
                    </div>
                    <button
                      type="submit"
                      disabled={disableProfileForm}
                      className="btn-primary flex items-center space-x-2"
                    >
                      {isSavingProfile ? (
                        <>
                          <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                          <span>保存中...</span>
                        </>
                      ) : (
                        <>
                          <Save className="h-5 w-5" />
                          <span>保存信息</span>
                        </>
                      )}
                    </button>
                  </div>
                </form>
              </div>
            )}

            {activeTab === 'password' && (
              <div className="space-y-4">
                <div>
                  <h2 className="text-lg font-semibold text-gray-900 mb-1">修改密码</h2>
                  <p className="text-sm text-gray-500">建议定期更换密码，确保账户安全</p>
                </div>

                {passwordError && (
                  <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
                    {passwordError}
                  </div>
                )}

                {passwordSuccess && (
                  <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg text-sm">
                    {passwordSuccess}
                  </div>
                )}

                <form onSubmit={handlePasswordSubmit} className="space-y-4">
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
                      disabled={isPasswordLoading}
                      className="btn-primary flex items-center justify-center space-x-2 w-full"
                    >
                      {isPasswordLoading ? (
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
            )}
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

function NavButton({
  active,
  onClick,
  icon,
  label,
}: {
  active: boolean
  onClick: () => void
  icon: ReactNode
  label: string
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center space-x-2 px-3 py-2 rounded-lg text-sm transition-colors ${
        active ? 'bg-primary-50 text-primary-700 font-semibold' : 'text-gray-700 hover:bg-gray-50'
      }`}
    >
      <span className="flex items-center space-x-2">
        {icon}
        <span>{label}</span>
      </span>
    </button>
  )
}
