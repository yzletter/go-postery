import { useCallback, useMemo, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { ExternalLink, RefreshCw, Search, Trash2 } from 'lucide-react'
import type { Comment } from '../../types'
import { apiDelete, apiGet } from '../../utils/api'
import { normalizeComment } from '../../utils/comment'
import { formatRelativeTime } from '../../utils/date'
import { normalizeId } from '../../utils/id'
import { groupComments } from '../postDetail/commentModel'

function CommentItem({
  comment,
  onDelete,
  disabled,
  isDeleting,
}: {
  comment: Comment
  onDelete: (commentId: string) => void
  disabled: boolean
  isDeleting: boolean
}) {
  const id = normalizeId(comment.id)

  return (
    <div className="flex items-start justify-between gap-3">
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
          <span className="font-semibold text-gray-900">{comment.author?.name || '匿名用户'}</span>
          <span className="text-xs text-gray-500">#{id}</span>
          <span className="text-xs text-gray-500">{formatRelativeTime(comment.createdAt)}</span>
        </div>
        <div className="mt-1 text-sm text-gray-700 whitespace-pre-wrap break-words">
          {comment.content || '（空）'}
        </div>
      </div>

      <button
        type="button"
        onClick={() => onDelete(id)}
        disabled={disabled}
        className="inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold text-red-700 bg-red-50 hover:bg-red-100 border border-red-100 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex-shrink-0"
      >
        <Trash2 className={`h-4 w-4 ${isDeleting ? 'animate-pulse' : ''}`} />
        <span className="hidden sm:inline">{isDeleting ? '删除中' : '删除'}</span>
      </button>
    </div>
  )
}

export default function AdminComments() {
  const [postIdInput, setPostIdInput] = useState('')
  const [postId, setPostId] = useState<string>('')
  const [comments, setComments] = useState<Comment[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const loadComments = useCallback(async (targetPostId: string) => {
    const normalizedPostId = normalizeId(targetPostId)
    if (!normalizedPostId) return

    setIsLoading(true)
    setError(null)

    try {
      const { data } = await apiGet<{
        comments: any[]
        total?: number
        hasMore?: boolean
      }>(`/posts/${encodeURIComponent(normalizedPostId)}/comments?pageNo=1&pageSize=20`)

      const parentRawList = Array.isArray(data?.comments) ? data.comments : []
      const parents = parentRawList.map((item: any) => normalizeComment(item))

      const replyResults = await Promise.allSettled(
        parents.map(async (parent) => {
          const parentId = normalizeId(parent.id)
          if (!parentId) return [] as Comment[]
          const { data: repliesData } = await apiGet<{
            comments: any[]
            total?: number
            hasMore?: boolean
          }>(
            `/posts/${encodeURIComponent(normalizedPostId)}/comments/${encodeURIComponent(parentId)}?pageNo=1&pageSize=20`
          )
          const rawReplies = Array.isArray(repliesData?.comments) ? repliesData.comments : []
          return rawReplies.map((reply: any) => normalizeComment(reply))
        })
      )
      const replies = replyResults.flatMap((result) => (result.status === 'fulfilled' ? result.value : []))
      const merged = [...parents, ...replies]
      const seen = new Set<string>()
      const deduped = merged.filter((item) => {
        const cid = normalizeId(item.id)
        if (!cid) return false
        if (seen.has(cid)) return false
        seen.add(cid)
        return true
      })

      setComments(deduped)
      setPostId(normalizedPostId)
    } catch (err) {
      const message = err instanceof Error ? err.message : '加载评论失败'
      setError(message)
      setComments([])
      setPostId(normalizedPostId)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    const normalized = normalizeId(postIdInput)
    if (!normalized) return
    void loadComments(normalized)
  }

  const handleDelete = async (commentId: string) => {
    const normalizedPostId = normalizeId(postId)
    const normalizedCommentId = normalizeId(commentId)
    if (!normalizedPostId || !normalizedCommentId) return

    if (!window.confirm(`确认删除评论 #${normalizedCommentId} 吗？此操作不可撤销。`)) {
      return
    }

    setDeletingId(normalizedCommentId)
    try {
      await apiDelete(
        `/posts/${encodeURIComponent(normalizedPostId)}/comments/${encodeURIComponent(normalizedCommentId)}`
      )
      await loadComments(normalizedPostId)
    } catch (err) {
      const message = err instanceof Error ? err.message : '删除失败'
      alert(message)
    } finally {
      setDeletingId(null)
    }
  }

  const groups = useMemo(() => groupComments(comments), [comments])

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-gray-900">评论管理</h2>
          <p className="text-sm text-gray-500">按帖子 ID 加载评论列表（后端暂无全量评论接口时使用此方式）</p>
        </div>

        <div className="flex items-center gap-2">
          {postId && (
            <Link
              to={`/post/${encodeURIComponent(postId)}`}
              className="btn-secondary !py-2"
              title="打开帖子"
            >
              <ExternalLink className="h-4 w-4" />
              打开帖子
            </Link>
          )}
          <button
            type="button"
            onClick={() => postId && void loadComments(postId)}
            disabled={!postId || isLoading}
            className="btn-secondary !py-2"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            刷新
          </button>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="h-4 w-4 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            value={postIdInput}
            onChange={(e) => setPostIdInput(e.target.value)}
            placeholder="输入帖子 ID（如 123）"
            className="input pl-9 h-11"
          />
        </div>
        <button type="submit" className="btn-primary h-11 px-6">
          加载评论
        </button>
      </form>

      {error && (
        <div className="p-3 rounded-xl border border-red-200 bg-red-50 text-red-700 text-sm">
          {error}
        </div>
      )}

      {!postId ? (
        <div className="p-6 rounded-2xl border border-gray-100 bg-white/60 text-gray-600 text-sm">
          请输入帖子 ID 后加载评论。
        </div>
      ) : isLoading ? (
        <div className="p-6 rounded-2xl border border-gray-100 bg-white/60 text-gray-600 text-sm">
          加载中...
        </div>
      ) : groups.length === 0 ? (
        <div className="p-6 rounded-2xl border border-gray-100 bg-white/60 text-gray-600 text-sm">
          该帖子暂无评论。
        </div>
      ) : (
        <div className="space-y-3">
          <div className="text-sm text-gray-500">共 {comments.length} 条评论</div>

          {groups.map(group => {
            const parentId = normalizeId(group.parent.id)
            return (
              <div
                key={parentId}
                className="border border-gray-100 rounded-xl bg-white/70 backdrop-blur-sm p-4 space-y-3"
              >
                <CommentItem
                  comment={group.parent}
                  onDelete={handleDelete}
                  disabled={Boolean(deletingId)}
                  isDeleting={deletingId === parentId}
                />

                {group.replies.length > 0 && (
                  <div className="pl-4 border-l border-gray-200 space-y-3">
                    {group.replies.map(reply => {
                      const replyId = normalizeId(reply.id)
                      return (
                        <div key={replyId} className="bg-gray-50/70 rounded-xl p-3">
                          <CommentItem
                            comment={reply}
                            onDelete={handleDelete}
                            disabled={Boolean(deletingId)}
                            isDeleting={deletingId === replyId}
                          />
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
