import { useParams, Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Clock, Edit, Trash2, Heart } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useState, useEffect, FormEvent, useMemo, useCallback } from 'react'
import type { Post, Comment } from '../types'
import { normalizePost } from '../utils/post'
import { normalizeComment } from '../utils/comment'
import { normalizeId } from '../utils/id'
import { useAuth } from '../contexts/AuthContext'
import { apiDelete, apiGet, apiPost } from '../utils/api'
import { buildCommentAuthorMap, groupComments } from './postDetail/commentModel'

export default function PostDetail() {
  const { id } = useParams<{ id: string }>() // 获取帖子ID
  const navigate = useNavigate()
  const { user } = useAuth()
  const [post, setPost] = useState<Post | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [isAuthor, setIsAuthor] = useState(false)
  const [likeCount, setLikeCount] = useState(0)
  const [hasLiked, setHasLiked] = useState(false)
  const [isLiking, setIsLiking] = useState(false)
  const [isCheckingLike, setIsCheckingLike] = useState(false)
  const [commentText, setCommentText] = useState('')
  const [comments, setComments] = useState<Comment[]>([])
  const [commentsError, setCommentsError] = useState<string | null>(null)
  const [isCommentsLoading, setIsCommentsLoading] = useState(true)
  const [replyTarget, setReplyTarget] = useState<Comment | null>(null)

  const commentGroups = useMemo(() => groupComments(comments), [comments])
  const commentAuthorById = useMemo(() => buildCommentAuthorMap(comments), [comments])

  const fetchComments = useCallback(async () => {
    if (!id) return
    setIsCommentsLoading(true)
    setCommentsError(null)
    try {
      const postIdStr = normalizeId(id)
      const { data } = await apiGet<{
        comments: any[]
        total?: number
        hasMore?: boolean
      }>(`/posts/${encodeURIComponent(postIdStr)}/comments?pageNo=1&pageSize=20`)

      const parentRawList = Array.isArray(data?.comments) ? data.comments : []
      const parents = parentRawList.map((c: any) => normalizeComment(c))

      const replyResults = await Promise.allSettled(
        parents.map(async (parent) => {
          const parentId = normalizeId(parent.id)
          if (!parentId) return [] as Comment[]
          const { data: repliesData } = await apiGet<any[]>(
            `/posts/${encodeURIComponent(postIdStr)}/comments/${encodeURIComponent(parentId)}`
          )
          const rawReplies = Array.isArray(repliesData) ? repliesData : []
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
    } catch (error) {
      console.error('获取评论失败:', error)
      setComments([])
      setCommentsError(error instanceof Error ? error.message : '获取评论失败')
    } finally {
      setIsCommentsLoading(false)
    }
  }, [id])

  const handleSubmitComment = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!commentText.trim() || !id) return

    try {
      setCommentsError(null)
      const replyTargetId = replyTarget ? normalizeId(replyTarget.id) : '0'
      const normalizedParentId = replyTarget?.parentId ? normalizeId(replyTarget.parentId) : ''
      const parentIdToSend =
        replyTarget && normalizedParentId && normalizedParentId !== '0'
          ? normalizedParentId
          : replyTargetId || '0'
      const replyIdToSend = replyTarget ? replyTargetId || '0' : '0'
      const postIdToSend = normalizeId(id)

      const { data } = await apiPost(`/posts/${encodeURIComponent(postIdToSend)}/comments`, {
        parent_id: parentIdToSend,
        reply_id: replyIdToSend,
        content: commentText.trim(),
      })

      const newComment = normalizeComment(
        data || { parent_id: parentIdToSend, reply_id: replyIdToSend, content: commentText.trim() }
      )
      setComments(prev => [
        {
          ...newComment,
          parentId: newComment.parentId ?? parentIdToSend,
          replyId: newComment.replyId ?? replyIdToSend,
        },
        ...prev,
      ])
      setCommentText('')
      setReplyTarget(null)
    } catch (error) {
      console.error('发表评论失败:', error)
      setCommentsError(error instanceof Error ? error.message : '发表评论失败')
      alert('发表评论失败，请稍后重试')
    }
  }

  const handleDeleteComment = async (commentId: string | number) => {
    if (!id) {
      alert('帖子ID缺失，无法删除评论')
      return
    }

    const commentIdStr = normalizeId(commentId)
    const postIdStr = normalizeId(id)
    const target = comments.find(c => normalizeId(c.id) === commentIdStr)
    if (!target) return

    const isCommentOwner =
      user && target.author?.id !== undefined && normalizeId(user.id) === normalizeId(target.author.id)
    const canDelete = isAuthor || isCommentOwner

    if (!canDelete) {
      alert('只有帖子作者或评论作者可以删除该评论')
      return
    }

    if (!window.confirm('确认删除这条评论吗？')) {
      return
    }

    if (!isAuthor && !isCommentOwner) {
      alert('没有权限删除该评论')
      return
    }

    try {
      await apiDelete(`/posts/${encodeURIComponent(postIdStr)}/comments/${encodeURIComponent(commentIdStr)}`)
      setComments(prev => prev.filter(c => normalizeId(c.id) !== commentIdStr))
      await fetchComments()
    } catch (error) {
      console.error('删除评论失败:', error)
      alert('删除评论失败，请稍后重试')
    }
  }

  useEffect(() => {
    let cancelled = false

    const fetchPost = async () => {
      if (!id) return
      
      setIsLoading(true)
      setLoadError(null)
      try {
        const { data } = await apiGet<Post>(`/posts/${id}`)
        if (!data) {
          throw new Error('帖子详情响应数据格式错误')
        }
        
        if (cancelled) return

        setPost(normalizePost(data))
      } catch (error) {
        console.error('Failed to fetch post:', error)
        if (!cancelled) {
          setLoadError(error instanceof Error ? error.message : '获取帖子详情失败')
          setPost(null)
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    fetchPost()

    return () => {
      cancelled = true
    }
  }, [id])

  useEffect(() => {
    const postAuthorId = normalizeId(post?.author?.id)
    const userId = normalizeId(user?.id)
    setIsAuthor(Boolean(postAuthorId && userId && postAuthorId === userId))
  }, [post?.author?.id, user?.id])

  // 加载评论列表
  useEffect(() => {
    fetchComments()
  }, [fetchComments])

  useEffect(() => {
    if (!post?.id) {
      setLikeCount(0)
      return
    }
    setLikeCount(post.likes ?? 0)
  }, [post])

  const fetchLikeStatus = useCallback(async () => {
    if (!post?.id || !user?.id) {
      setHasLiked(false)
      return
    }

    const normalizedId = normalizeId(post.id)
    setIsCheckingLike(true)
    try {
      const { data } = await apiGet<boolean>(`/posts/${encodeURIComponent(normalizedId)}/likes`)
      setHasLiked(Boolean(data))
    } catch (error) {
      console.error('检查点赞状态失败:', error)
      setHasLiked(false)
    } finally {
      setIsCheckingLike(false)
    }
  }, [post, user?.id])

  useEffect(() => {
    void fetchLikeStatus()
  }, [fetchLikeStatus])

  const handleLikePost = async () => {
    if (!post?.id || isLiking || isCheckingLike) {
      return
    }
    if (!user) {
      alert('请先登录后再点赞')
      navigate('/login')
      return
    }

    const normalizedId = normalizeId(post.id)
    const willLike = !hasLiked
    const diff = willLike ? 1 : -1
    const optimisticCount = Math.max(0, likeCount + diff)
    const prevState = { liked: hasLiked, count: likeCount }

    setIsLiking(true)
    setHasLiked(willLike)
    setLikeCount(optimisticCount)
    setPost(prev => (prev ? { ...prev, likes: optimisticCount } : prev))

    try {
      if (willLike) {
        await apiPost(`/posts/${encodeURIComponent(normalizedId)}/likes`, null)
      } else {
        await apiDelete(`/posts/${encodeURIComponent(normalizedId)}/likes`)
      }
    } catch (error) {
      console.error('点赞操作失败:', error)
      setHasLiked(prevState.liked)
      setLikeCount(prevState.count)
      setPost(prev => (prev ? { ...prev, likes: prevState.count } : prev))
      alert(willLike ? '点赞失败，请稍后重试' : '取消点赞失败，请稍后重试')
    } finally {
      setIsLiking(false)
    }
  }

  // 删除帖子功能
  const handleDeletePost = async () => {
    // 确保id存在且用户确认删除操作
    if (!id || !window.confirm('确定要删除这篇帖子吗？此操作不可撤销。')) {
      return
    }
    
    try {
      await apiDelete(`/posts/${encodeURIComponent(id)}`)
      alert('帖子删除成功')
      navigate('/')
    } catch (error) {
      // 处理错误情况
      console.error('删除帖子失败:', error)
      alert('删除帖子失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto space-y-6">
        <div className="card text-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600 mx-auto mb-4"></div>
          <p className="text-gray-500">加载中...</p>
        </div>
      </div>
    )
  }

  if (!post) {
    return (
      <div className="max-w-4xl mx-auto space-y-6">
        <div className="card text-center py-12">
          <p className="text-gray-500">{loadError || '帖子不存在或加载失败'}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* 返回按钮 */}
      <Link
        to="/"
        className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
      >
        <ArrowLeft className="h-5 w-5" />
        <span>返回首页</span>
      </Link>

      {/* 帖子内容 */}
      <article className="card p-6 sm:p-8">
        {/* 标题和元信息 */}
        <div className="mb-6">
          <div className="flex items-start justify-between mb-4">
            <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 flex-1">
              {post.title}
            </h1>
            {/* 操作按钮 */}
            {isAuthor && (
              <div className="flex flex-col sm:flex-row gap-2 ml-4">
                <Link
                  to={`/edit/${post.id}`}
                  className="btn-secondary !px-3 !py-2"
                >
                  <Edit className="h-4 w-4" />
                  编辑
                </Link>
                <button
                  onClick={handleDeletePost}
                  className="btn-danger !px-3 !py-2"
                >
                  <Trash2 className="h-4 w-4" />
                  删除
                </button>
              </div>
            )}
          </div>

          {/* 作者信息 */}
          <div className="flex items-center space-x-4 mb-4">
            <Link
              to={`/users/${post.author.id}`}
              state={{ username: post.author.name }}
              className="flex-shrink-0"
            >
              <img
                src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${post.author.id}`}
                alt={post.author.name}
                className="w-10 h-10 rounded-full"
              />
            </Link>
            <div>
              <Link
                to={`/users/${post.author.id}`}
                state={{ username: post.author.name }}
                className="font-medium text-gray-900 hover:text-primary-600 transition-colors"
              >
                {post.author.name}
              </Link>
              <div className="text-sm text-gray-500 flex items-center space-x-1">
                <Clock className="h-3 w-3" />
                <span>
                  {formatDistanceToNow(new Date(post.createdAt), {
                    addSuffix: true,
                    locale: zhCN
                  })}
                </span>
              </div>
            </div>
          </div>
        </div>

        {post.tags && post.tags.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-6 px-1">
            {post.tags.map(tag => (
              <span
                key={tag}
                className="inline-flex items-center gap-1 px-3 py-1 rounded-full border border-primary-100 bg-primary-50/70 text-primary-700 text-sm font-medium shadow-sm"
              >
                {tag}
              </span>
            ))}
          </div>
        )}

        {/* 正文内容 */}
        <div className="max-w-none mb-6">
          <div className="whitespace-pre-wrap text-gray-800 leading-relaxed text-base sm:text-lg">
            {post.content}
          </div>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between border-t border-gray-200/60 pt-6 mt-8 gap-3">
          <div className="flex items-center space-x-4">
            <button
              type="button"
              onClick={handleLikePost}
              disabled={isLiking || isCheckingLike}
              className={`inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm font-semibold shadow-sm ring-1 transition-colors ${
                hasLiked
                  ? 'bg-primary-50 text-primary-700 ring-primary-200 hover:bg-primary-100'
                  : 'bg-white/70 text-gray-800 ring-gray-200/70 hover:bg-white hover:ring-gray-300'
              } disabled:opacity-70 disabled:cursor-not-allowed`}
            >
              <Heart
                className="h-4 w-4"
                fill={hasLiked ? 'currentColor' : 'none'}
              />
              <span>{hasLiked ? '取消点赞' : '点赞'}</span>
              <span className={hasLiked ? 'text-primary-700' : 'text-gray-500'}>{likeCount}</span>
            </button>
            <span className="text-sm text-gray-500">
              {likeCount > 0 ? `${likeCount} 人觉得这篇内容不错` : '成为第一个点赞的人'}
              {isLiking && !hasLiked && ' · 提交中...'}
            </span>
          </div>
          {post.views !== undefined && (
            <span className="text-sm text-gray-500">阅读 {post.views}</span>
          )}
        </div>
      </article>

      {/* 评论区域 */}
      <div className="card">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">
          评论 ({comments.length})
        </h2>

        {/* 评论表单 */}
        <form onSubmit={handleSubmitComment} className="mb-6">
          {replyTarget && (
            <div className="flex items-center justify-between text-sm text-gray-600 mb-2">
              <div className="flex items-center space-x-2">
                <span className="text-gray-500">正在回复</span>
                <span className="font-medium text-gray-900">@{replyTarget.author.name}</span>
              </div>
              <button
                type="button"
                onClick={() => setReplyTarget(null)}
                className="text-primary-600 hover:text-primary-700"
              >
                取消
              </button>
            </div>
          )}
          <textarea
            value={commentText}
            onChange={(e) => setCommentText(e.target.value)}
            placeholder="写下你的评论..."
            rows={3}
            className="textarea mb-3"
          />
          <div className="flex justify-end">
            <button type="submit" className="btn-primary">
              发表评论
            </button>
          </div>
        </form>

        {/* 评论列表 */}
        {isCommentsLoading ? (
          <div className="flex items-center text-gray-500 text-sm space-x-2">
            <div className="w-4 h-4 border-2 border-primary-600 border-t-transparent rounded-full animate-spin" />
            <span>评论加载中...</span>
          </div>
        ) : commentsError ? (
          <p className="text-sm text-red-600">{commentsError}</p>
        ) : comments.length === 0 ? (
          <p className="text-gray-500 text-sm">暂时还没有评论，快来抢沙发吧～</p>
        ) : (
          <div className="space-y-8">
            {commentGroups.map(({ parent, replies }) => (
              <div key={parent.id} className="space-y-3">
                <div className="flex space-x-4">
                  <Link
                    to={`/users/${parent.author.id}`}
                    state={{ username: parent.author.name }}
                    className="flex-shrink-0"
                  >
                    <img
                      src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${parent.author.id}`}
                      alt={parent.author.name}
                      className="w-10 h-10 rounded-full"
                    />
                  </Link>
                  <div className="flex-1">
                    <div className="bg-white/70 ring-1 ring-gray-200/60 rounded-xl p-4 mb-2">
                      <div className="flex items-center justify-between mb-2">
                        <Link
                          to={`/users/${parent.author.id}`}
                          state={{ username: parent.author.name }}
                          className="font-medium text-gray-900 hover:text-primary-600 transition-colors"
                        >
                          {parent.author.name}
                        </Link>
                        <span className="text-xs text-gray-500">
                          {formatDistanceToNow(new Date(parent.createdAt), {
                            addSuffix: true,
                            locale: zhCN
                          })}
                        </span>
                      </div>
                      <p className="text-gray-700">{parent.content}</p>
                    </div>
                    <div className="flex items-center space-x-4 text-sm text-gray-500">
                      <button
                        type="button"
                        onClick={() => setReplyTarget(parent)}
                        className="hover:text-primary-600 transition-colors"
                      >
                        回复
                      </button>
                      {(isAuthor ||
                        (user && normalizeId(user.id) === normalizeId(parent.author.id))) && (
                        <button
                          onClick={() => handleDeleteComment(parent.id)}
                          className="hover:text-red-600 transition-colors"
                        >
                          删除
                        </button>
                      )}
                    </div>
                  </div>
                </div>

                {replies.length > 0 && (
                  <div className="ml-12 border-l border-gray-100 pl-6 space-y-3">
                    {replies.map((reply) => (
                      <div key={reply.id} className="flex space-x-3">
                        <Link
                          to={`/users/${reply.author.id}`}
                          state={{ username: reply.author.name }}
                          className="flex-shrink-0"
                        >
                          <img
                            src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${reply.author.id}`}
                            alt={reply.author.name}
                            className="w-9 h-9 rounded-full"
                          />
                        </Link>
                        <div className="flex-1">
                          <div className="bg-white/70 ring-1 ring-gray-200/60 rounded-xl p-3 mb-2">
                            <div className="flex items-center justify-between mb-1.5">
                              <div className="flex items-center space-x-2">
                                <Link
                                  to={`/users/${reply.author.id}`}
                                  state={{ username: reply.author.name }}
                                  className="font-medium text-gray-900 hover:text-primary-600 transition-colors text-sm"
                                >
                                  {reply.author.name}
                                </Link>
                                {reply.replyId &&
                                  reply.parentId &&
                                  reply.replyId !== reply.parentId && (
                                    <div className="flex items-center text-xs text-gray-500 space-x-1">
                                      <span>回复</span>
                                      <Link
                                        to={`/users/${commentAuthorById.get(normalizeId(reply.replyId))?.id ?? ''}`}
                                        className="text-primary-600 hover:text-primary-700"
                                      >
                                        @{commentAuthorById.get(normalizeId(reply.replyId))?.name || '用户'}
                                      </Link>
                                    </div>
                                  )}
                              </div>
                              <span className="text-[11px] text-gray-500">
                                {formatDistanceToNow(new Date(reply.createdAt), {
                                  addSuffix: true,
                                  locale: zhCN
                                })}
                              </span>
                            </div>
                            <p className="text-gray-700 text-sm">{reply.content}</p>
                          </div>
                          <div className="flex items-center space-x-3 text-xs text-gray-500">
                            <button
                              type="button"
                              onClick={() => setReplyTarget(reply)}
                              className="hover:text-primary-600 transition-colors"
                            >
                              回复
                            </button>
                            {(isAuthor ||
                              (user && normalizeId(user.id) === normalizeId(reply.author.id))) && (
                              <button
                                onClick={() => handleDeleteComment(reply.id)}
                                className="hover:text-red-600 transition-colors"
                              >
                                删除
                              </button>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
