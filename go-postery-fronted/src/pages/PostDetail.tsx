import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Clock } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { Post } from '../types'

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


}

// 模拟评论数据已移除

export default function PostDetail() {
  useParams<{ id: string }>() // 获取帖子ID（当前使用模拟数据）

  // 评论功能已移除

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




        </div>

        {/* 正文内容 */}
        <div className="prose prose-gray max-w-none mb-6">
          <div className="whitespace-pre-wrap text-gray-700 leading-relaxed">
            {mockPost.content}
          </div>
        </div>


      </article>

      {/* 评论功能已移除 */}
    </div>
  )
}

