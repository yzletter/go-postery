import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Eye, Heart, MessageSquare, Clock, Tag, ThumbsUp } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { Post, Comment } from '../types'
import { useState } from 'react'

// 模拟数据
const mockPost: Post = {
  id: '1',
  title: '欢迎来到 Go Postery 论坛！',
  content: `这是一个现代化的论坛平台，欢迎大家分享想法和讨论话题。

## 主要特性

- 🎨 现代化的用户界面设计
- ⚡ 快速响应和流畅交互
- 📱 完全响应式设计，支持移动端
- 🔍 强大的搜索和筛选功能
- 💬 实时评论和互动

## 使用指南

1. 注册账号并完善个人信息
2. 浏览感兴趣的板块和话题
3. 发布你的第一个帖子
4. 参与讨论，与其他用户互动

希望你能在这里找到志同道合的朋友，分享知识和经验！`,
  author: {
    id: '1',
    name: '管理员',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin'
  },
  createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
  views: 256,
  likes: 42,
  comments: 18,
  tags: ['公告', '欢迎'],
  category: '公告'
}

const mockComments: Comment[] = [
  {
    id: '1',
    content: '这个论坛界面真的很漂亮！期待更多功能。',
    author: {
      id: '2',
      name: '前端开发者',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=developer'
    },
    createdAt: new Date(Date.now() - 1 * 60 * 60 * 1000).toISOString(),
    likes: 12,
  },
  {
    id: '2',
    content: '感谢分享，学到了很多！',
    author: {
      id: '3',
      name: 'UI设计师',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=designer'
    },
    createdAt: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
    likes: 5,
  },
]

export default function PostDetail() {
  useParams<{ id: string }>() // 获取帖子ID（当前使用模拟数据）
  const [liked, setLiked] = useState(false)
  const [commentText, setCommentText] = useState('')
  const [comments, setComments] = useState(mockComments)

  const handleLike = () => {
    setLiked(!liked)
  }

  const handleSubmitComment = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!commentText.trim()) return

    const newComment: Comment = {
      id: Date.now().toString(),
      content: commentText,
      author: {
        id: 'current-user',
        name: '当前用户',
        avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=user'
      },
      createdAt: new Date().toISOString(),
      likes: 0,
    }

    setComments([newComment, ...comments])
    setCommentText('')
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
              {mockPost.title}
            </h1>
            {mockPost.category && (
              <span className="ml-4 px-3 py-1 bg-primary-100 text-primary-700 text-sm font-medium rounded-full">
                {mockPost.category}
              </span>
            )}
          </div>

          {/* 作者信息 */}
          <div className="flex items-center space-x-4 mb-4">
            <img
              src={mockPost.author.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${mockPost.author.id}`}
              alt={mockPost.author.name}
              className="w-10 h-10 rounded-full"
            />
            <div>
              <div className="font-medium text-gray-900">{mockPost.author.name}</div>
              <div className="text-sm text-gray-500 flex items-center space-x-1">
                <Clock className="h-3 w-3" />
                <span>
                  {formatDistanceToNow(new Date(mockPost.createdAt), {
                    addSuffix: true,
                    locale: zhCN
                  })}
                </span>
              </div>
            </div>
          </div>

          {/* 标签 */}
          {mockPost.tags && mockPost.tags.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-4">
              {mockPost.tags.map(tag => (
                <span
                  key={tag}
                  className="inline-flex items-center space-x-1 px-3 py-1 bg-gray-100 text-gray-700 text-sm rounded-full"
                >
                  <Tag className="h-3 w-3" />
                  <span>{tag}</span>
                </span>
              ))}
            </div>
          )}

          {/* 统计信息 */}
          <div className="flex items-center space-x-6 text-sm text-gray-500 pb-4 border-b border-gray-200">
            <span className="flex items-center space-x-1">
              <Eye className="h-4 w-4" />
              <span>{mockPost.views} 次浏览</span>
            </span>
            <span className="flex items-center space-x-1">
              <Heart className="h-4 w-4" />
              <span>{mockPost.likes} 个赞</span>
            </span>
            <span className="flex items-center space-x-1">
              <MessageSquare className="h-4 w-4" />
              <span>{mockPost.comments} 条评论</span>
            </span>
          </div>
        </div>

        {/* 正文内容 */}
        <div className="prose prose-gray max-w-none mb-6">
          <div className="whitespace-pre-wrap text-gray-700 leading-relaxed">
            {mockPost.content}
          </div>
        </div>

        {/* 操作按钮 */}
        <div className="flex items-center space-x-4 pt-4 border-t border-gray-200">
          <button
            onClick={handleLike}
            className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-colors ${
              liked
                ? 'bg-primary-100 text-primary-700'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            <ThumbsUp className={`h-5 w-5 ${liked ? 'fill-current' : ''}`} />
            <span>点赞</span>
          </button>
        </div>
      </article>

      {/* 评论区域 */}
      <div className="card">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">
          评论 ({comments.length})
        </h2>

        {/* 评论表单 */}
        <form onSubmit={handleSubmitComment} className="mb-6">
          <textarea
            value={commentText}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setCommentText(e.target.value)}
            placeholder="写下你的评论..."
            rows={4}
            className="textarea mb-3"
          />
          <div className="flex justify-end">
            <button type="submit" className="btn-primary">
              发表评论
            </button>
          </div>
        </form>

        {/* 评论列表 */}
        <div className="space-y-6">
          {comments.map(comment => (
            <div key={comment.id} className="flex space-x-4">
              <img
                src={comment.author.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${comment.author.id}`}
                alt={comment.author.name}
                className="w-10 h-10 rounded-full flex-shrink-0"
              />
              <div className="flex-1">
                <div className="bg-gray-50 rounded-lg p-4 mb-2">
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium text-gray-900">
                      {comment.author.name}
                    </span>
                    <span className="text-xs text-gray-500">
                      {formatDistanceToNow(new Date(comment.createdAt), {
                        addSuffix: true,
                        locale: zhCN
                      })}
                    </span>
                  </div>
                  <p className="text-gray-700">{comment.content}</p>
                </div>
                <div className="flex items-center space-x-4 text-sm text-gray-500">
                  <button className="flex items-center space-x-1 hover:text-primary-600 transition-colors">
                    <ThumbsUp className="h-4 w-4" />
                    <span>{comment.likes}</span>
                  </button>
                  <button className="hover:text-primary-600 transition-colors">
                    回复
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

