import { useState, useEffect, useRef, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { MessageSquare, Eye, Heart, Clock, Loader2 } from 'lucide-react'
import { Post } from '../types'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

// 生成模拟数据的函数
const generateMockPost = (id: string, index: number): Post => {
  const categories = ['公告', '技术讨论', '设计', '问答']
  const authors = [
    { id: '1', name: '管理员' },
    { id: '2', name: '前端开发者' },
    { id: '3', name: 'UI设计师' },
    { id: '4', name: '后端工程师' },
    { id: '5', name: '产品经理' },
    { id: '6', name: '测试工程师' },
  ]
  const author = authors[index % authors.length]

  const titles = [
    '欢迎来到 Go Postery 论坛！',
    'React 18 新特性深度解析',
    '如何设计一个优雅的用户界面？',
    'Go 语言性能优化技巧分享',
    'Vue 3 Composition API 实战指南',
    'Python 数据分析入门教程',
    'TypeScript 高级类型系统详解',
    'Node.js 微服务架构实践',
    '前端工程化最佳实践',
    '数据库设计原则与优化',
    'Docker 容器化部署指南',
    'GraphQL 与 REST API 对比',
  ]

  const contents = [
    '这是一个现代化的论坛平台，欢迎大家分享想法和讨论话题。',
    'React 18 带来了很多令人兴奋的新特性，包括并发渲染、自动批处理等。让我们一起来探讨这些新功能的使用场景和最佳实践。',
    'UI设计不仅仅是美观，更重要的是用户体验。今天我们来讨论一些设计原则和最佳实践。',
    '分享一些在 Go 语言开发中遇到的性能问题和优化方案，希望对大家有帮助。',
    'Vue 3 的 Composition API 提供了更灵活的组合式开发方式，让我们深入了解其使用场景。',
    'Python 在数据分析领域有着广泛的应用，本文将介绍常用的数据分析库和技巧。',
    'TypeScript 的类型系统非常强大，掌握高级类型可以让代码更加健壮和可维护。',
    '微服务架构是现代应用开发的重要模式，Node.js 提供了很好的支持。',
    '前端工程化是提高开发效率的关键，包括构建工具、代码规范、自动化测试等。',
    '良好的数据库设计是应用性能的基础，本文将介绍设计原则和优化技巧。',
    'Docker 让应用的部署变得简单，本文将介绍如何使用 Docker 进行容器化部署。',
    'GraphQL 和 REST 各有优势，本文将对比两者的特点和使用场景。',
  ]

  return {
    id,
    title: titles[index % titles.length],
    content: contents[index % contents.length],
    author: {
      id: author.id,
      name: author.name,
      avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${author.id}${index}`
    },
    createdAt: new Date(Date.now() - (index * 60 * 60 * 1000)).toISOString(),
    views: Math.floor(Math.random() * 1000) + 100,
    likes: Math.floor(Math.random() * 200) + 10,
    comments: Math.floor(Math.random() * 100) + 5,

  }
}

// 总数据量限制
const TOTAL_POSTS_LIMIT = 20

// 模拟 API 获取帖子列表
const fetchPosts = async (page: number, pageSize: number = 10): Promise<Post[]> => {
  // 模拟网络延迟
  await new Promise(resolve => setTimeout(resolve, 500))
  
  const startIndex = (page - 1) * pageSize
  const posts: Post[] = []
  
  // 限制总数据量为 20 条
  const remainingPosts = Math.max(0, TOTAL_POSTS_LIMIT - startIndex)
  const currentPageSize = Math.min(pageSize, remainingPosts)
  
  for (let i = 0; i < currentPageSize; i++) {
    const index = startIndex + i
    posts.push(generateMockPost(`${page}-${i + 1}`, index))
  }
  
  return posts
}

export default function Home() {
  const [searchParams, setSearchParams] = useSearchParams()

  const [searchQuery, setSearchQuery] = useState<string>('')
  const [posts, setPosts] = useState<Post[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [isLoading, setIsLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [isInitialLoading, setIsInitialLoading] = useState(true)
  const observerTarget = useRef<HTMLDivElement>(null)

  const pageSize = 10

  // 从 URL 参数中读取搜索关键词
  useEffect(() => {
    const search = searchParams.get('search')
    setSearchQuery(search || '')
  }, [searchParams])

  // 加载帖子数据
  const loadPosts = useCallback(async (page: number, reset: boolean = false) => {
    if (isLoading) return
    
    setIsLoading(true)
    try {
      const newPosts = await fetchPosts(page, pageSize)
      
      if (reset) {
        setPosts(newPosts)
      } else {
        setPosts(prev => [...prev, ...newPosts])
      }
      
      // 模拟：如果返回的数据少于 pageSize，说明没有更多数据了
      setHasMore(newPosts.length === pageSize)
      setCurrentPage(page)
    } catch (error) {
      console.error('Failed to load posts:', error)
    } finally {
      setIsLoading(false)
      setIsInitialLoading(false)
    }
  }, [isLoading])

  // 当搜索改变时，重置并重新加载
  useEffect(() => {
    setIsInitialLoading(true)
    setCurrentPage(1)
    setHasMore(true)
    setPosts([])
    loadPosts(1, true)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchQuery])

  // 无限滚动：监听滚动到底部
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !isLoading && !isInitialLoading) {
          loadPosts(currentPage + 1, false)
        }
      },
      { threshold: 0.1 }
    )

    const currentTarget = observerTarget.current
    if (currentTarget) {
      observer.observe(currentTarget)
    }

    return () => {
      if (currentTarget) {
        observer.unobserve(currentTarget)
      }
    }
  }, [hasMore, isLoading, isInitialLoading, currentPage, loadPosts])

  // 搜索筛选（只搜索标题、内容、作者）
  const filteredPosts = posts.filter(post => {
    const searchMatch = !searchQuery || 
      post.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      post.content.toLowerCase().includes(searchQuery.toLowerCase()) ||
      post.author.name.toLowerCase().includes(searchQuery.toLowerCase())
    
    return searchMatch
  })

  return (
    <div className="space-y-6">
      {/* 搜索结果显示 */}
      {searchQuery && (
        <div className="card bg-primary-50 border-primary-200">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">
                搜索关键词: <span className="font-semibold text-primary-700">"{searchQuery}"</span>
              </p>
              <p className="text-sm text-gray-500 mt-1">
                找到 {filteredPosts.length} 个结果
              </p>
            </div>
            <button
              onClick={() => {
                setSearchQuery('')
                setSearchParams({}, { replace: true })
              }}
              className="text-sm text-primary-600 hover:text-primary-700 font-medium"
            >
              清除搜索
            </button>
          </div>
        </div>
      )}



      {/* 初始加载状态 */}
      {isInitialLoading && (
        <div className="card text-center py-12">
          <Loader2 className="h-8 w-8 text-primary-600 animate-spin mx-auto mb-4" />
          <p className="text-gray-500">加载中...</p>
        </div>
      )}

      {/* 帖子列表 */}
      {!isInitialLoading && (
        <>
          <div className="space-y-4">
            {filteredPosts.map(post => (
              <Link
                key={post.id}
                to={`/post/${post.id}`}
                className="card block hover:shadow-lg transition-all"
              >
                <div className="flex items-start space-x-4">
                  {/* 用户头像 */}
                  <img
                    src={post.author.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${post.author.id}`}
                    alt={post.author.name}
                    className="w-12 h-12 rounded-full flex-shrink-0"
                  />
                  
                  <div className="flex-1 min-w-0">
                    {/* 标题 */}
                    <div className="flex items-start justify-between mb-2">
                      <h2 className="text-xl font-semibold text-gray-900 hover:text-primary-600 transition-colors line-clamp-2">
                        {post.title}
                      </h2>
                    </div>

                    {/* 内容预览 */}
                    <p className="text-gray-600 mb-3 line-clamp-2">
                      {post.content}
                    </p>



                    {/* 元信息 */}
                    <div className="flex items-center justify-between text-sm text-gray-500">
                      <div className="flex items-center space-x-4">
                        <span className="font-medium text-gray-700">{post.author.name}</span>
                        <span className="flex items-center space-x-1">
                          <Clock className="h-4 w-4" />
                          <span>
                            {formatDistanceToNow(new Date(post.createdAt), {
                              addSuffix: true,
                              locale: zhCN
                            })}
                          </span>
                        </span>
                      </div>
                      
                      <div className="flex items-center space-x-4">
                        <span className="flex items-center space-x-1">
                          <Eye className="h-4 w-4" />
                          <span>{post.views}</span>
                        </span>
                        <span className="flex items-center space-x-1">
                          <Heart className="h-4 w-4" />
                          <span>{post.likes}</span>
                        </span>
                        <span className="flex items-center space-x-1">
                          <MessageSquare className="h-4 w-4" />
                          <span>{post.comments}</span>
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>

          {/* 无限滚动触发点 */}
          <div ref={observerTarget} className="h-10" />

          {/* 加载更多指示器 */}
          {isLoading && !isInitialLoading && (
            <div className="flex justify-center items-center py-8">
              <Loader2 className="h-6 w-6 text-primary-600 animate-spin mr-2" />
              <span className="text-gray-600">加载更多...</span>
            </div>
          )}

          {/* 已经到底了提示 */}
          {!hasMore && !isInitialLoading && filteredPosts.length > 0 && (
            <div className="card text-center py-8 bg-gray-50 border-dashed border-2 border-gray-200">
              <div className="flex flex-col items-center space-y-2">
                <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center">
                  <MessageSquare className="h-6 w-6 text-gray-400" />
                </div>
                <p className="text-gray-600 font-medium">已经到底了</p>
                <p className="text-sm text-gray-500">没有更多帖子可以加载</p>
              </div>
            </div>
          )}

          {/* 空状态 */}
          {filteredPosts.length === 0 && !isLoading && (
            <div className="card text-center py-12">
              <MessageSquare className="h-16 w-16 text-gray-300 mx-auto mb-4" />
              <p className="text-gray-500 text-lg">暂无帖子</p>
            </div>
          )}
        </>
      )}
    </div>
  )
}

