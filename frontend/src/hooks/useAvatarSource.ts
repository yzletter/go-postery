import { useEffect, useMemo, useState } from 'react'
import {
  buildFallbackAvatarUrl,
  getImmediateAvatarUrl,
  isAvatarObjectKey,
  normalizeAvatarValue,
  resolveAvatarUrl,
} from '../utils/avatar'

type UseAvatarSourceOptions = {
  name?: string
  id?: string
  fallbackSeed?: string
}

type AvatarSourceState = {
  avatar: string
  src: string
  status: 'fallback' | 'loading' | 'ready' | 'failed'
}

const buildAvatarSourceState = (avatar: string): AvatarSourceState => {
  if (!avatar) {
    return { avatar, src: '', status: 'fallback' }
  }

  const immediateSrc = getImmediateAvatarUrl(avatar)
  if (immediateSrc) {
    return { avatar, src: immediateSrc, status: 'ready' }
  }

  if (isAvatarObjectKey(avatar)) {
    return { avatar, src: '', status: 'loading' }
  }

  return { avatar, src: '', status: 'failed' }
}

export const useAvatarSource = (avatar?: string, options: UseAvatarSourceOptions = {}) => {
  const { fallbackSeed, id, name } = options
  const normalizedAvatar = useMemo(() => normalizeAvatarValue(avatar), [avatar])
  const fallbackSrc = useMemo(
    () => buildFallbackAvatarUrl({ fallbackSeed, id, name }),
    [fallbackSeed, id, name]
  )
  const [state, setState] = useState<AvatarSourceState>(() =>
    buildAvatarSourceState(normalizedAvatar)
  )

  useEffect(() => {
    let cancelled = false
    const initialState = buildAvatarSourceState(normalizedAvatar)
    setState(initialState)

    if (initialState.status !== 'loading') {
      return () => {
        cancelled = true
      }
    }

    resolveAvatarUrl(normalizedAvatar)
      .then((nextSrc) => {
        if (!cancelled) {
          setState({
            avatar: normalizedAvatar,
            src: nextSrc,
            status: nextSrc ? 'ready' : 'failed',
          })
        }
      })
      .catch(() => {
        if (!cancelled) {
          setState({ avatar: normalizedAvatar, src: '', status: 'failed' })
        }
      })

    return () => {
      cancelled = true
    }
  }, [normalizedAvatar])

  // Props can change before the effect above runs. Derive the new avatar's
  // initial state during render so the previous user's image never leaks into
  // the next row.
  const currentState =
    state.avatar === normalizedAvatar ? state : buildAvatarSourceState(normalizedAvatar)
  const isLoading = currentState.status === 'loading'
  const src = currentState.status === 'ready' ? currentState.src : isLoading ? '' : fallbackSrc

  return { src, fallbackSrc, isLoading }
}
