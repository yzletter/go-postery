import type { ApiResponse } from '../types'

const ensureApiPrefix = (rawBaseUrl: string) => {
  const base = rawBaseUrl.replace(/\/+$/, '')
  return base.endsWith('/api/v1') ? base : `${base}/api/v1`
}

export const API_BASE_URL = ensureApiPrefix(import.meta.env.VITE_API_BASE_URL || 'http://localhost:8765')
export const AUTH_API_BASE_URL = ensureApiPrefix(import.meta.env.VITE_AUTH_API_URL || API_BASE_URL)

export type ApiRequestOptions = Omit<RequestInit, 'body'> & {
  body?: BodyInit | Record<string, unknown> | null
  baseUrl?: string
  skipAuthToken?: boolean
}

export function getAuthToken(): string | null {
  return localStorage.getItem('token')
}

function normalizeBearerToken(value: string): string {
  const trimmed = value.trim()
  if (!trimmed) return ''
  return trimmed.replace(/^Bearer\s+/i, '').trim()
}

function persistAuthToken(token: string | null) {
  if (!token) {
    localStorage.removeItem('token')
    return
  }

  const normalized = normalizeBearerToken(token)
  if (!normalized) {
    localStorage.removeItem('token')
    return
  }

  localStorage.setItem('token', normalized)
}

function buildHeaders(options: ApiRequestOptions, token: string | null, isJsonBody: boolean) {
  const normalizedToken = token ? normalizeBearerToken(token) : ''
  const headers: HeadersInit = {
    ...(isJsonBody ? { 'Content-Type': 'application/json' } : {}),
    ...(normalizedToken ? { Authorization: `Bearer ${normalizedToken}` } : {}),
    ...options.headers,
  }

  return headers
}

export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {}
): Promise<ApiResponse<T>> {
  const {
    baseUrl = API_BASE_URL,
    skipAuthToken,
    body,
    ...rest
  } = options

  const token = skipAuthToken ? null : getAuthToken()
  const isFormData = body instanceof FormData
  const normalizedBody: BodyInit | undefined =
    body === undefined || body === null
      ? undefined
      : isFormData
        ? body
        : typeof body === 'string'
          ? body
          : JSON.stringify(body)

  const headers = buildHeaders(options, token, !isFormData && normalizedBody !== undefined)

  const response = await fetch(`${baseUrl}${path}`, {
    ...rest,
    headers,
    credentials: 'include',
    body: normalizedBody,
  })

  const nextToken =
    response.headers.get('Authorization') ??
    response.headers.get('x-jwt-token')
  if (nextToken !== null) {
    persistAuthToken(nextToken)
  }
  if (response.status === 401) {
    persistAuthToken(null)
  }

  let payload: ApiResponse<T> | null = null
  try {
    payload = await response.json()
  } catch (error) {
    payload = null
  }

  const isSuccess = response.ok && payload && (payload.code === 0 || payload.code === undefined)
  if (!isSuccess || !payload) {
    const message = payload?.msg || `请求失败: ${response.status} ${response.statusText}`
    throw new Error(message)
  }

  return payload
}

export function apiGet<T>(path: string, options: ApiRequestOptions = {}) {
  return apiRequest<T>(path, { ...options, method: 'GET' })
}

export function apiPost<T>(
  path: string,
  body?: BodyInit | Record<string, unknown> | null,
  options: ApiRequestOptions = {}
) {
  return apiRequest<T>(path, { ...options, method: 'POST', body })
}

export function apiDelete<T>(path: string, options: ApiRequestOptions = {}) {
  return apiRequest<T>(path, { ...options, method: 'DELETE' })
}
