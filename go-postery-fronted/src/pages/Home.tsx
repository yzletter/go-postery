import { useState, useEffect, useRef, useCallback } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { MessageSquare, Clock, Loader2, Eye, Heart, Flame } from 'lucide-react'
import { Post, ApiResponse } from '../types'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// 生成模拟数据的函数
const generateMockPost = (id: number, index: number): Post => {
  const authors = [
    { id: 1, name: '管理员' },
    { id: 2, name: '前端开发者' },
    { id: 3, name: 'UI设计师' },
    { id: 4, name: '后端工程师' },
    { id: 5, name: '产品经理' },
    { id: 6, name: '测试工程师' },
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
      name: author.name
    },
    createdAt: new Date(Date.now() - (index * 60 * 60 * 1000)).toISOString(),
    views: Math.floor(Math.random() * 1000) + 100,
    likes: Math.floor(Math.random() * 200) + 10,
    comments: Math.floor(Math.random() * 100) + 5,
  }
}

// 总数据量限制
const TOTAL_POSTS_LIMIT = 20
// 保留模拟数据方法的引用，便于需要时启用本地数据
void generateMockPost
void TOTAL_POSTS_LIMIT

interface PostListResult {
  posts: Post[]
  total: number
  hasMore: boolean
}

const mockHotPosts = [
  { id: 1, title: 'React 18 并发特性最佳实践', heat: 985 },
  { id: 2, title: 'Go 微服务网关设计要点', heat: 912 },
  { id: 3, title: 'Tailwind 设计系统落地经验', heat: 876 },
  { id: 4, title: '前端性能优化 25 条检查清单', heat: 844 },
  { id: 5, title: '数据库索引失效的常见原因', heat: 828 },
  { id: 6, title: 'Vue3 + Vite 项目工程化模板', heat: 801 },
  { id: 7, title: 'Rust 学习路径与上手案例', heat: 776 },
  { id: 8, title: 'K8s 部署流水线实战分享', heat: 754 },
  { id: 9, title: '前后端接口约定与错误码规范', heat: 731 },
  { id: 10, title: '设计师和工程师协作的 7 个技巧', heat: 702 },
]

// API 获取帖子列表
const fetchPosts = async (page: number, pageSize: number = 10): Promise<PostListResult> => {
  try {
    // 启用后端调用进行接口测试
    console.log('帖子列表API调用已启用，进行接口测试')
    
    const response = await fetch(`${API_BASE_URL}/posts?pageNo=${page}&pageSize=${pageSize}`, {
      credentials: 'include', // 关键：确保Cookie随请求发送
    })
    
    const result: ApiResponse = await response.json()
    
    if (!response.ok || result.code !== 0) {
      throw new Error(result.msg || '获取帖子列表失败')
    }

    const responseData = result.data
    if (!responseData || !responseData.posts) {
      throw new Error('帖子列表响应数据格式错误')
    }
    
    const postsWithStats: Post[] = responseData.posts.map((p: Post, idx: number) => ({
      ...p,
      views: p.views ?? Math.floor(Math.random() * 500) + 50 + idx,
      likes: p.likes ?? Math.floor(Math.random() * 80) + 5 + idx,
      comments: p.comments ?? Math.floor(Math.random() * 40) + idx,
    }))

    return {
      posts: postsWithStats,
      total: responseData.total ?? 0,
      hasMore: Boolean(responseData.hasMore),
    }
    
    /* 模拟数据代码，暂时注释
    console.log('帖子列表API调用已禁用，使用模拟数据')
    
    // 模拟网络延迟
    await new Promise(resolve => setTimeout(resolve, 300))
    
    // 返回模拟数据
    const posts: Post[] = []
    const startIndex = (page - 1) * pageSize
    
    // 限制总数据量为 20 条
    const remainingPosts = Math.max(0, TOTAL_POSTS_LIMIT - startIndex)
    const currentPageSize = Math.min(pageSize, remainingPosts)
    
    for (let i = 0; i < currentPageSize; i++) {
      const index = startIndex + i
      posts.push(generateMockPost(`${page}-${i + 1}`, index))
    }
    
    return posts
    */
  } catch (error) {
    console.error('Failed to fetch posts:', error)
    // 接口测试期间，直接抛出错误而不是回退到模拟数据
    throw error
  }
}

