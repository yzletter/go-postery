import { useState, useEffect, useCallback, FormEvent, ReactNode, type ChangeEvent } from 'react'
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
  Image,
  MapPin,
  Globe,
  Calendar,
} from 'lucide-react'
import UserAvatar from '../components/UserAvatar'
import { useAuth } from '../contexts/AuthContext'
import { apiGet, apiPost, AUTH_API_BASE_URL } from '../utils/api'
import { normalizeUserDetail } from '../utils/user'
import type { ModifyUserProfileRequest, UserDetail } from '../types'

const AVATAR_MAX_FILE_SIZE = 2 * 1024 * 1024
const AVATAR_MIN_DIMENSION = 128
const AVATAR_MAX_DIMENSION = 2048
const AVATAR_ALLOWED_TYPES: Record<string, string> = {
  'image/jpeg': 'jpg',
  'image/png': 'png',
  'image/webp': 'webp',
}

type AvatarUploadPolicy = {
  policy?: string
  signature?: string
  x_oss_signature_version?: string
  x_oss_credential?: string
  x_oss_date?: string
  security_token?: string
  host?: string
  dir?: string
  callback?: string
}

type AvatarUploadSignResponse = {
  response?: string
}

type PasswordStatusResponse = {
  has_password: boolean
}

type AuthIdentityResponse = {
  phone: string
  email: string
}

const AVATAR_SIZE_LABEL = `${Math.round(AVATAR_MAX_FILE_SIZE / 1024 / 1024)}MB`
const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null

const readAvatarDimensions = (file: File) =>
  new Promise<{ width: number; height: number }>((resolve, reject) => {
    const objectUrl = URL.createObjectURL(file)
    const image = new window.Image()

    image.onload = () => {
      const dimensions = {
        width: image.naturalWidth,
        height: image.naturalHeight,
      }
      URL.revokeObjectURL(objectUrl)
      resolve(dimensions)
    }

    image.onerror = () => {
      URL.revokeObjectURL(objectUrl)
      reject(new Error('无法读取图片尺寸，请重新选择文件'))
    }

    image.src = objectUrl
  })

const validateAvatarFile = async (file: File) => {
  if (!AVATAR_ALLOWED_TYPES[file.type]) {
    throw new Error('头像仅支持 JPG、PNG 或 WebP 格式')
  }

  if (file.size > AVATAR_MAX_FILE_SIZE) {
    throw new Error(`头像大小不能超过 ${AVATAR_SIZE_LABEL}`)
  }

  const { width, height } = await readAvatarDimensions(file)
  if (width < AVATAR_MIN_DIMENSION || height < AVATAR_MIN_DIMENSION) {
    throw new Error(`头像尺寸至少需要 ${AVATAR_MIN_DIMENSION} x ${AVATAR_MIN_DIMENSION}`)
  }

  if (width > AVATAR_MAX_DIMENSION || height > AVATAR_MAX_DIMENSION) {
    throw new Error(`头像尺寸不能超过 ${AVATAR_MAX_DIMENSION} x ${AVATAR_MAX_DIMENSION}`)
  }
}

const parseAvatarUploadPolicy = (raw: unknown): Required<AvatarUploadPolicy> => {
  if (typeof raw !== 'string' || !raw.trim()) {
    throw new Error('头像上传签名无效')
  }

  const parsed = JSON.parse(raw) as AvatarUploadPolicy
  const policy = parsed.policy?.trim()
  const signature = parsed.signature?.trim()
  const signatureVersion = parsed.x_oss_signature_version?.trim() || 'OSS4-HMAC-SHA256'
  const credential = parsed.x_oss_credential?.trim()
  const date = parsed.x_oss_date?.trim()
  const securityToken = parsed.security_token?.trim()
  const host = parsed.host?.trim()
  const dir = parsed.dir?.trim()
  const callback = parsed.callback?.trim()

  if (!policy || !signature || !credential || !date || !securityToken || !host || !dir || !callback) {
    throw new Error('头像上传签名字段不完整')
  }

  return {
    policy,
    signature,
    x_oss_signature_version: signatureVersion,
    x_oss_credential: credential,
    x_oss_date: date,
    security_token: securityToken,
    host,
    dir,
    callback,
  }
}

const buildAvatarObjectKey = (dir: string, file: File) => {
  const extension = AVATAR_ALLOWED_TYPES[file.type] || 'png'
  const suffix = Math.random().toString(36).slice(2, 8)
  return `${dir}avatar-${Date.now()}-${suffix}.${extension}`
}

