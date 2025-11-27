import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { ApiResponse, Post } from '../types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export default function EditPost() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [isLoading, setIsLoading] = useState(true)

  // 获取帖子详情
  useEffect(() => {
    const fetchPost = async () => {
      if (!id) {
        navigate('/')
        return
      }

      try {
        const response = await fetch(`${API_BASE_URL}/posts/${id}`, {
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
          throw new Error(result.msg || '获取帖子详情失败')
        }

        const postData: Post = result.data
        setTitle(postData.title)
        setContent(postData.content)
      } catch (error) {
        console.error('获取帖子详情失败:', error)
        alert('获取帖子详情失败: ' + (error instanceof Error ? error.message : '未知错误'))
        navigate('/')
      } finally {
        setIsLoading(false)
      }
    }

    fetchPost()
  }, [id, navigate])

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    
    if (!id) {
      alert('帖子ID无效')
      return
    }

    try {
      const response = await fetch(`${API_BASE_URL}/posts/update`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ id: Number(id), title, content }),
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
        throw new Error(result.msg || '更新帖子失败')
      }

      alert('帖子修改成功')
      navigate(`/post/${id}`)
    } catch (error) {
      console.error('更新帖子失败:', error)
      alert('更新帖子失败: ' + (error instanceof Error ? error.message : '未知错误'))
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

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* 返回按钮 */}
      <button
        onClick={() => navigate(-1)}
        className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
      >
        <ArrowLeft className="h-5 w-5" />
        <span>返回</span>
      </button>

      {/* 编辑帖子表单 */}
      <form onSubmit={handleSubmit} className="card space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">编辑帖子</h1>

        {/* 标题 */}
        <div>
          <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
            标题
          </label>
          <input
            id="title"
            type="text"
            value={title}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setTitle(e.target.value)}
            placeholder="输入帖子标题..."
            required
            className="input"
          />
        </div>

        {/* 内容 */}
        <div>
          <label htmlFor="content" className="block text-sm font-medium text-gray-700 mb-2">
            内容
          </label>
          <textarea
            id="content"
            value={content}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setContent(e.target.value)}
            placeholder="输入帖子内容..."
            rows={12}
            required
            className="textarea"
          />
          <p className="mt-2 text-sm text-gray-500">
            支持 Markdown 格式
          </p>
        </div>

        {/* 提交按钮 */}
        <div className="flex justify-end space-x-4 pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={() => navigate(-1)}
            className="btn-secondary"
          >
            取消
          </button>
          <button
            type="submit"
            className="btn-primary"
          >
            更新帖子
          </button>
        </div>
      </form>
    </div>
  )
}