export default function Home() {
  const [posts, setPosts] = useState<Post[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [isLoading, setIsLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [isInitialLoading, setIsInitialLoading] = useState(true)
  const navigate = useNavigate()
  const observerTarget = useRef<HTMLDivElement>(null)

  const pageSize = 10

  // 加载帖子数据
  const loadPosts = useCallback(async (page: number, reset: boolean = false) => {
    if (isLoading) return
    
    setIsLoading(true)
    try {
      // 如果不是初始加载（即加载更多），添加1秒延时
      if (!reset) {
        await new Promise(resolve => setTimeout(resolve, 500))
      }
      
      const { posts: newPosts, hasMore: hasMoreFromApi } = await fetchPosts(page, pageSize)
      
      if (reset) {
        setPosts(newPosts)
      } else {
        setPosts(prev => [...prev, ...newPosts])
      }
      
      setHasMore(hasMoreFromApi)
      setCurrentPage(page)
    } catch (error) {
      console.error('Failed to load posts:', error)
    } finally {
      setIsLoading(false)
      setIsInitialLoading(false)
    }
  }, [isLoading])

  // 初始加载帖子
  useEffect(() => {
    setIsInitialLoading(true)
    setCurrentPage(1)
    setHasMore(true)
    setPosts([])
    loadPosts(1, true)
  }, [])

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

  return (
    <div className="grid lg:grid-cols-[minmax(0,2fr)_minmax(240px,320px)] gap-6 items-start">
      <section className="space-y-6 lg:-ml-2 xl:-ml-4">
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
                <div className="space-y-3">
                  {posts.map(post => (
                    <article
                      key={post.id}
                      role="link"
                      tabIndex={0}
                      onClick={() => navigate(`/post/${post.id}`)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault()
                          navigate(`/post/${post.id}`)
                        }
                      }}
                      className="card p-4 lg:p-5 hover:shadow-lg transition-all cursor-pointer"
                    >
                      <div className="flex items-start space-x-4">
                        {/* 用户头像 */}
                        <Link
                          to={`/users/${post.author.id}`}
                          state={{ username: post.author.name }}
                          onClick={(e) => e.stopPropagation()}
                          className="flex-shrink-0"
                        >
                          <img
                            src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${post.author.id}`}
                            alt={post.author.name}
                            className="w-11 h-11 rounded-full"
                          />
                        </Link>
                        
                        <div className="flex-1 min-w-0">
                          {/* 标题 */}
                          <div className="flex items-start justify-between mb-1.5">
                            <h2 className="text-lg font-semibold text-gray-900 hover:text-primary-600 transition-colors line-clamp-2">
                              {post.title}
                            </h2>
                          </div>

                          {/* 内容预览 */}
                          <p className="text-gray-600 mb-2 line-clamp-2 text-sm leading-relaxed">
                            {post.content}
                          </p>

                          {/* 元信息 */}
                          <div className="flex items-center justify-between text-xs text-gray-500">
                            <div className="flex items-center space-x-3">
                              <Link
                                to={`/users/${post.author.id}`}
                                state={{ username: post.author.name }}
                                onClick={(e) => e.stopPropagation()}
                                className="font-medium text-gray-700 hover:text-primary-600"
                              >
                                {post.author.name}
                              </Link>
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
                            <div className="flex items-center space-x-3 text-gray-500">
                              <span className="flex items-center space-x-1">
                                <Eye className="h-4 w-4" />
                                <span>{post.views ?? 0}</span>
                              </span>
                              <span className="flex items-center space-x-1">
                                <Heart className="h-4 w-4" />
                                <span>{post.likes ?? 0}</span>
                              </span>
                              <span className="flex items-center space-x-1">
                                <MessageSquare className="h-4 w-4" />
                                <span>{post.comments ?? 0}</span>
                              </span>
                            </div>

                          </div>
                        </div>
                      </div>
                    </article>
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
            {!hasMore && !isInitialLoading && posts.length > 0 && (
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
            {posts.length === 0 && !isLoading && (
              <div className="card text-center py-12">
                <MessageSquare className="h-16 w-16 text-gray-300 mx-auto mb-4" />
                <p className="text-gray-500 text-lg">暂无帖子</p>
              </div>
            )}
          </>
        )}
      </section>

      <aside className="space-y-4 w-full">
        <div className="card sticky top-24 max-w-[320px]">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-2">
              <Flame className="h-5 w-5 text-primary-600" />
              <h2 className="text-lg font-semibold text-gray-900">热门榜单</h2>
            </div>
            <span className="text-xs text-gray-500">示例数据</span>
          </div>
          <ol className="space-y-3">
            {mockHotPosts.map((hot, index) => (
              <li key={hot.id} className="flex items-start space-x-3">
                <div className="w-8 h-8 rounded-lg bg-primary-50 text-primary-700 font-semibold flex items-center justify-center">
                  {index + 1}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-gray-900 hover:text-primary-600 transition-colors line-clamp-2">
                    {hot.title}
                  </p>
                  <p className="text-xs text-gray-500 mt-1">热度 {hot.heat}</p>
                </div>
              </li>
            ))}
          </ol>
        </div>
      </aside>
    </div>
  )
}
