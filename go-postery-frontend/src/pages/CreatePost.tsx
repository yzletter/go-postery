import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import PostForm from '../components/PostForm'
import { normalizePost } from '../utils/post'
import { apiPost } from '../utils/api'
import { useTagsInput } from '../hooks/useTagsInput'

export default function CreatePost() {
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const tagsInput = useTagsInput()

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const validation = tagsInput.validateForSubmit()
    if (!validation.ok) {
      alert(validation.error)
      return
    }
    
    try {
      const { data } = await apiPost('/posts', { title, content, tags: validation.tags })
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

      <PostForm
        heading="发布新帖子"
        submitLabel="发布帖子"
        title={title}
        content={content}
        onTitleChange={setTitle}
        onContentChange={setContent}
        onSubmit={handleSubmit}
        onCancel={() => navigate(-1)}
        tagsInput={tagsInput}
        tagPrefix="#"
      />
    </div>
  )
}
