import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Post } from '../types'
import { normalizePost } from '../utils/post'
import { apiGet, apiPost } from '../utils/api'

export default function EditPost() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')
  const [tagError, setTagError] = useState<string | null>(null)
  const [isComposing, setIsComposing] = useState(false)
  const [isLoading, setIsLoading] = useState(true)

  // 获取帖子详情
  useEffect(() => {
    const fetchPost = async () => {
      if (!id) {
        navigate('/')
        return
      }

      try {
        const { data } = await apiGet<Post>(`/posts/${id}`)
        if (!data) {
          throw new Error('获取帖子详情失败')
        }
        const postData: Post = normalizePost(data)
        setTitle(postData.title)
        setContent(postData.content)
        setTags(postData.tags ?? [])
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

  const handleAddTag = () => {
    const value = tagInput.trim()

    if (!value) {
      setTagError('标签内容不能为空')
      return
    }

    if (value.length > 6) {
      setTagError('每个标签不超过 6 个字')
      return
    }

    if (tags.length >= 4) {
      setTagError('最多添加 4 个标签')
      return
    }

    if (tags.includes(value)) {
      setTagError('请不要重复添加标签')
      return
    }

    setTags(prev => [...prev, value])
    setTagInput('')
    setTagError(null)
  }

  const handleRemoveTag = (tag: string) => {
    setTags(prev => prev.filter(t => t !== tag))
    setTagError(null)
  }

  const handleTagKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !isComposing && !(e.nativeEvent as any)?.isComposing) {
      e.preventDefault()
      handleAddTag()
    }
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    
    if (!id) {
      alert('帖子ID无效')
      return
    }

    const normalizedTags = tags.map(tag => tag.trim()).filter(Boolean)
    if (normalizedTags.length > 4) {
      alert('最多添加 4 个标签')
      return
    }
    if (normalizedTags.some(tag => tag.length > 6)) {
      alert('每个标签不超过 6 个字')
      return
    }

    try {
      await apiPost('/posts/update', { id, title, content, tags: normalizedTags })

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

        {/* 标签 */}
        <div>
          <label htmlFor="tags" className="block text-sm font-medium text-gray-700 mb-2">
            标签
          </label>
          <div className="flex flex-col space-y-3">
            <div className="flex space-x-3">
              <input
                id="tags"
                type="text"
                value={tagInput}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                  setTagInput(e.target.value)
                  if (tagError) {
                    setTagError(null)
                  }
                }}
                onCompositionStart={() => setIsComposing(true)}
                onCompositionEnd={() => setIsComposing(false)}
                onKeyDown={handleTagKeyDown}
                placeholder="输入标签，按回车或点击添加"
                maxLength={6}
                className="input"
              />
              <button
                type="button"
                onClick={handleAddTag}
                className="btn-secondary whitespace-nowrap"
                disabled={tags.length >= 4}
              >
                添加标签
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {tags.map(tag => (
                <span
                  key={tag}
                  className="inline-flex items-center gap-1 px-3 py-1 rounded-full border border-primary-100 bg-primary-50/70 text-primary-700 text-sm font-medium shadow-sm"
                >
                  {tag}
                  <button
                    type="button"
                    onClick={() => handleRemoveTag(tag)}
                    className="ml-1 text-primary-500 hover:text-primary-700"
                    aria-label={`移除标签 ${tag}`}
                  >
                    &times;
                  </button>
                </span>
              ))}
              {tags.length === 0 && (
                <span className="text-sm text-gray-500">最多 4 个标签，每个不超过 6 个字</span>
              )}
            </div>
            {tagError && (
              <p className="text-sm text-red-500">{tagError}</p>
            )}
          </div>
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
