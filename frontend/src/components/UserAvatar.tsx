import type { ImgHTMLAttributes } from 'react'
import { useAvatarSource } from '../hooks/useAvatarSource'

type UserAvatarProps = Omit<ImgHTMLAttributes<HTMLImageElement>, 'src'> & {
  avatar?: string
  name?: string
  userId?: string
  fallbackSeed?: string
}

export default function UserAvatar({
  avatar,
  name,
  userId,
  fallbackSeed,
  alt,
  onError,
  ...imgProps
}: UserAvatarProps) {
  const { src, fallbackSrc } = useAvatarSource(avatar, {
    name,
    id: userId,
    fallbackSeed,
  })

  return (
    <img
      {...imgProps}
      src={src}
      alt={alt ?? name ?? '用户头像'}
      onError={(event) => {
        if (event.currentTarget.getAttribute('src') !== fallbackSrc) {
          event.currentTarget.src = fallbackSrc
        }
        onError?.(event)
      }}
    />
  )
}
