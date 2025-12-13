import type { Comment } from '../../types'
import { normalizeId } from '../../utils/id'

export type CommentGroup = {
  parent: Comment
  replies: Comment[]
}

export const groupComments = (comments: Comment[]): CommentGroup[] => {
  const idSet = new Set(comments.map(c => normalizeId(c.id)))
  const repliesMap = new Map<string, Comment[]>()
  const parents: Comment[] = []

  comments.forEach((comment) => {
    const parentIdStr = comment.parentId ? normalizeId(comment.parentId) : '0'
    const isParent = parentIdStr === '0' || !idSet.has(parentIdStr)
    if (isParent) {
      parents.push(comment)
      return
    }

    const bucket = repliesMap.get(parentIdStr) ?? []
    bucket.push(comment)
    repliesMap.set(parentIdStr, bucket)
  })

  return parents.map(parent => ({
    parent,
    replies: repliesMap.get(normalizeId(parent.id)) ?? [],
  }))
}

export const buildCommentAuthorMap = (comments: Comment[]) => {
  const map = new Map<string, { id: string; name: string }>()
  comments.forEach(comment => {
    map.set(normalizeId(comment.id), {
      id: normalizeId(comment.author.id),
      name: comment.author.name,
    })
  })
  return map
}

