import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import type { User } from '../types'
import { md5Hash } from '../utils/crypto'
import { apiGet, apiPost, AUTH_API_BASE_URL } from '../utils/api'
import { normalizeId } from '../utils/id'
import { normalizeUserDetail } from '../utils/user'

interface AuthContextType {
  user: User | null
  loginWithPassword: (identifier: string, password: string) => Promise<boolean>
  loginWithPhone: (phone: string, code: string) => Promise<boolean>
  changePassword: (oldPassword: string, newPassword: string) => Promise<boolean>
  updateUser: (updates: Partial<User>) => void
  logout: () => void
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
    id: resolvedId || Date.now().toString(),
    name: resolvedName,
    email: responseUser.email ?? responseUser.Email,
    avatar: responseUser.avatar ?? responseUser.Avatar ?? raw?.avatar ?? raw?.Avatar,
  }
}

const normalizePasswordInput = (value: string) => {
  const trimmed = value.trim()
  if (!trimmed) return ''
  return trimmed.length === 32 ? trimmed : md5Hash(trimmed)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  // 从 localStorage 恢复登录状态
  useEffect(() => {
    const savedUser = localStorage.getItem('user')
    if (savedUser) {
      try {
        const parsed = JSON.parse(savedUser)
        const normalized = parsed?.id ? { ...parsed, id: normalizeId(parsed.id) } : parsed
        setUser(normalized)
      } catch (error) {
        console.error('Failed to parse saved user:', error)
        localStorage.removeItem('user')
      }
    }
    setIsLoading(false)
  }, [])

  const persistUser = (rawUser: any, fallbackName: string) => {
    const newUser = normalizeUserFromResponse(rawUser, fallbackName)
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

  const logout = async () => {
    try {
      await apiPost('/auth/logout', null, { baseUrl: AUTH_API_BASE_URL })
    } catch (error) {
      console.warn('登出 API 调用失败，但已清除本地状态', error)
    } finally {
      setUser(null)
      localStorage.removeItem('user')
      localStorage.removeItem('token')
    }
  }

  return (
    <AuthContext.Provider value={{ user, loginWithPassword, loginWithPhone, changePassword, updateUser, logout, isLoading }}>
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
