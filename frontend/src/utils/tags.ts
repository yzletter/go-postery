export type TagsConstraints = {
  maxTags?: number
  maxTagLength?: number
}

export type TagValidationResult =
  | { ok: true; tags: string[] }
  | { ok: false; tags: string[]; error: string }

export const normalizeTags = (rawTags: string[]): string[] => {
  const seen = new Set<string>()
  const tags: string[] = []

  rawTags.forEach((raw) => {
    const tag = raw.trim()
    if (!tag) return
    if (seen.has(tag)) return
    seen.add(tag)
    tags.push(tag)
  })

  return tags
}

export const validateTags = (rawTags: string[], constraints: TagsConstraints = {}): TagValidationResult => {
  const { maxTags = 4, maxTagLength = 6 } = constraints
  const tags = normalizeTags(rawTags)

  if (tags.length > maxTags) {
    return { ok: false, tags, error: `最多添加 ${maxTags} 个标签` }
  }

  const tooLong = tags.find(tag => tag.length > maxTagLength)
  if (tooLong) {
    return { ok: false, tags, error: `每个标签不超过 ${maxTagLength} 个字` }
  }

  return { ok: true, tags }
}

