import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import PostForm from '../components/PostForm'
import type { Post } from '../types'
import { normalizePost } from '../utils/post'
import { apiGet, apiPost } from '../utils/api'
import { useTagsInput } from '../hooks/useTagsInput'

export default function EditPost() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [isLoading, setIsLoading] = useState(true)

  const tagsInput = useTagsInput()
  const { setTags } = tagsInput

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
  }, [id, navigate, setTags])

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    
    if (!id) {
      alert('帖子ID无效')
      return
    }

    const validation = tagsInput.validateForSubmit()
    if (!validation.ok) {
      alert(validation.error)
      return
    }

    try {
      await apiPost(`/posts/${encodeURIComponent(id)}`, { title, content, tags: validation.tags })

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

      <PostForm
        heading="编辑帖子"
        submitLabel="更新帖子"
        title={title}
        content={content}
        onTitleChange={setTitle}
        onContentChange={setContent}
        onSubmit={handleSubmit}
        onCancel={() => navigate(-1)}
        tagsInput={tagsInput}
        tagClassName="inline-flex items-center gap-1 px-3 py-1 rounded-full border border-primary-100 bg-primary-50/70 text-primary-700 text-sm font-medium shadow-sm"
        tagRemoveClassName="ml-1 text-primary-500 hover:text-primary-700"
      />
    </div>
  )
}
