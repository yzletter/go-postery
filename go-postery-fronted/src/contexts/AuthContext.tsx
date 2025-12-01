import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User, ApiResponse } from '../types'
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
  const AUTH_API_BASE_URL =
    import.meta.env.VITE_AUTH_API_URL ||
    import.meta.env.VITE_API_BASE_URL ||
    'http://localhost:8080'

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
      const payloadPassword = password.length === 32 ? password : md5Hash(password)

      const response = await fetch(`${AUTH_API_BASE_URL}/login/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: username, password: payloadPassword }),
        credentials: 'include', // 关键：确保Cookie随请求发送
      })

      const result: ApiResponse = await response.json()
      
      if (!response.ok || result.code !== 0) {
        throw new Error(result.msg || '登录失败')
      }

      const responseData = result.data || {}
      const responseUser = responseData.user || {}
      const newUser: User = {
        id: responseUser.id ?? responseUser.Id ?? Date.now(),
        name: responseUser.name ?? username,
        email: responseUser.email,
      }

      setUser(newUser)
      localStorage.setItem('user', JSON.stringify(newUser))
      localStorage.setItem('token', `cookie-auth-${Date.now()}`)
      
      setIsLoading(false)
      return true
    } catch (error) {
      console.error('Login error:', error)
      setIsLoading(false)
      return false
    }
  }

  const register = async (name: string, password: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      const payloadPassword = password.length === 32 ? password : md5Hash(password)

      const response = await fetch(`${AUTH_API_BASE_URL}/register/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, password: payloadPassword }),
        credentials: 'include', // 关键：确保Cookie随请求发送
      })

      const result: ApiResponse = await response.json()
      
      if (!response.ok || result.code !== 0) {
        throw new Error(result.msg || '注册失败')
      }

      const responseData = result.data || {}
      const responseUser = responseData.user || {}
      const newUser: User = {
        id: responseUser.id ?? responseUser.Id ?? Date.now(),
        name: responseUser.name ?? name,
        email: responseUser.email,
      }

      setUser(newUser)
      localStorage.setItem('user', JSON.stringify(newUser))
      localStorage.setItem('token', `cookie-auth-${Date.now()}`)
      setIsLoading(false)
      return true
    } catch (error) {
      console.error('Register error:', error)
      setIsLoading(false)
      return false
    }
  }

  const changePassword = async (oldPassword: string, newPassword: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      // 对于Cookie认证，主要依赖浏览器自动发送Cookie
      // 同时尝试从localStorage获取token作为备选
      const token = localStorage.getItem('token')
      
      console.log('修改密码：尝试发送认证请求...')
      
      // 对密码进行MD5哈希处理
      const hashedOldPassword = md5Hash(oldPassword)
      const hashedNewPassword = md5Hash(newPassword)
      
      const response = await fetch(`${AUTH_API_BASE_URL}/modify_pass/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          // 如果有token，添加到Authorization头；否则依赖Cookie
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
        body: JSON.stringify({ old_pass: hashedOldPassword, new_pass: hashedNewPassword }),
        credentials: 'include', // 关键：确保Cookie随请求发送
      })

      const result: ApiResponse = await response.json()
      
      if (result.code !== 0) {
        throw new Error(result.msg || '修改密码失败')
      }

      setIsLoading(false)
      return true
    } catch (error) {
      console.error('Change password error:', error)
      setIsLoading(false)
      // 如果后端不可用或响应格式错误，使用模拟逻辑（仅用于开发演示）
      if (error instanceof TypeError && error.message.includes('fetch')) {
        console.warn('后端 API 不可用，使用模拟修改密码（仅用于开发）')
        console.warn('模拟Cookie认证：浏览器会自动发送JWT Cookie')
        if (oldPassword && newPassword) {
          // 对输入密码进行MD5哈希，与后端逻辑保持一致
          const hashedOldPassword = md5Hash(oldPassword)
          const hashedNewPassword = md5Hash(newPassword)
          
          // 模拟验证：旧密码不能是123456的MD5哈希值
          if (hashedOldPassword === 'e10adc3949ba59abbe56e057f20f883e') { // 123456的MD5值
            console.warn('模拟修改密码：旧密码验证失败')
            return false
          }
          console.warn('模拟修改密码：密码修改成功')
          console.warn(`旧密码哈希: ${hashedOldPassword}, 新密码哈希: ${hashedNewPassword}`)
          return true
        }
      }
      // 处理响应格式错误的情况，也使用模拟逻辑
      if (error instanceof Error && error.message.includes('响应不是JSON格式')) {
        console.warn('后端响应格式错误，使用模拟修改密码（仅用于开发）')
        if (oldPassword && newPassword) {
          // 对输入密码进行MD5哈希，与后端逻辑保持一致
          const hashedOldPassword = md5Hash(oldPassword)
          const hashedNewPassword = md5Hash(newPassword)
          
          // 模拟验证：旧密码不能是123456的MD5哈希值
          if (hashedOldPassword === 'e10adc3949ba59abbe56e057f20f883e') { // 123456的MD5值
            console.warn('模拟修改密码：旧密码验证失败')
            return false
          }
          console.warn('模拟修改密码：密码修改成功')
          console.warn(`旧密码哈希: ${hashedOldPassword}, 新密码哈希: ${hashedNewPassword}`)
          return true
        }
      }
      return false
    }
  }

  const logout = async () => {
    try {
      // 调用后端登出接口 /logout
      try {
        await fetch(`${AUTH_API_BASE_URL}/logout`, {
          method: 'GET',
          credentials: 'include', // 关键：确保Cookie随请求发送
        })
      } catch (error) {
        // 忽略登出 API 错误，继续清除本地状态
        console.warn('登出 API 调用失败，但已清除本地状态')
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
