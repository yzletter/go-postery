import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User } from '../types'
import { md5Hash } from '../utils/crypto'
import { apiGet, apiPost, AUTH_API_BASE_URL } from '../utils/api'
import { normalizeId } from '../utils/id'

interface AuthContextType {
  user: User | null
  login: (username: string, password: string) => Promise<boolean>
  register: (name: string, password: string) => Promise<boolean>
  changePassword: (oldPassword: string, newPassword: string) => Promise<boolean>
  logout: () => void
  isLoading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

const normalizeUserFromResponse = (raw: any, fallbackName: string): User => {
  const responseUser = raw?.user ?? raw ?? {}
  const resolvedId = normalizeId(responseUser.id ?? responseUser.Id ?? raw?.Id ?? raw?.id)
  return {
    id: resolvedId || Date.now().toString(),
    name: responseUser.name ?? responseUser.Name ?? raw?.Name ?? fallbackName,
    email: responseUser.email ?? responseUser.Email,
  }
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
    localStorage.setItem('token', `cookie-auth-${Date.now()}`)
    return newUser
  }

  const login = async (username: string, password: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const payloadPassword = password.length === 32 ? password : md5Hash(password)
      const { data } = await apiPost('/login/submit', { name: username, password: payloadPassword }, {
        baseUrl: AUTH_API_BASE_URL,
        skipAuthToken: true,
      })

      persistUser(data, username)
      return true
    } catch (error) {
      console.error('Login error:', error)
      return false
    } finally {
      setIsLoading(false)
    }
  }

  const register = async (name: string, password: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const payloadPassword = password.length === 32 ? password : md5Hash(password)
      const { data } = await apiPost('/register/submit', { name, password: payloadPassword }, {
        baseUrl: AUTH_API_BASE_URL,
        skipAuthToken: true,
      })

      persistUser(data, name)
      return true
    } catch (error) {
      console.error('Register error:', error)
      return false
    } finally {
      setIsLoading(false)
    }
  }

  const changePassword = async (oldPassword: string, newPassword: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const hashedOldPassword = md5Hash(oldPassword)
      const hashedNewPassword = md5Hash(newPassword)
      
      await apiPost('/modify_pass/submit', { old_pass: hashedOldPassword, new_pass: hashedNewPassword }, {
        baseUrl: AUTH_API_BASE_URL,
      })
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
      await apiGet('/logout', { baseUrl: AUTH_API_BASE_URL })
    } catch (error) {
      console.warn('登出 API 调用失败，但已清除本地状态', error)
    } finally {
      setUser(null)
      localStorage.removeItem('user')
      localStorage.removeItem('token')
    }
  }

  return (
    <AuthContext.Provider value={{ user, login, register, changePassword, logout, isLoading }}>
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