const uploadAvatarFile = async (file: File) => {
  const { data } = await apiGet<AvatarUploadSignResponse>('/users/me/upload')
  const policy = parseAvatarUploadPolicy(data?.response)
  const objectKey = buildAvatarObjectKey(policy.dir, file)

  const formData = new FormData()
  formData.append('success_action_status', '200')
  formData.append('policy', policy.policy)
  formData.append('x-oss-signature', policy.signature)
  formData.append('x-oss-signature-version', policy.x_oss_signature_version)
  formData.append('x-oss-credential', policy.x_oss_credential)
  formData.append('x-oss-date', policy.x_oss_date)
  formData.append('key', objectKey)
  formData.append('x-oss-security-token', policy.security_token)
  formData.append('callback', policy.callback)
  formData.append('file', file)

  const response = await fetch(policy.host, {
    method: 'POST',
    body: formData,
  })

  const responseText = await response.text()
  if (response.status !== 200) {
    throw new Error(
      response.status === 203
        ? '头像已上传，但后端回调保存失败'
        : '头像上传失败，请稍后重试'
    )
  }

  if (responseText.trim()) {
    let callbackResult: unknown
    try {
      callbackResult = JSON.parse(responseText) as unknown
    } catch {
      throw new Error('头像回调响应格式错误')
    }

    if (isRecord(callbackResult)) {
      const code = Number(callbackResult.code)
      if (Number.isFinite(code) && code !== 0) {
        const message =
          typeof callbackResult.msg === 'string' ? callbackResult.msg.trim() : ''
        throw new Error(message || '头像保存失败')
      }
    }
  }

  return objectKey
}

