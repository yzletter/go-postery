import { Bot, Braces, Code2, Coffee, Goal, Grid, Pi, Server } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

export type CategoryItem = { key: string; label: string; icon: LucideIcon; tag?: string }

export const categories: CategoryItem[] = [
  { key: 'all', label: '全部', icon: Grid },
  { key: 'frontend', label: '前端', tag: '前端', icon: Code2 },
  { key: 'backend', label: '后端', tag: '后端', icon: Server },
  { key: 'go', label: 'Go', tag: 'Go', icon: Goal },
  { key: 'cpp', label: 'C++', tag: 'C++', icon: Braces },
  { key: 'java', label: 'Java', tag: 'Java', icon: Coffee },
  { key: 'python', label: 'Python', tag: 'Python', icon: Pi },
  { key: 'ai', label: 'AI', tag: 'AI', icon: Bot },
]

export const getRequestTag = (categoryKey: string): string => {
  const match = categories.find(cat => cat.key === categoryKey)
  return match?.tag ?? match?.label ?? categoryKey
}

export const FETCH_TIMEOUT_MS = 8000
export const DEFAULT_PAGE_SIZE = 10
export const CATEGORY_PAGE_SIZE = 10
