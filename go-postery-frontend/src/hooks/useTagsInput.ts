import { useCallback, useMemo, useState } from 'react'
import type { TagsConstraints } from '../utils/tags'
import { normalizeTags, validateTags } from '../utils/tags'

export type UseTagsInputOptions = TagsConstraints & {
  initialTags?: string[]
}

export type UseTagsInputResult = {
  tags: string[]
  setTags: (tags: string[]) => void
  tagInput: string
  tagError: string | null
  isComposing: boolean
  handleTagInputChange: (value: string) => void
  handleCompositionStart: () => void
  handleCompositionEnd: () => void
  handleTagKeyDown: (event: React.KeyboardEvent<HTMLInputElement>) => void
  addTag: () => boolean
  removeTag: (tag: string) => void
  validateForSubmit: () => ReturnType<typeof validateTags>
}

export function useTagsInput(options: UseTagsInputOptions = {}): UseTagsInputResult {
  const { initialTags = [], maxTags = 4, maxTagLength = 6 } = options
  const constraints = useMemo(() => ({ maxTags, maxTagLength }), [maxTags, maxTagLength])

  const [tags, setTags] = useState<string[]>(() => normalizeTags(initialTags))
  const [tagInput, setTagInput] = useState('')
  const [tagError, setTagError] = useState<string | null>(null)
  const [isComposing, setIsComposing] = useState(false)

  const addTag = useCallback(() => {
    const value = tagInput.trim()

    if (!value) {
      setTagError('标签内容不能为空')
      return false
    }

    if (value.length > maxTagLength) {
      setTagError(`每个标签不超过 ${maxTagLength} 个字`)
      return false
    }

    if (tags.length >= maxTags) {
      setTagError(`最多添加 ${maxTags} 个标签`)
      return false
    }

    if (tags.includes(value)) {
      setTagError('请不要重复添加标签')
      return false
    }

    setTags(prev => [...prev, value])
    setTagInput('')
    setTagError(null)
    return true
  }, [maxTagLength, maxTags, tagInput, tags])

  const removeTag = useCallback((tag: string) => {
    setTags(prev => prev.filter(t => t !== tag))
    setTagError(null)
  }, [])

  const handleTagInputChange = useCallback(
    (value: string) => {
      setTagInput(value)
      if (tagError) {
        setTagError(null)
      }
    },
    [tagError]
  )

  const handleCompositionStart = useCallback(() => setIsComposing(true), [])
  const handleCompositionEnd = useCallback(() => setIsComposing(false), [])

  const handleTagKeyDown = useCallback(
    (event: React.KeyboardEvent<HTMLInputElement>) => {
      if (event.key !== 'Enter') return

      const composing = isComposing || Boolean((event.nativeEvent as any)?.isComposing)
      if (composing) return

      event.preventDefault()
      addTag()
    },
    [addTag, isComposing]
  )

  const validateForSubmit = useCallback(() => validateTags(tags, constraints), [constraints, tags])

  return {
    tags,
    setTags,
    tagInput,
    tagError,
    isComposing,
    handleTagInputChange,
    handleCompositionStart,
    handleCompositionEnd,
    handleTagKeyDown,
    addTag,
    removeTag,
    validateForSubmit,
  }
}

