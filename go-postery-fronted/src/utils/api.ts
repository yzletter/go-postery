// API 工具函数
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

// 获取认证 token
export function getAuthToken(): string | null {
  return localStorage.getItem('token')
}

// 创建带认证头的 fetch 请求
export async function authenticatedFetch(
  endpoint: string,
  options: RequestInit = {}
): Promise<Response> {
  const token = getAuthToken()
  
  const headers = {
    'Content-Type': 'application/json',
    ...(token && { 'Authorization': `Bearer ${token}` }),
    ...options.headers,
  }

  return fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })
}

// 检查响应是否成功，如果不成功则抛出错误
export async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({
      message: `请求失败: ${response.status} ${response.statusText}`,
    }))
    throw new Error(errorData.message || '请求失败')
  }
  return response.json()
}

