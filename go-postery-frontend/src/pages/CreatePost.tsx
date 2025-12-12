import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { normalizePost } from '../utils/post'
import { apiPost } from '../utils/api'

export default function CreatePost() {
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')
  const [tagError, setTagError] = useState<string | null>(null)

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
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddTag()
    }
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

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
      const { data } = await apiPost('/posts/new', { title, content, tags: normalizedTags })
      const createdPost = normalizePost(data || {})
      if (createdPost.id) {
        console.log('帖子创建成功，帖子ID:', createdPost.id)
        navigate(`/post/${createdPost.id}`)
      } else {
        navigate('/')
      }
      
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
                  className="inline-flex items-center px-3 py-1 rounded-full bg-primary-50 text-primary-700 text-sm border border-primary-100"
                >
                  #{tag}
                  <button
                    type="button"
                    onClick={() => handleRemoveTag(tag)}
                    className="ml-2 text-primary-500 hover:text-primary-700"
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
            发布帖子
          </button>
        </div>
      </form>
    </div>
  )
}
