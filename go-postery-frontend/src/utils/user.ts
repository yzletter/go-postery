import type { UserDetail } from '../types'
import { normalizeId } from './id'

// Normalize user detail response to front-end friendly shape.
export const normalizeUserDetail = (raw: any): UserDetail => {
  const genderRaw = raw?.gender ?? raw?.Gender
  const parsedGender = typeof genderRaw === 'string' ? Number.parseInt(genderRaw, 10) : genderRaw
  const normalizedGender = Number.isFinite(parsedGender) ? Number(parsedGender) : 0

  return {
    id: normalizeId(raw?.id ?? raw?.Id),
    name: raw?.name ?? raw?.Name ?? '未命名用户',
    email: raw?.email ?? raw?.Email,
    avatar: raw?.avatar ?? raw?.Avatar,
    bio: raw?.bio ?? raw?.Bio,
    gender: normalizedGender,
    birthday: raw?.birthday ?? raw?.BirthDay ?? raw?.birthDay,
    location: raw?.location ?? raw?.Location,
    country: raw?.country ?? raw?.Country,
    lastLoginIP: raw?.last_login_ip ?? raw?.lastLoginIP ?? raw?.LastLoginIP,
  }
}
