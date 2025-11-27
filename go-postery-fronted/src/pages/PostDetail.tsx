import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Clock } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useState, useEffect } from 'react'
import { Post, ApiResponse } from '../types'

// 模拟数据
const mockPost: Post = {
  id: '1',
  title: '欢迎来到 Go Postery 论坛！',
  content: `这是一个现代化的论坛平台，欢迎大家分享想法和讨论话题。

## 主要特性

- 🎨 现代化的用户界面设计
- ⚡ 快速响应和流畅交互
- 📱 完全响应式设计，支持移动端
- 💬 实时评论和互动

## 使用指南

1. 注册账号并完善个人信息
2. 浏览感兴趣的板块和话题
3. 发布你的第一个帖子
4. 参与讨论，与其他用户互动

希望你能在这里找到志同道合的朋友，分享知识和经验！`,
  author: {
    id: '1',
    name: '管理员'
  },
  createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

// 模拟评论数据已移除

export default function PostDetail() {
  const { id } = useParams<{ id: string }>() // 获取帖子ID
  const [post, setPost] = useState<Post | null>(null)
  const [isLoading, setIsLoading] = useState(true)

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
      <article className="card">
        {/* 标题和元信息 */}
        <div className="mb-6">
          <div className="flex items-start justify-between mb-4">
            <h1 className="text-3xl font-bold text-gray-900 flex-1">
              {post.title}
            </h1>
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
        <div className="prose prose-gray max-w-none mb-6">
          <div className="whitespace-pre-wrap text-gray-700 leading-relaxed">
            {post.content}
          </div>
        </div>


      </article>

      {/* 评论功能已移除 */}
    </div>
  )
}