export default function Settings() {
  const navigate = useNavigate()
  const { user, changePassword, setPassword, updateUser } = useAuth()
  const [activeTab, setActiveTab] = useState<'profile' | 'password'>('profile')
  const [nickname, setNickname] = useState(user?.name || '')
  const [avatarObjectKey, setAvatarObjectKey] = useState('')
  const [avatarPreviewUrl, setAvatarPreviewUrl] = useState('')
  const [bio, setBio] = useState('')
  const [gender, setGender] = useState<number>(0)
  const [birthday, setBirthday] = useState('')
  const [location, setLocation] = useState('')
  const [country, setCountry] = useState('')
  const [profileSuccess, setProfileSuccess] = useState('')
  const [profileError, setProfileError] = useState('')
  const [isProfileLoading, setIsProfileLoading] = useState(false)
  const [isSavingProfile, setIsSavingProfile] = useState(false)
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordError, setPasswordError] = useState('')
  const [passwordSuccess, setPasswordSuccess] = useState('')
  const [isPasswordLoading, setIsPasswordLoading] = useState(false)
  const [isPasswordStatusLoading, setIsPasswordStatusLoading] = useState(false)
  const [hasPassword, setHasPassword] = useState(true)
  const [boundPhone, setBoundPhone] = useState('')
  const [verificationCode, setVerificationCode] = useState('')
  const [isSendingCode, setIsSendingCode] = useState(false)
  const [codeCountdown, setCodeCountdown] = useState(0)
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

  const updateAvatarPreview = useCallback((nextUrl: string) => {
    setAvatarPreviewUrl((prev) => {
      if (prev && prev !== nextUrl) {
        URL.revokeObjectURL(prev)
      }
      return nextUrl
    })
  }, [])

  useEffect(() => {
    return () => {
      if (avatarPreviewUrl) {
        URL.revokeObjectURL(avatarPreviewUrl)
      }
    }
  }, [avatarPreviewUrl])

  useEffect(() => {
    if (activeTab !== 'password') return

    let cancelled = false
    setIsPasswordStatusLoading(true)
    setPasswordError('')

    Promise.all([
      apiGet<PasswordStatusResponse>('/auth/password/status', { baseUrl: AUTH_API_BASE_URL }),
      apiGet<AuthIdentityResponse>('/auth/auth_identity', { baseUrl: AUTH_API_BASE_URL }),
    ])
      .then(([statusResponse, identityResponse]) => {
        if (cancelled) return
        setHasPassword(statusResponse.data?.has_password ?? true)
        setBoundPhone(identityResponse.data?.phone?.trim() ?? '')
      })
      .catch((err) => {
        if (cancelled) return
        setPasswordError(err instanceof Error ? err.message : '获取密码状态失败')
      })
      .finally(() => {
        if (!cancelled) setIsPasswordStatusLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [activeTab])

  useEffect(() => {
    if (codeCountdown <= 0) return undefined
    const timer = window.setInterval(() => {
      setCodeCountdown((prev) => (prev > 0 ? prev - 1 : 0))
    }, 1000)
    return () => window.clearInterval(timer)
  }, [codeCountdown])

  const handleSendPasswordCode = async () => {
    setPasswordError('')
    if (!boundPhone) {
      setPasswordError('当前账号未绑定手机号，无法设置密码')
      return
    }

    setIsSendingCode(true)
    try {
      await apiPost(
        '/auth/sms',
        { phone: boundPhone },
        { baseUrl: AUTH_API_BASE_URL, skipAuthToken: true }
      )
      setCodeCountdown(60)
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : '验证码发送失败')
    } finally {
      setIsSendingCode(false)
    }
  }

  const fetchProfile = useCallback(async () => {
    if (!user?.id) return

    setIsProfileLoading(true)
    setProfileError('')
    try {
      const { data } = await apiGet<UserDetail>(`/users/${user.id}`)
      const detail = data ? normalizeUserDetail(data) : null

      if (detail) {
        setNickname(detail.name || user?.name || '')
        setAvatarObjectKey(detail.avatar || '')
        updateAvatarPreview('')
        setBio(detail.bio || '')
        setGender(detail.gender ?? 0)
        setBirthday(normalizeBirthdayInput(detail.birthday))
        setLocation(detail.location || '')
        setCountry(detail.country || '')
        updateUser({
          name: detail.name || user?.name || '',
          avatar: detail.avatar || undefined,
        })
      } else {
        setNickname(user?.name || '')
        setAvatarObjectKey('')
        updateAvatarPreview('')
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
  }, [normalizeBirthdayInput, updateAvatarPreview, updateUser, user?.id, user?.name])

  useEffect(() => {
    void fetchProfile()
  }, [fetchProfile])

  if (!user) {
    navigate('/login')
    return null
  }

  const displayName = nickname.trim() || user?.name || '用户'
  const avatarPreview = avatarPreviewUrl || avatarObjectKey
  const disableProfileForm = isProfileLoading || isSavingProfile || isUploadingAvatar

  const handleAvatarChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) return

    setProfileSuccess('')
    setProfileError('')
    setIsUploadingAvatar(true)

    try {
      await validateAvatarFile(file)
      const objectKey = await uploadAvatarFile(file)
      updateAvatarPreview(URL.createObjectURL(file))
      setAvatarObjectKey(objectKey)
      updateUser({ avatar: objectKey })
      setProfileSuccess('头像上传成功')
    } catch (err) {
      const message = err instanceof Error ? err.message : '头像上传失败'
      setProfileError(message)
    } finally {
      setIsUploadingAvatar(false)
    }
  }

  const handleProfileSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setProfileSuccess('')
    setProfileError('')

    const normalizedNickname = nickname.trim()
    if (!normalizedNickname) {
      setProfileError('昵称不能为空')
      return
    }

    const payload: ModifyUserProfileRequest = {
      nickname: normalizedNickname,
      avatar: avatarObjectKey.trim(),
      bio: bio.trim(),
      gender,
      birthday,
      location: location.trim(),
      country: country.trim(),
    }

    setIsSavingProfile(true)

    try {
      await apiPost('/users/me', payload as Record<string, unknown>)
      updateAvatarPreview('')
      updateUser({
        name: normalizedNickname,
        avatar: avatarObjectKey.trim() || undefined,
      })
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

    if (hasPassword && oldPassword === newPassword) {
      setPasswordError('新密码不能与旧密码相同')
      return
    }

    if (!hasPassword && !verificationCode.trim()) {
      setPasswordError('请输入验证码')
      return
    }

    setIsPasswordLoading(true)

    try {
      const ok = hasPassword
        ? await changePassword(oldPassword, newPassword)
        : await setPassword(newPassword, verificationCode)
      if (ok) {
        setOldPassword('')
        setNewPassword('')
        setConfirmPassword('')
        setVerificationCode('')
        setPasswordSuccess(hasPassword ? '密码修改成功' : '密码设置成功')
        setHasPassword(true)
      } else {
        setPasswordError(hasPassword ? '修改密码失败，请检查旧密码是否正确' : '设置密码失败，请检查验证码是否正确')
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
                  <UserAvatar
                    avatar={avatarPreview}
                    name={displayName}
                    userId={user.id}
                    className="w-14 h-14 rounded-full border border-white shadow-sm"
                  />
                  <div>
                    <p className="text-sm font-semibold text-gray-900">{displayName}</p>
                    <p className="text-xs text-gray-500">用户 ID：{user.id ?? '—'}</p>
                    <p className="text-xs text-gray-500">
                      {isUploadingAvatar ? '头像上传中...' : '头像展示会通过 presign 临时地址加载'}
                    </p>
                  </div>
                </div>

                <form onSubmit={handleProfileSubmit} className="space-y-4">
                  <div className="grid md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">昵称</label>
                      <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                          <UserIcon className="h-5 w-5 text-gray-400" />
                        </div>
                        <input
                          type="text"
                          value={nickname}
                          onChange={(e) => setNickname(e.target.value)}
                          className="input pl-10"
                          placeholder="展示给其他用户的昵称"
                          disabled={disableProfileForm}
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">上传头像</label>
                      <div className="rounded-xl border border-gray-200 bg-gray-50 p-3 space-y-2">
                        <div className="flex items-center gap-2 text-sm text-gray-700">
                          <Image className="h-4 w-4 text-gray-400" />
                          <span>支持 JPG / PNG / WebP</span>
                        </div>
                        <input
                          type="file"
                          accept="image/jpeg,image/png,image/webp"
                          onChange={handleAvatarChange}
                          className="block w-full text-sm text-gray-600 file:mr-3 file:rounded-lg file:border-0 file:bg-white file:px-3 file:py-2 file:text-sm file:font-medium file:text-primary-700 hover:file:bg-primary-50"
                          disabled={disableProfileForm}
                        />
                        <p className="text-xs text-gray-500">
                          大小不超过 {AVATAR_SIZE_LABEL}，尺寸需在 {AVATAR_MIN_DIMENSION} x {AVATAR_MIN_DIMENSION} 到 {AVATAR_MAX_DIMENSION} x {AVATAR_MAX_DIMENSION} 之间
                        </p>
                        {avatarObjectKey ? (
                          <p className="text-[11px] text-gray-400 break-all">对象键：{avatarObjectKey}</p>
                        ) : null}
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
                      {isUploadingAvatar
                        ? '头像已单独上传，正在等待上传完成...'
                        : isSavingProfile
                          ? '正在保存到服务器...'
                          : '保存后刷新个人主页即可查看变更'}
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
                  <h2 className="text-lg font-semibold text-gray-900 mb-1">
                    {hasPassword ? '修改密码' : '设置密码'}
                  </h2>
                  <p className="text-sm text-gray-500">
                    {hasPassword ? '建议定期更换密码，确保账户安全' : '设置密码后可使用手机号或邮箱和密码登录'}
                  </p>
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

                {isPasswordStatusLoading ? (
                  <div className="py-8 text-center text-sm text-gray-500">正在获取密码状态...</div>
                ) : (
                <form onSubmit={handlePasswordSubmit} className="space-y-4">
                  {hasPassword ? (
                    <PasswordField
                      id="oldPassword"
                      label="当前密码"
                      value={oldPassword}
                      onChange={setOldPassword}
                      show={showOldPassword}
                      setShow={setShowOldPassword}
                    />
                  ) : (
                    <div>
                      <label htmlFor="verificationCode" className="block text-sm font-medium text-gray-700 mb-2">
                        手机验证码
                      </label>
                      <div className="flex gap-3">
                        <input
                          id="verificationCode"
                          type="text"
                          value={verificationCode}
                          onChange={(e) => setVerificationCode(e.target.value)}
                          placeholder={boundPhone ? `验证码将发送至 ${boundPhone}` : '未绑定手机号'}
                          required
                          className="input flex-1"
                        />
                        <button
                          type="button"
                          onClick={handleSendPasswordCode}
                          disabled={!boundPhone || isSendingCode || codeCountdown > 0}
                          className="btn-secondary whitespace-nowrap"
                        >
                          {codeCountdown > 0 ? `${codeCountdown}s后重试` : isSendingCode ? '发送中...' : '发送验证码'}
                        </button>
                      </div>
                    </div>
                  )}
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
                          <span>{hasPassword ? '修改中...' : '设置中...'}</span>
                        </>
                      ) : (
                        <>
                          <Save className="h-5 w-5" />
                          <span>{hasPassword ? '保存新密码' : '设置密码'}</span>
                        </>
                      )}
                    </button>
                  </div>
                </form>
                )}
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
