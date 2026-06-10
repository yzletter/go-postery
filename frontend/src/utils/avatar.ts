import { apiPost } from './api'
import { buildIdSeed, normalizeId } from './id'

const USER_AVATAR_PREFIX = 'users/avatar/'
const AVATAR_CACHE_TTL_MS = 10_000
const FALLBACK_AVATAR_STYLES = ['adventurer', 'avataaars', 'big-ears', 'micah', 'notionists']

type AvatarIdentity = {
  name?: string
  id?: string
  fallbackSeed?: string
}

type AvatarPresignResponse = {
  url?: string
}

type AvatarCacheEntry = {
  url: string
  expiresAt: number
}

const avatarUrlCache = new Map<string, AvatarCacheEntry>()
const avatarRequestCache = new Map<string, Promise<string>>()

export const normalizeAvatarValue = (value?: string | null) =>
  typeof value === 'string' ? value.trim() : ''

export const isDirectAvatarUrl = (value: string) =>
  /^(https?:)?\/\//.test(value) || value.startsWith('data:') || value.startsWith('blob:')

export const isAvatarObjectKey = (value: string) => value.startsWith(USER_AVATAR_PREFIX)

export const buildFallbackAvatarUrl = ({ name, id, fallbackSeed }: AvatarIdentity = {}) => {
  const normalizedId = normalizeId(id).trim()
  const seedSource = fallbackSeed?.trim() || normalizedId || name?.trim() || 'user'
  const randomSeed = buildIdSeed(seedSource, seedSource.length || 1)
  const style = FALLBACK_AVATAR_STYLES[randomSeed % FALLBACK_AVATAR_STYLES.length]
  const token = encodeURIComponent(`${seedSource}-${randomSeed}`)
  return `https://api.dicebear.com/7.x/${style}/svg?seed=${token}`
}

export async function resolveAvatarUrl(avatar?: string | null): Promise<string> {
  const normalizedAvatar = normalizeAvatarValue(avatar)
  if (!normalizedAvatar) return ''

  if (isDirectAvatarUrl(normalizedAvatar)) {
    return normalizedAvatar
  }

  if (!isAvatarObjectKey(normalizedAvatar)) {
    throw new Error('invalid avatar object key')
  }

  const now = Date.now()
  const cached = avatarUrlCache.get(normalizedAvatar)
  if (cached && cached.expiresAt > now) {
    return cached.url
  }

  const inflight = avatarRequestCache.get(normalizedAvatar)
  if (inflight) {
    return inflight
  }

  const request = apiPost<AvatarPresignResponse>('/users/presign', { avatar: normalizedAvatar })
    .then(({ data }) => {
      const url = normalizeAvatarValue(data?.url)
      if (!url) {
        throw new Error('empty avatar presign url')
      }
      avatarUrlCache.set(normalizedAvatar, {
        url,
        expiresAt: Date.now() + AVATAR_CACHE_TTL_MS,
      })
      return url
    })
    .finally(() => {
      avatarRequestCache.delete(normalizedAvatar)
    })

  avatarRequestCache.set(normalizedAvatar, request)
  return request
}
