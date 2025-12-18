import type { FollowRelation, FollowUser } from '../types'
import { apiDelete, apiGet, apiPost } from './api'
import { normalizeId } from './id'

const normalizeFollowUser = (raw: any): FollowUser | null => {
  if (!raw) return null
  const id = normalizeId(raw.id ?? raw.Id)
  const name = raw.name ?? raw.Name ?? ''
  const avatar = raw.avatar ?? raw.Avatar ?? ''

  if (!id || !name) return null
  return { id, name, avatar: avatar || undefined }
}

export async function listFollowers(): Promise<FollowUser[]> {
  const { data } = await apiGet<{
    followers: unknown[]
    total?: number
    hasMore?: boolean
  }>('/users/me/followers?pageNo=1&pageSize=100')
  const rawList = Array.isArray(data?.followers) ? data.followers : []
  return rawList.map(normalizeFollowUser).filter((u): u is FollowUser => Boolean(u))
}

export async function listFollowees(): Promise<FollowUser[]> {
  const { data } = await apiGet<{
    followees: unknown[]
    total?: number
    hasMore?: boolean
  }>('/users/me/followees?pageNo=1&pageSize=100')
  const rawList = Array.isArray(data?.followees) ? data.followees : []
  return rawList.map(normalizeFollowUser).filter((u): u is FollowUser => Boolean(u))
}

export async function followUser(targetUserId: string): Promise<void> {
  const id = normalizeId(targetUserId)
  await apiPost(`/users/${encodeURIComponent(id)}/follow`, null)
}

export async function unfollowUser(targetUserId: string): Promise<void> {
  const id = normalizeId(targetUserId)
  await apiDelete(`/users/${encodeURIComponent(id)}/follow`)
}

export async function getFollowRelation(targetUserId: string): Promise<FollowRelation> {
  const id = normalizeId(targetUserId)
  const { data } = await apiGet<unknown>(`/users/${encodeURIComponent(id)}/follow`)
  const parsed = typeof data === 'number' ? data : Number(data)
  return parsed === 0 || parsed === 1 || parsed === 2 || parsed === 3 ? (parsed as FollowRelation) : 0
}

export const isFollowing = (relation: FollowRelation) => relation === 1 || relation === 3
