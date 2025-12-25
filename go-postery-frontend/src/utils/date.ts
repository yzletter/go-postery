import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export const formatRelativeTime = (value?: string, fallback?: string) => {
  const safeFallback = fallback ?? value ?? ''
  if (!value) return safeFallback
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return safeFallback
  return formatDistanceToNow(date, { addSuffix: true, locale: zhCN })
}
