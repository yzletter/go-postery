export const toOptionalNumber = (value: unknown): number | undefined => {
  if (value === null || value === undefined) return undefined

  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : undefined
  }

  if (typeof value === 'bigint') {
    const coerced = Number(value)
    return Number.isFinite(coerced) ? coerced : undefined
  }

  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (!trimmed) return undefined
    const coerced = Number(trimmed)
    return Number.isFinite(coerced) ? coerced : undefined
  }

  return undefined
}

