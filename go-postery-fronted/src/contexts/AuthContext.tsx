import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User, ApiResponse } from '../types'
import { md5Hash } from '../utils/crypto'

interface AuthContextType {
  user: User | null
  login: (username: string, password: string) => Promise<boolean>
  register: (name: string, email: string, password: string) => Promise<boolean>
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
      const response = await fetch(`${AUTH_API_BASE_URL}/login/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name: username, password }),
        credentials: 'include', // 关键：确保Cookie随请求发送
      })

      const result: ApiResponse = await response.json()
      
      // 根据API文档：code为0表示成功，1表示失败
      if (result.code !== 0) {
        throw new Error(result.msg || '登录失败')
      }

      // 根据API文档，登录响应中只有user对象，没有token
      const responseData = result.data
      if (!responseData || !responseData.user) {
        throw new Error('登录响应数据格式错误')
      }
      
      const avatarSeed = responseData.user.name || username
      const newUser: User = {
        id: responseData.user.id || Date.now().toString(),
        name: responseData.user.name,
        email: responseData.user.email || `${username}@example.com`,
        avatar: responseData.user.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${avatarSeed}`
      }
      
      setUser(newUser)
      
      // 对于Cookie认证，JWT存储在Cookie中，前端无法直接读取HttpOnly Cookie
      // 但我们可以创建一个模拟token用于本地状态管理
      const mockToken = `cookie-auth-${Date.now()}`
      localStorage.setItem('token', mockToken)
      localStorage.setItem('user', JSON.stringify(newUser))
      
      console.log('登录成功：JWT存储在Cookie中，浏览器会自动发送')
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
          // 模拟Cookie认证：创建模拟token并保存
          const mockToken = `mock-cookie-jwt-${Date.now()}`
          localStorage.setItem('token', mockToken)
          localStorage.setItem('user', JSON.stringify(newUser))
          console.log('模拟登录：JWT令牌已保存到localStorage')
          setIsLoading(false)
          return true
        }
      }
      // 处理响应格式错误的情况，也使用模拟登录
      if (error instanceof Error && error.message.includes('响应数据格式错误')) {
        console.warn('后端响应格式错误，使用模拟登录（仅用于开发）')
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

  const register = async (name: string, email: string, password: string): Promise<boolean> => {
    setIsLoading(true)
    try {
      // 根据API文档，注册只需要name和password，但为了向后兼容性保留email参数
      const response = await fetch(`${AUTH_API_BASE_URL}/register/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, password }),
        credentials: 'include', // 关键：确保Cookie随请求发送
      })

      const result: ApiResponse = await response.json()
      
      // 根据API文档：code为0表示成功，1表示失败
      if (result.code !== 0) {
        throw new Error(result.msg || '注册失败')
      }

      // 根据API文档，注册响应中只有user对象，没有token
      const responseData = result.data
      if (!responseData || !responseData.user) {
        throw new Error('注册响应数据格式错误')
      }
      
      const avatarSeed = responseData.user.name || name
      const newUser: User = {
        id: responseData.user.id || Date.now().toString(),
        name: responseData.user.name,
        email: responseData.user.email,
        avatar: responseData.user.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${avatarSeed}`
      }
      
      setUser(newUser)
      // 注意：根据API文档，注册响应不包含token，所以不保存token
      localStorage.setItem('user', JSON.stringify(newUser))
      setIsLoading(false)
      return true
    } catch (error) {
      console.error('Register error:', error)
      setIsLoading(false)
      // 如果后端不可用，使用模拟注册（仅用于开发演示）
      if (error instanceof TypeError && error.message.includes('fetch')) {
        console.warn('后端 API 不可用，使用模拟注册（仅用于开发）')
        if (name && email && password) {
          const newUser: User = {
            id: Date.now().toString(),
            name: name,
            email: email,
            avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${name}`
          }
          setUser(newUser)
          // 模拟注册时也保存token
          const mockToken = `mock-jwt-token-${Date.now()}`
          localStorage.setItem('token', mockToken)
          localStorage.setItem('user', JSON.stringify(newUser))
          setIsLoading(false)
          return true
        }
      }
      // 处理响应格式错误的情况，也使用模拟注册
      if (error instanceof Error && error.message.includes('响应数据格式错误')) {
        console.warn('后端响应格式错误，使用模拟注册（仅用于开发）')
        if (name && email && password) {
          const newUser: User = {
            id: Date.now().toString(),
            name: name,
            email: email,
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

      // 检查响应状态
      if (!response.ok) {
        throw new Error(`HTTP错误: ${response.status}`)
      }
      
      // 检查内容类型
      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        throw new Error('响应不是JSON格式')
      }

      const result: ApiResponse = await response.json()
      
      // 根据API文档：code为0表示成功，1表示失败
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

