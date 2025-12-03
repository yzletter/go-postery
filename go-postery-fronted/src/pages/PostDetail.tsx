import { useParams, Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Clock, Edit, Trash2 } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useState, useEffect, FormEvent, useMemo } from 'react'
import { Post, ApiResponse, Comment } from '../types'
import { normalizePost } from '../utils/post'
import { normalizeComment } from '../utils/comment'
import { useAuth } from '../contexts/AuthContext'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export default function PostDetail() {
  const { id } = useParams<{ id: string }>() // 获取帖子ID
  const navigate = useNavigate()
  const { user } = useAuth()
  const [post, setPost] = useState<Post | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isAuthor, setIsAuthor] = useState(false)
  const [commentText, setCommentText] = useState('')
  const [comments, setComments] = useState<Comment[]>([])
  const [isCommentsLoading, setIsCommentsLoading] = useState(true)
  const [replyTarget, setReplyTarget] = useState<Comment | null>(null)

  const commentGroups = useMemo(() => {
    const idSet = new Set(comments.map(c => String(c.id)))
    const repliesMap = new Map<string, Comment[]>()
    const parents: Comment[] = []

    comments.forEach((c) => {
      const parentId = c.parentId ?? 0
      const parentIdStr = String(parentId)
      if (!parentId || !idSet.has(parentIdStr)) {
        parents.push(c)
        return
      }
      const bucket = repliesMap.get(parentIdStr) ?? []
      bucket.push(c)
      repliesMap.set(parentIdStr, bucket)
    })

    return parents.map(parent => ({
      parent,
      replies: repliesMap.get(String(parent.id)) ?? [],
    }))
  }, [comments])

  const handleSubmitComment = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!commentText.trim() || !id) return

    try {
      const parentIdToSend = replyTarget ? Number(replyTarget.id) : 0

      const response = await fetch(`${API_BASE_URL}/comment/new`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          post_id: Number(id),
          parent_id: parentIdToSend,
          content: commentText.trim(),
        }),
      })

      const result: ApiResponse = await response.json()
      if (!response.ok || result.code !== 0) {
        throw new Error(result.msg || '发表评论失败')
      }

      const newComment = normalizeComment(
        result.data || { parent_id: parentIdToSend, content: commentText.trim() }
      )
      setComments(prev => [
        {
          ...newComment,
          parentId: newComment.parentId ?? parentIdToSend,
        },
        ...prev,
      ])
      setCommentText('')
      setReplyTarget(null)
    } catch (error) {
      console.error('发表评论失败:', error)
      alert('发表评论失败，请稍后重试')
    }
  }

  const handleDeleteComment = async (commentId: string | number) => {
    const target = comments.find(c => String(c.id) === String(commentId))
    if (!target) return

    const isCommentOwner =
      user && target.author?.id !== undefined && String(user.id) === String(target.author.id)
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
      const belongs = await checkCommentOwnership(commentId)
      if (!belongs) {
        alert('只能删除自己的评论')
        return
      }
    }

    try {
      const response = await fetch(`${API_BASE_URL}/comment/delete/${commentId}`, {
        method: 'GET',
        credentials: 'include',
      })
      const result: ApiResponse = await response.json()
      if (!response.ok || result.code !== 0) {
        throw new Error(result.msg || '删除评论失败')
      }
      setComments(prev => prev.filter(c => String(c.id) !== String(commentId)))
    } catch (error) {
      console.error('删除评论失败:', error)
      alert('删除评论失败，请稍后重试')
    }
  }

  const checkCommentOwnership = async (commentId: string | number): Promise<boolean> => {
    try {
      const response = await fetch(`${API_BASE_URL}/comment/belong?id=${commentId}`, {
        method: 'GET',
        credentials: 'include',
      })
      const result: ApiResponse = await response.json()
      if (!response.ok || result.code !== 0) {
        return false
      }
      return true
    } catch (error) {
      console.error('检查评论所有权失败:', error)
      return false
    }
  }

  // 创建一个函数来检查帖子是否属于当前用户
  const checkPostOwnership = async (postId: string): Promise<boolean> => {
    try {
      // 使用GET请求，参数名为id而不是postId
      const response = await fetch(`${API_BASE_URL}/posts/belong?id=${postId}`, {
        method: 'GET',
        credentials: 'include',
      })
      
      const result: ApiResponse = await response.json()
      
      if (!response.ok || result.code !== 0) {
        return false
      }
      
      return true
    } catch (error) {
      console.error('检查帖子所有权失败:', error)
      return false
    }
  }

  useEffect(() => {
    const fetchPost = async () => {
      if (!id) return
      
      setIsLoading(true)
      try {
        // 启用后端调用进行接口测试
        console.log('帖子详情API调用已启用，进行接口测试')
        
        const response = await fetch(`${API_BASE_URL}/posts/${id}`, {
          credentials: 'include', // 关键：确保Cookie随请求发送
        })
        
        const result: ApiResponse = await response.json()
        
        if (!response.ok || result.code !== 0) {
          throw new Error(result.msg || '获取帖子详情失败')
        }

        const responseData = result.data
        if (!responseData) {
          throw new Error('帖子详情响应数据格式错误')
        }
        
        setPost(normalizePost(responseData))
        
        // 检查帖子所有权
        const ownership = await checkPostOwnership(id)
        setIsAuthor(ownership)
      } catch (error) {
        console.error('Failed to fetch post:', error)
        // 接口测试期间，直接抛出错误而不是回退到模拟数据
        throw error
      } finally {
        setIsLoading(false)
      }
    }

    fetchPost()
  }, [id])

  // 加载评论列表
  useEffect(() => {
    const fetchComments = async () => {
      if (!id) return
      setIsCommentsLoading(true)
      try {
        const response = await fetch(`${API_BASE_URL}/comment/list/${id}`, {
          credentials: 'include',
        })
        const result: ApiResponse = await response.json()
        if (!response.ok || result.code !== 0) {
          throw new Error(result.msg || '获取评论失败')
        }

        const data = Array.isArray(result.data)
          ? result.data
          : Array.isArray(result.data?.comments)
            ? result.data.comments
            : []
        setComments(data.map((c: any) => normalizeComment(c)))
      } catch (error) {
        console.error('获取评论失败:', error)
        setComments([])
      } finally {
        setIsCommentsLoading(false)
      }
    }

    fetchComments()
  }, [id])

  // 删除帖子功能
  const handleDeletePost = async () => {
    // 确保id存在且用户确认删除操作
    if (!id || !window.confirm('确定要删除这篇帖子吗？此操作不可撤销。')) {
      return
    }
    
    try {
      // 发送GET请求到后端API (更新路径为/posts/delete/:id)
      const response = await fetch(`${API_BASE_URL}/posts/delete/${id}`, {
        method: 'GET',
        credentials: 'include',
      })
      
      const result = await response.json()
      
      if (!response.ok || result.code !== 0) {
        throw new Error(result.msg || '删除帖子失败')
      }
      
      // 删除成功，显示成功消息并导航回主页
      alert(result.msg || '帖子删除成功')
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
          <p className="text-gray-500">帖子不存在或加载失败</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
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
                        (user && String(user.id) === String(parent.author.id))) && (
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
                              <Link
                                to={`/users/${reply.author.id}`}
                                state={{ username: reply.author.name }}
                                className="font-medium text-gray-900 hover:text-primary-600 transition-colors text-sm"
                              >
                                {reply.author.name}
                              </Link>
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
                              (user && String(user.id) === String(reply.author.id))) && (
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
