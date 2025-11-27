import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { ApiResponse } from '../types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

export default function CreatePost() {
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    
    try {
      // 修改路由为 localhost:8080/posts/new
      const response = await fetch('http://localhost:8080/posts/new', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ title, content }),
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
        throw new Error(result.msg || '创建帖子失败')
      }

      // 后端只返回帖子ID，直接跳转到首页
      console.log('帖子创建成功，帖子ID:', result.data)
      navigate('/')
      
    } catch (error) {
      console.error('Failed to create post:', error)
      alert('创建帖子失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
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

      {/* 发帖表单 */}
      <form onSubmit={handleSubmit} className="card space-y-6">
        <h1 className="text-3xl font-bold text-gray-900">发布新帖子</h1>

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
            发布帖子
          </button>
        </div>
      </form>
    </div>
  )
}

