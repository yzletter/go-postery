import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import type { User } from '../types'
import { md5Hash } from '../utils/crypto'
import {
  ApiError,
  apiGet,
  apiPost,
  AUTH_API_BASE_URL,
  clearAuthToken,
  onUnauthorized,
} from '../utils/api'
import { normalizeId } from '../utils/id'
import { normalizeUserDetail } from '../utils/user'

interface AuthContextType {
  user: User | null
  loginWithPassword: (identifier: string, password: string) => Promise<boolean>
  loginWithPhone: (phone: string, code: string) => Promise<boolean>
  changePassword: (oldPassword: string, newPassword: string) => Promise<boolean>
  setPassword: (newPassword: string, code: string) => Promise<boolean>
  updateUser: (updates: Partial<User>) => void
  logout: () => Promise<void>
  isLoading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

const normalizeUserFromResponse = (raw: any, fallbackName: string): User => {
  const responseUser = raw?.user ?? raw ?? {}
  const resolvedId = normalizeId(
    responseUser.id ??
      responseUser.Id ??
      responseUser.ID ??
      raw?.Id ??
      raw?.ID ??
      raw?.id
  )
  const resolvedName =
    responseUser.nickname ??
    responseUser.Nickname ??
    responseUser.name ??
    responseUser.Name ??
    raw?.nickname ??
    raw?.Nickname ??
    raw?.name ??
    raw?.Name ??
    fallbackName
  return {
    id: resolvedId || undefined,
    name: String(resolvedName ?? '').trim(),
    avatar: responseUser.avatar ?? responseUser.Avatar ?? raw?.avatar ?? raw?.Avatar,
  }
}

const normalizePasswordInput = (value: string) => {
  if (!value) return ''
  return /^[a-f\d]{32}$/i.test(value) ? value : md5Hash(value)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const clearLocalAuth = useCallback(() => {
    setUser(null)
    localStorage.removeItem('user')
    clearAuthToken()
  }, [])

  // 任意接口返回 401 时，同步清理全局登录态。
  useEffect(() => {
    return onUnauthorized(clearLocalAuth)
  }, [clearLocalAuth])

  // 本地用户信息只作为身份展示缓存，启动时仍以 BFF 登录状态为准。
  useEffect(() => {
    let isActive = true

    const restoreSession = async () => {
      const savedUser = localStorage.getItem('user')
      if (!savedUser) {
        clearAuthToken()
        if (isActive) setIsLoading(false)
        return
      }

      try {
        const parsed = JSON.parse(savedUser)
        const normalized = normalizeUserFromResponse(parsed, '')
        if (!normalized.id || !normalized.name) {
          throw new Error('本地用户信息不完整')
        }

        try {
          await apiGet('/auth/status', { baseUrl: AUTH_API_BASE_URL })
        } catch (error) {
          if (error instanceof ApiError && error.status === 401) {
            return
          }
          // 后端临时不可用时保留已有身份，后续请求仍会在 401 时统一退出。
          console.warn('Failed to verify saved session:', error)
        }

        if (isActive) {
          setUser(normalized)
          localStorage.setItem('user', JSON.stringify(normalized))
        }
      } catch (error) {
        console.error('Failed to parse saved user:', error)
        if (isActive) {
          clearLocalAuth()
        }
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
      }
    }

    void restoreSession()

    return () => {
      isActive = false
    }
  }, [clearLocalAuth])

  const persistUser = (rawUser: any, fallbackName: string) => {
    const newUser = normalizeUserFromResponse(rawUser, fallbackName)
    if (!newUser.id || !newUser.name) {
      clearLocalAuth()
      throw new Error('登录响应缺少用户身份信息')
    }
    setUser(newUser)
    localStorage.setItem('user', JSON.stringify(newUser))
    return newUser
  }

  const updateUser = useCallback((updates: Partial<User>) => {
    setUser((prev) => {
      if (!prev) return prev
      const next = { ...prev, ...updates }
      localStorage.setItem('user', JSON.stringify(next))
      return next
    })
  }, [])

  useEffect(() => {
    if (!user?.id || user.avatar) return

    let cancelled = false

    apiGet(`/users/${user.id}`)
      .then(({ data }) => {
        if (cancelled || !data) return
        const detail = normalizeUserDetail(data)
        if (!detail.avatar) return
        updateUser({ avatar: detail.avatar })
      })
      .catch((error) => {
        console.warn('Failed to hydrate user avatar:', error)
      })

    return () => {
      cancelled = true
    }
  }, [updateUser, user?.avatar, user?.id])

  const loginWithPassword = async (identifier: string, password: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const trimmedIdentifier = identifier.trim()
      const payloadPassword = normalizePasswordInput(password)
      const { data } = await apiPost('/auth/login/password', { identifier: trimmedIdentifier, password: payloadPassword }, {
        baseUrl: AUTH_API_BASE_URL,
        skipAuthToken: true,
      })

      persistUser(data, trimmedIdentifier)
      return true
    } catch (error) {
      console.error('Login error:', error)
      return false
    } finally {
      setIsLoading(false)
    }
  }

  const loginWithPhone = async (phone: string, code: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const trimmedPhone = phone.trim()
      const trimmedCode = code.trim()
      const { data } = await apiPost('/auth/login/phone', { phone: trimmedPhone, code: trimmedCode }, {
        baseUrl: AUTH_API_BASE_URL,
        skipAuthToken: true,
      })

      persistUser(data, trimmedPhone)
      return true
    } catch (error) {
      console.error('Login phone error:', error)
      return false
    } finally {
      setIsLoading(false)
    }
  }

  const changePassword = async (oldPassword: string, newPassword: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const hashedOldPassword = normalizePasswordInput(oldPassword)
      const hashedNewPassword = normalizePasswordInput(newPassword)
      
      await apiPost(
        '/auth/password/update',
        { old_password: hashedOldPassword, new_password: hashedNewPassword },
        { baseUrl: AUTH_API_BASE_URL }
      )
      return true
    } catch (error) {
      console.error('Change password error:', error)
      return false
    } finally {
      setIsLoading(false)
    }
  }

  const setPassword = async (newPassword: string, code: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      await apiPost(
        '/auth/password/set',
        { new_password: normalizePasswordInput(newPassword), code: code.trim() },
        { baseUrl: AUTH_API_BASE_URL }
      )
      return true
    } catch (error) {
      console.error('Set password error:', error)
      return false
    } finally {
      setIsLoading(false)
    }
  }

  const logout = useCallback(async () => {
    try {
      await apiPost('/auth/logout', null, { baseUrl: AUTH_API_BASE_URL })
    } catch (error) {
      console.warn('登出 API 调用失败，但已清除本地状态', error)
    } finally {
      clearLocalAuth()
    }
  }, [clearLocalAuth])

  return (
    <AuthContext.Provider value={{ user, loginWithPassword, loginWithPhone, changePassword, setPassword, updateUser, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
