import { useParams, Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Clock, Edit, Trash2 } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useState, useEffect } from 'react'
import { Post, ApiResponse } from '../types'

export default function PostDetail() {
  const { id } = useParams<{ id: string }>() // 获取帖子ID
  const navigate = useNavigate()
  const [post, setPost] = useState<Post | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isAuthor, setIsAuthor] = useState(false)

  // 创建一个函数来检查帖子是否属于当前用户
  const checkPostOwnership = async (postId: string): Promise<boolean> => {
    try {
      // 使用GET请求，参数名为id而不是postId
      const response = await fetch(`http://localhost:8080/posts/belong?id=${postId}`, {
        method: 'GET',
        credentials: 'include',
      })
      
      if (!response.ok) {
        throw new Error(`HTTP错误: ${response.status}`)
      }
      
      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        throw new Error('响应不是JSON格式')
      }
      
      const result: ApiResponse = await response.json()
      
      if (result.code !== 0) {
        throw new Error(result.msg || '检查帖子所有权失败')
      }
      
      // 根据API文档，data字段应该是"true"或"false"字符串
      // 需要将其转换为布尔值
      return result.data === "true" || result.data === true
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
        
        const response = await fetch(`http://localhost:8080/posts/${id}`, {
          credentials: 'include', // 关键：确保Cookie随请求发送
        })
        
        // 检查响应状态
        if (!response.ok) {
          throw new Error(`HTTP错误: ${response.status}`)
        }
        
        // 检查内容类型
        const contentType = response.headers.get('content-type')
        if (!contentType || !contentType.includes('application/json')) {
          throw new Error('响应不是JSON格式')
        }
        
        const result: ApiResponse = await response.json()
        
        // 根据API文档：code为0表示成功，1表示失败
        if (result.code !== 0) {
          throw new Error(result.msg || '获取帖子详情失败')
        }

        // 根据API文档，帖子详情在data中
        const responseData = result.data
        if (!responseData) {
          throw new Error('帖子详情响应数据格式错误')
        }
        
        setPost(responseData)
        
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

  // 删除帖子功能
  const handleDeletePost = async () => {
    // 确保id存在且用户确认删除操作
    if (!id || !window.confirm('确定要删除这篇帖子吗？此操作不可撤销。')) {
      return
    }
    
    try {
      // 发送GET请求到后端API (更新路径为/posts/delete/:id)
      const response = await fetch(`http://localhost:8080/posts/delete/${id}`, {
        method: 'GET',
        credentials: 'include',
      })
      
      // 检查响应状态
      if (!response.ok) {
        throw new Error(`HTTP错误: ${response.status}`)
      }
      
      // 检查内容类型
      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        throw new Error('响应不是JSON格式')
      }
      
      // 解析响应数据
      const result = await response.json()
      
      // 根据API文档：code为0表示成功，1表示失败
      if (result.code !== 0) {
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

  // 评论功能已移除

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
            <img
              src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${post.author.id}`}
              alt={post.author.name}
              className="w-10 h-10 rounded-full"
            />
            <div>
              <div className="font-medium text-gray-900">{post.author.name}</div>
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
    </div>
  )
}