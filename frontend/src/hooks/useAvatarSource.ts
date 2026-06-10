import { useEffect, useMemo, useState } from 'react'
import { buildFallbackAvatarUrl, normalizeAvatarValue, resolveAvatarUrl } from '../utils/avatar'

type UseAvatarSourceOptions = {
  name?: string
  id?: string
  fallbackSeed?: string
}

export const useAvatarSource = (avatar?: string, options: UseAvatarSourceOptions = {}) => {
  const fallbackSrc = useMemo(
    () => buildFallbackAvatarUrl(options),
    [options.fallbackSeed, options.id, options.name]
  )
  const [src, setSrc] = useState(fallbackSrc)

  useEffect(() => {
    let cancelled = false
    const normalizedAvatar = normalizeAvatarValue(avatar)

    setSrc(fallbackSrc)

    if (!normalizedAvatar) {
      return () => {
        cancelled = true
      }
    }

    resolveAvatarUrl(normalizedAvatar)
      .then((nextSrc) => {
        if (!cancelled) {
          setSrc(nextSrc || fallbackSrc)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setSrc(fallbackSrc)
        }
      })

    return () => {
      cancelled = true
    }
  }, [avatar, fallbackSrc])

  return { src, fallbackSrc }
}
