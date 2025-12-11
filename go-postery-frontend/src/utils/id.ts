// Helpers to keep backend int64 IDs precise in the frontend.
// Always treat IDs as strings when sending to or receiving from the API.
export const normalizeId = (value: unknown): string => {
  if (value === null || value === undefined) return ''
  return typeof value === 'bigint' ? value.toString() : String(value)
}

// Build a small numeric seed for UI-only randomness (e.g. tag picking) without precision loss.
export const buildIdSeed = (value: unknown, fallback: number = 0): number => {
  const digits = normalizeId(value).replace(/\D/g, '')
  if (!digits) return fallback
  const tail = digits.slice(-9) // keep within 32-bit safe range
  const parsed = Number.parseInt(tail, 10)
  return Number.isFinite(parsed) ? parsed : fallback
}
