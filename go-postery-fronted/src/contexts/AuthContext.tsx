import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User } from '../types'
import { md5Hash } from '../utils/crypto'

interface AuthContextType {
  user: User | null
  login: (username: string, password: string) => Promise<boolean>
  register: (name: string, password: string) => Promise<boolean>
  changePassword: (oldPassword: string, newPassword: string) => Promise<boolean>
  logout: () => void
  isLoading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const AUTH_API_BASE_URL = import.meta.env.VITE_AUTH_API_URL || 'http://localhost:8080'

  // 从 localStorage 恢复登录状态
  useEffect(() => {
    const savedUser = localStorage.getItem('user')
    if (savedUser) {
      try {
        setUser(JSON.parse(savedUser))
      } catch (error) {
        console.error('Failed to parse saved user:', error)
        localStorage.removeItem('user')
      }
    }
    setIsLoading(false)
  }, [])

  const login = async (username: string, password: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const response = await fetch(`${AUTH_API_BASE_URL}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: username, password }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: '登录失败' }))
        throw new Error(errorData.message || '登录失败')
      }

      const data = await response.json()
      
      // 假设后端返回格式: { user: {...}, token: "..." }
      const avatarSeed = data.user.name || data.user.email || username
      const newUser: User = {
        id: data.user.id,
        name: data.user.name,
        email: data.user.email,
        avatar: data.user.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${avatarSeed}`
      }
      
      setUser(newUser)
      // 保存 token 用于后续 API 请求
      if (data.token) {
        localStorage.setItem('token', data.token)
      }
      localStorage.setItem('user', JSON.stringify(newUser))
      setIsLoading(false)
      return true
    } catch (error) {
      console.error('Login error:', error)
      setIsLoading(false)
      // 如果后端不可用，使用模拟登录（仅用于开发演示）
      if (error instanceof TypeError && error.message.includes('fetch')) {
        console.warn('后端 API 不可用，使用模拟登录（仅用于开发）')
        if (username && password) {
          const newUser: User = {
            id: Date.now().toString(),
            name: username,
            email: `${username}@example.com`, // 模拟邮箱
            avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${username}`
          }
          setUser(newUser)
          localStorage.setItem('user', JSON.stringify(newUser))
          setIsLoading(false)
          return true
        }
      }
      return false
    }
  }

  const register = async (name: string, password: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      // TODO: 替换为你的后端 API 地址
      const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
      
      const response = await fetch(`${API_BASE_URL}/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, password }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: '注册失败' }))
        throw new Error(errorData.message || '注册失败')
      }

      const data = await response.json()
      
      // 假设后端返回格式: { user: {...}, token: "..." }
      const newUser: User = {
        id: data.user.id,
        name: data.user.name,
        email: data.user.email || '', // 邮箱变为可选
        avatar: data.user.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${data.user.name}`
      }
      
      setUser(newUser)
      // 保存 token 用于后续 API 请求
      if (data.token) {
        localStorage.setItem('token', data.token)
      }
      localStorage.setItem('user', JSON.stringify(newUser))
      setIsLoading(false)
      return true
    } catch (error) {
      console.error('Register error:', error)
      setIsLoading(false)
      // 如果后端不可用，使用模拟注册（仅用于开发演示）
      if (error instanceof TypeError && error.message.includes('fetch')) {
        console.warn('后端 API 不可用，使用模拟注册（仅用于开发）')
        if (name && password) {
          const newUser: User = {
            id: Date.now().toString(),
            name: name,
            email: '', // 不再需要邮箱
            avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${name}`
          }
          setUser(newUser)
          localStorage.setItem('user', JSON.stringify(newUser))
          setIsLoading(false)
          return true
        }
      }
      return false
    }
  }

  const changePassword = async (oldPassword: string, newPassword: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
      const token = localStorage.getItem('token')
      
      // 对旧密码和新密码进行MD5哈希
      const hashedOldPassword = md5Hash(oldPassword)
      const hashedNewPassword = md5Hash(newPassword)
      
      const response = await fetch(`${API_BASE_URL}/auth/change-password`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
        body: JSON.stringify({ oldPassword: hashedOldPassword, newPassword: hashedNewPassword }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: '修改密码失败' }))
        throw new Error(errorData.message || '修改密码失败')
      }

      setIsLoading(false)
      return true
    } catch (error) {
      console.error('Change password error:', error)
      setIsLoading(false)
      // 如果后端不可用，使用模拟（仅用于开发演示）
      if (error instanceof TypeError && error.message.includes('fetch')) {
        console.warn('后端 API 不可用，使用模拟修改密码（仅用于开发）')
        if (oldPassword && newPassword && newPassword.length >= 6) {
          setIsLoading(false)
          return true
        }
      }
      return false
    }
  }

  const logout = async () => {
    try {
      // 可选：调用后端登出接口
      const token = localStorage.getItem('token')
      if (token) {
        const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
        try {
          await fetch(`${API_BASE_URL}/auth/logout`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          })
        } catch (error) {
          // 忽略登出 API 错误，继续清除本地状态
          console.warn('登出 API 调用失败，但已清除本地状态')
        }
      }
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

