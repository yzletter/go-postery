import { useParams, Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Clock, Edit, Trash2, Heart } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useState, useEffect, FormEvent, useMemo, useCallback } from 'react'
import { Post, Comment } from '../types'
import { normalizePost } from '../utils/post'
import { normalizeComment } from '../utils/comment'
import { normalizeId } from '../utils/id'
import { useAuth } from '../contexts/AuthContext'
import { apiGet, apiPost } from '../utils/api'

const LIKED_POSTS_KEY = 'go-postery-liked-posts'

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
  const [commentText, setCommentText] = useState('')
  const [comments, setComments] = useState<Comment[]>([])
  const [commentsError, setCommentsError] = useState<string | null>(null)
  const [isCommentsLoading, setIsCommentsLoading] = useState(true)
  const [replyTarget, setReplyTarget] = useState<Comment | null>(null)

  const commentGroups = useMemo(() => {
    const idSet = new Set(comments.map(c => normalizeId(c.id)))
    const repliesMap = new Map<string, Comment[]>()
    const parents: Comment[] = []

    comments.forEach((c) => {
      const parentIdStr = c.parentId ? normalizeId(c.parentId) : '0'
      const isParent = parentIdStr === '0' || !idSet.has(parentIdStr)
      if (isParent) {
        parents.push(c)
        return
      }
      const bucket = repliesMap.get(parentIdStr) ?? []
      bucket.push(c)
      repliesMap.set(parentIdStr, bucket)
    })

    return parents.map(parent => ({
      parent,
      replies: repliesMap.get(normalizeId(parent.id)) ?? [],
    }))
  }, [comments])

  const commentAuthorById = useMemo(() => {
    const map = new Map<string, { id: string; name: string }>()
    comments.forEach(c => {
      map.set(normalizeId(c.id), { id: normalizeId(c.author.id), name: c.author.name })
    })
    return map
  }, [comments])

  const fetchComments = useCallback(async () => {
    if (!id) return
    setIsCommentsLoading(true)
    setCommentsError(null)
    try {
      const { data } = await apiGet<Comment[] | { comments: Comment[] }>(`/comment/list/${id}`)
      const normalizedList = Array.isArray(data)
        ? data
        : Array.isArray((data as any)?.comments)
          ? (data as any).comments
          : []
      setComments(normalizedList.map((c: any) => normalizeComment(c)))
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

      const { data } = await apiPost('/comment/new', {
        post_id: postIdToSend,
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

    if (!isAuthor) {
      const belongs = await checkCommentOwnership(commentIdStr)
      if (!belongs) {
        alert('只能删除自己的评论')
        return
      }
    }

    try {
      await apiGet(`/comment/delete/${postIdStr}/${commentIdStr}`)
      setComments(prev => prev.filter(c => normalizeId(c.id) !== commentIdStr))
      await fetchComments()
    } catch (error) {
      console.error('删除评论失败:', error)
      alert('删除评论失败，请稍后重试')
    }
  }

  const checkCommentOwnership = useCallback(async (commentId: string | number): Promise<boolean> => {
    try {
      const commentIdStr = normalizeId(commentId)
      await apiGet(`/comment/belong?id=${commentIdStr}`)
      return true
    } catch (error) {
      console.error('检查评论所有权失败:', error)
      return false
    }
  }, [])

  const checkPostOwnership = useCallback(async (postId: string): Promise<boolean> => {
    try {
      const postIdStr = normalizeId(postId)
      await apiGet(`/posts/belong?id=${postIdStr}`)
      return true
    } catch (error) {
      console.error('检查帖子所有权失败:', error)
      return false
    }
  }, [])

  const readLikedPosts = useCallback((): Record<string, boolean> => {
    try {
      const raw = localStorage.getItem(LIKED_POSTS_KEY)
      const parsed = raw ? JSON.parse(raw) : {}
      if (parsed && typeof parsed === 'object') {
        return parsed as Record<string, boolean>
      }
    } catch (error) {
      console.warn('读取点赞记录失败:', error)
    }
    return {}
  }, [])

  const rememberLikedPost = useCallback((postId: string) => {
    try {
      const likedMap = readLikedPosts()
      likedMap[postId] = true
      localStorage.setItem(LIKED_POSTS_KEY, JSON.stringify(likedMap))
    } catch (error) {
      console.warn('保存点赞记录失败:', error)
    }
  }, [readLikedPosts])

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
        
        // 检查帖子所有权
        const ownership = await checkPostOwnership(id)
        if (!cancelled) {
          setIsAuthor(ownership)
        }
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
  }, [id, checkPostOwnership])

  // 加载评论列表
  useEffect(() => {
    fetchComments()
  }, [fetchComments])

  useEffect(() => {
    if (!post?.id) {
      setLikeCount(0)
      setHasLiked(false)
      return
    }
    const normalizedId = normalizeId(post.id)
    setLikeCount(post.likes ?? 0)
    setHasLiked(Boolean(readLikedPosts()[normalizedId]))
  }, [post, readLikedPosts])

  const handleLikePost = async () => {
    if (!post?.id || hasLiked || isLiking) {
      return
    }

    const normalizedId = normalizeId(post.id)
    const nextCount = likeCount + 1

    setIsLiking(true)
    setLikeCount(nextCount)
    setPost(prev => (prev ? { ...prev, likes: nextCount } : prev))

    try {
      const { data } = await apiPost('/posts/like', { id: normalizedId, post_id: normalizedId })
      const serverLikes =
        (data as any)?.likes ??
        (data as any)?.Likes ??
        (data as any)?.likeCount ??
        (data as any)?.like_count ??
        (data as any)?.LikeCount
      const parsedServerLikes = typeof serverLikes === 'number' ? serverLikes : Number(serverLikes)
      const finalCount = Number.isFinite(parsedServerLikes) ? parsedServerLikes : nextCount

      setLikeCount(finalCount)
      setPost(prev => (prev ? { ...prev, likes: finalCount } : prev))
      setHasLiked(true)
      rememberLikedPost(normalizedId)
    } catch (error) {
      console.error('点赞失败:', error)
      setLikeCount(prev => Math.max(0, prev - 1))
      setPost(prev => (prev ? { ...prev, likes: Math.max(0, (prev.likes ?? 1) - 1) } : prev))
      alert('点赞失败，请稍后重试')
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
      const { msg } = await apiGet(`/posts/delete/${id}`)
      alert(msg || '帖子删除成功')
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
    <div className="max-w-6xl mx-auto space-y-6 lg:-ml-1">
      {/* 返回按钮 */}
      <Link
        to="/"
        className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
      >
        <ArrowLeft className="h-5 w-5" />
        <span>返回首页</span>
      </Link>

      {/* 帖子内容 */}
      <article className="card px-12">
        {/* 标题和元信息 */}
        <div className="mb-6">
          <div className="flex items-start justify-between mb-4">
            <h1 className="text-3xl font-bold text-gray-900 flex-1">
              {post.title}
            </h1>
            {/* 操作按钮 */}
            {isAuthor && (
              <div className="flex space-x-2 ml-4">
                <Link
                  to={`/edit/${post.id}`}
                  className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors"
                >
                  <Edit className="h-4 w-4 mr-1" />
                  编辑
                </Link>
                <button
                  onClick={handleDeletePost}
                  className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors"
                >
                  <Trash2 className="h-4 w-4 mr-1" />
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

        {/* 正文内容 */}
        <div className="prose prose-gray max-w-none mb-6 px-4">
          <div className="whitespace-pre-wrap text-gray-700 leading-relaxed text-lg">
            {post.content}
          </div>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between border-t border-gray-100 pt-6 mt-8 space-y-3 sm:space-y-0">
          <div className="flex items-center space-x-4">
            <button
              type="button"
              onClick={handleLikePost}
              disabled={hasLiked || isLiking}
              className={`inline-flex items-center px-4 py-2 rounded-full border text-sm font-medium transition-colors ${
                hasLiked
                  ? 'bg-primary-50 border-primary-200 text-primary-700 cursor-default'
                  : 'bg-gray-50 border-gray-200 text-gray-700 hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 disabled:opacity-70 disabled:cursor-not-allowed'
              }`}
            >
              <Heart
                className="h-4 w-4 mr-2"
                fill={hasLiked ? 'currentColor' : 'none'}
              />
              <span>{hasLiked ? '已点赞' : '点赞'}</span>
              <span className="ml-2 text-gray-500">{likeCount}</span>
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
                    <div className="bg-gray-50 rounded-lg p-4 mb-2">
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
                          <div className="bg-gray-50 rounded-lg p-3 mb-2">
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
