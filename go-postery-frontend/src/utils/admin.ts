import type { User } from '../types'
import { normalizeId } from './id'

const splitList = (value: unknown): string[] => {
  if (typeof value !== 'string') return []
  return value
    .split(/[,，\s]+/)
    .map(item => item.trim())
    .filter(Boolean)
}

const ADMIN_IDS = splitList(import.meta.env.VITE_ADMIN_IDS)
const ADMIN_NAMES = splitList(import.meta.env.VITE_ADMIN_NAMES).map(name => name.toLowerCase())
const HAS_ALLOWLIST = ADMIN_IDS.length > 0 || ADMIN_NAMES.length > 0

export const isAdminUser = (user: User | null | undefined): boolean => {
  if (!user) return false

  const id = normalizeId(user.id)
  const name = (user.name || '').trim()
  const lowerName = name.toLowerCase()

  if (!HAS_ALLOWLIST) {
    return lowerName === 'admin'
  }

  if (id && ADMIN_IDS.includes(id)) return true
  if (lowerName && ADMIN_NAMES.includes(lowerName)) return true

  return false
}

