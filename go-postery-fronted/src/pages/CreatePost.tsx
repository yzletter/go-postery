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
    
    // 暂时禁用后端调用，直接跳转到首页
    console.log('创建帖子API调用已禁用，直接跳转到首页')
    alert('帖子创建功能已禁用（仅用于用户接口测试），点击确定返回首页')
    navigate('/')
    return
    
    /* 原始的后端调用代码，暂时注释
        const response = await fetch(`${API_BASE_URL}/posts`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
          },
          body: JSON.stringify({ title, content }),
          credentials: 'include', // 关键：确保Cookie随请求发送
        })

      const result: ApiResponse = await response.json()
      
      // 根据API文档：code为0表示成功，1表示失败
      if (result.code !== 0) {
        throw new Error(result.msg || '创建帖子失败')
      }

      console.log('帖子创建成功:', result.data)
      navigate('/')
    } catch (error) {
      console.error('Failed to create post:', error)
      // 如果后端不可用，使用模拟创建（仅用于开发演示）
      if (error instanceof TypeError && error.message.includes('fetch')) {
        console.warn('后端 API 不可用，使用模拟创建帖子（仅用于开发）')
        navigate('/')
        return
      }
      // 处理响应格式错误的情况，也使用模拟创建
      if (error instanceof Error && error.message.includes('响应数据格式错误')) {
        console.warn('后端响应格式错误，使用模拟创建帖子（仅用于开发）')
        navigate('/')
        return
      }
      alert('创建帖子失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
    */
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

