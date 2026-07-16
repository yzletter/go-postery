import type { ImgHTMLAttributes } from 'react'
import { useAvatarSource } from '../hooks/useAvatarSource'

type UserAvatarProps = Omit<ImgHTMLAttributes<HTMLImageElement>, 'src'> & {
  avatar?: string
  name?: string
  userId?: string
  fallbackSeed?: string
}

const TRANSPARENT_PIXEL =
  'data:image/gif;base64,R0lGODlhAQABAAD/ACwAAAAAAQABAAACADs='

export default function UserAvatar({
  avatar,
  name,
  userId,
  fallbackSeed,
  alt,
  className,
  decoding,
  onError,
  ...imgProps
}: UserAvatarProps) {
  const { src, fallbackSrc, isLoading } = useAvatarSource(avatar, {
    name,
    id: userId,
    fallbackSeed,
  })

  return (
    <img
      {...imgProps}
      src={isLoading ? TRANSPARENT_PIXEL : src}
      alt={alt ?? name ?? '用户头像'}
      className={`${isLoading ? 'bg-gray-100' : ''} ${className ?? ''}`.trim()}
      decoding={decoding ?? 'async'}
      aria-busy={isLoading || undefined}
      onError={(event) => {
        if (event.currentTarget.getAttribute('src') !== fallbackSrc) {
          event.currentTarget.src = fallbackSrc
        }
        onError?.(event)
      }}
    />
  )
}
