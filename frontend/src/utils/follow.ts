import type { FollowRelation, FollowUser } from '../types'
import { apiGet, apiPost } from './api'
import { normalizeId } from './id'

const normalizeFollowUser = (raw: any): FollowUser | null => {
  if (!raw) return null
  const id = normalizeId(raw.id ?? raw.Id)
  const name =
    raw.nickname ??
    raw.Nickname ??
    raw.name ??
    raw.Name ??
    ''
  const avatar = raw.avatar ?? raw.Avatar ?? ''

  if (!id || !name) return null
  return { id, name, avatar: avatar || undefined }
}

export type FollowListResult = {
  users: FollowUser[]
  total: number
  hasMore: boolean
}

const normalizeTotal = (raw: unknown, fallback: number) => {
  const total = Number(raw)
  return Number.isFinite(total) && total >= 0 ? total : fallback
}

export async function listFollowers(): Promise<FollowListResult> {
  const { data } = await apiGet<{
    followers: unknown[]
    total?: number
    hasMore?: boolean
    has_more?: boolean
  }>('/users/me/followers?pageNo=1&pageSize=100')
  const rawList = Array.isArray(data?.followers) ? data.followers : []
  const users = rawList.map(normalizeFollowUser).filter((u): u is FollowUser => Boolean(u))
  return {
    users,
    total: normalizeTotal(data?.total, users.length),
    hasMore: Boolean(data?.hasMore ?? data?.has_more),
  }
}

export async function listFollowees(): Promise<FollowListResult> {
  const { data } = await apiGet<{
    followees?: unknown[]
    followers?: unknown[]
    total?: number
    hasMore?: boolean
    has_more?: boolean
  }>('/users/me/followees?pageNo=1&pageSize=100')
  const rawList = Array.isArray(data?.followees)
    ? data.followees
    : Array.isArray(data?.followers)
      ? data.followers
      : []
  const users = rawList.map(normalizeFollowUser).filter((u): u is FollowUser => Boolean(u))
  return {
    users,
    total: normalizeTotal(data?.total, users.length),
    hasMore: Boolean(data?.hasMore ?? data?.has_more),
  }
}

export async function followUser(targetUserId: string): Promise<void> {
  const id = normalizeId(targetUserId)
  await apiPost(`/users/${encodeURIComponent(id)}/follow`, null)
}

export async function unfollowUser(targetUserId: string): Promise<void> {
  const id = normalizeId(targetUserId)
  await apiPost(`/users/${encodeURIComponent(id)}/unfollow`, null)
}

export async function getFollowRelation(targetUserId: string): Promise<FollowRelation> {
  const id = normalizeId(targetUserId)
  const { data } = await apiGet<unknown>(`/users/${encodeURIComponent(id)}/follow`)
  const parsed = typeof data === 'number' ? data : Number(data)
  if (parsed === 0 || parsed === 1 || parsed === 2 || parsed === 3) {
    return parsed as FollowRelation
  }
  throw new Error('关注关系响应数据格式错误')
}

export const isFollowing = (relation: FollowRelation) => relation === 1 || relation === 3
