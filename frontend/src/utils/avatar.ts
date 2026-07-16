import { apiPost } from './api'
import { buildIdSeed, normalizeId } from './id'

const USER_AVATAR_PREFIX = 'users/avatar/'
// The BFF currently issues avatar URLs that are valid for 30 minutes. Keep a
// small safety margin while avoiding a new presign request on every remount.
const AVATAR_CACHE_TTL_MS = 25 * 60 * 1000
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

const getCachedObjectUrl = (objectKey: string) => {
  const cached = avatarUrlCache.get(objectKey)
  if (!cached) return ''
  if (cached.expiresAt > Date.now()) return cached.url
  return ''
}

// Direct URLs and valid presign-cache entries can be used during the initial
// render, before React effects run.
export const getImmediateAvatarUrl = (avatar?: string | null) => {
  const normalizedAvatar = normalizeAvatarValue(avatar)
  if (!normalizedAvatar) return ''
  if (isDirectAvatarUrl(normalizedAvatar)) return normalizedAvatar
  if (!isAvatarObjectKey(normalizedAvatar)) return ''
  return getCachedObjectUrl(normalizedAvatar)
}

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

  const immediateUrl = getImmediateAvatarUrl(normalizedAvatar)
  if (immediateUrl) return immediateUrl

  if (!isAvatarObjectKey(normalizedAvatar)) {
    throw new Error('invalid avatar object key')
  }

  // Cache inspection can also happen during React render, so expiry cleanup is
  // intentionally performed here instead of in the synchronous read helper.
  const expiredCache = avatarUrlCache.get(normalizedAvatar)
  if (expiredCache && expiredCache.expiresAt <= Date.now()) {
    avatarUrlCache.delete(normalizedAvatar)
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
