// API 工具函数
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

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
    credentials: 'include', // 关键：确保Cookie随请求发送
  })
}

// 检查响应是否成功，如果不成功则抛出错误
export async function handleResponse<T>(response: Response): Promise<T> {
  const data = await response.json().catch(() => null)
  const msg = (data && (data.msg || data.message)) || `请求失败: ${response.status} ${response.statusText}`

  if (!response.ok || !data || (typeof data.code === 'number' && data.code !== 0)) {
    throw new Error(msg)
  }

  return data as T
}
