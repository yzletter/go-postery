import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { MessageSquare, Clock, Loader2, Eye, Heart, Flame, UserPlus, Star, Grid, Code2, Server, Bot, Goal, Braces, Coffee, Pi, Gift, Sparkles } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Post } from '../types'
import { normalizePost } from '../utils/post'
import { buildIdSeed } from '../utils/id'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { apiGet } from '../utils/api'

type CategoryItem = { key: string; label: string; icon: LucideIcon; tag?: string }

const categories: CategoryItem[] = [
  { key: 'follow', label: '关注', icon: Star },
  { key: 'all', label: '全部', icon: Grid },
  { key: 'frontend', label: '前端', tag: '前端', icon: Code2 },
  { key: 'backend', label: '后端', tag: '后端', icon: Server },
  { key: 'go', label: 'Go', tag: 'Go', icon: Goal },
  { key: 'cpp', label: 'C++', tag: 'C++', icon: Braces },
  { key: 'java', label: 'Java', tag: 'Java', icon: Coffee },
  { key: 'python', label: 'Python', tag: 'Python', icon: Pi },
  { key: 'ai', label: 'AI', tag: 'AI', icon: Bot },
]

const getRequestTag = (categoryKey: string): string => {
  const match = categories.find(cat => cat.key === categoryKey)
  return match?.tag ?? match?.label ?? categoryKey
}

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
    id: String(id),
    title: titles[index % titles.length],
    content: contents[index % contents.length],
    author: {
      id: String(author.id),
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

const mockRecommendUsers = [
  { id: 101, name: '前端小能手', title: '分享 React / TS 实战', followers: 12.4 },
  { id: 102, name: 'Go 语言爱好者', title: 'Go / 微服务 / 云原生', followers: 8.6 },
  { id: 103, name: '设计灵感库', title: 'UI/UX 灵感与案例', followers: 15.2 },
  { id: 104, name: '后端老王', title: '性能调优与架构实践', followers: 6.8 },
  { id: 105, name: '产品拆解手册', title: '产品思考与需求分析', followers: 9.1 },
  { id: 106, name: '测试小白进阶', title: '自动化测试 / 质量保障', followers: 5.4 },
]

// API 获取帖子列表
const FETCH_TIMEOUT_MS = 8000
const DEFAULT_PAGE_SIZE = 10
const CATEGORY_PAGE_SIZE = 10

const fetchPosts = async (page: number, pageSize: number = DEFAULT_PAGE_SIZE, categoryKey?: string): Promise<PostListResult> => {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS)

  try {
    const useAllEndpoint = !categoryKey || categoryKey === 'all' || categoryKey === 'follow'
    const tag = categoryKey ? getRequestTag(categoryKey) : ''
    const path = useAllEndpoint
      ? `/posts?pageNo=${page}&pageSize=${pageSize}`
      : `/posts_tag?pageNo=${page}&pageSize=${pageSize}&tag=${encodeURIComponent(tag)}`

    const { data } = await apiGet<{
      posts: any[]
      total?: number
      hasMore?: boolean
    }>(path, { signal: controller.signal })

    const rawPosts = Array.isArray(data?.posts)
      ? data.posts
      : data?.posts == null
        ? []
        : null

    if (!data || rawPosts === null) {
      throw new Error('帖子列表响应数据格式错误')
    }

    const postsWithStats: Post[] = rawPosts.map((p: any) => {
      const normalized = normalizePost(p)
      return {
        ...normalized,
        views: normalized.views ?? 0,
        likes: normalized.likes ?? 0,
        comments: normalized.comments ?? 0,
      }
    })

    return {
      posts: postsWithStats,
      total: data.total ?? postsWithStats.length,
      hasMore: typeof data.hasMore === 'boolean' ? data.hasMore : postsWithStats.length >= pageSize,
    }
  } catch (error) {
    if ((error as any)?.name === 'AbortError') {
      console.error('Fetch posts request timeout')
      throw new Error('请求超时，请检查后端服务状态')
    }
    console.error('Failed to fetch posts:', error)
    throw error
  } finally {
    clearTimeout(timeoutId)
  }
}

export default function Home() {
  const [posts, setPosts] = useState<Post[]>([])
  const [currentPage, setCurrentPage] = useState(0)
  const [isLoading, setIsLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [isInitialLoading, setIsInitialLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const navigate = useNavigate()
  const observerTarget = useRef<HTMLDivElement>(null)
  const isLoadingRef = useRef(false)

  // 确保进入首页时回到顶部
  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: 'auto' })
  }, [])

  const categoryPool = useMemo(() => categories.filter(c => c.key !== 'all'), [])

  const decoratePosts = useCallback((list: Post[], options: { offset?: number; fallbackCategory?: string } = {}): Post[] => {
    const { offset = 0, fallbackCategory } = options
    const hasCategories = categoryPool.length > 0
    return list.map((post, idx) => {
      const seed = buildIdSeed(post.id, idx + offset)
      const randomCategory = hasCategories ? categoryPool[seed % categoryPool.length] : undefined
      const category = post.category ?? fallbackCategory ?? randomCategory?.key
      const tags = (post.tags ?? [])
        .map(tag => (typeof tag === 'string' ? tag.trim() : ''))
        .filter(Boolean)
      return {
        ...post,
        category,
        tags: tags.length > 0 ? tags : undefined,
      }
    })
  }, [categoryPool])

  const getPageSizeForCategory = useCallback((categoryKey: string) => {
    if (!categoryKey || categoryKey === 'all' || categoryKey === 'follow') {
      return DEFAULT_PAGE_SIZE
    }
    return CATEGORY_PAGE_SIZE
  }, [])

  // 加载帖子数据
  const loadPosts = useCallback(async (page: number, reset: boolean = false, categoryKey?: string) => {
    if (isLoadingRef.current) return
    
    const targetCategory = categoryKey ?? selectedCategory
    const pageSize = getPageSizeForCategory(targetCategory)

    if (reset) {
      setIsInitialLoading(true)
      setPosts([])
      setHasMore(true)
      setCurrentPage(0)
      setError(null)
    } else {
      setError(null)
    }

    isLoadingRef.current = true
    setIsLoading(true)
    try {
      if (!reset) {
        await new Promise(resolve => setTimeout(resolve, 500))
      }
      
      const { posts: newPosts, hasMore: hasMoreFromApi } = await fetchPosts(page, pageSize, targetCategory)
      setPosts(prev => {
        const offset = reset ? 0 : prev.length
        const decorated = decoratePosts(newPosts, {
          offset,
          fallbackCategory: targetCategory !== 'all' ? targetCategory : undefined
        })
        return reset ? decorated : [...prev, ...decorated]
      })
      setHasMore(hasMoreFromApi)
      setCurrentPage(page)
    } catch (error) {
      console.error('Failed to load posts:', error)
      setError(error instanceof Error ? error.message : '加载帖子失败')
      setHasMore(false)
    } finally {
      setIsLoading(false)
      setIsInitialLoading(false)
      isLoadingRef.current = false
    }
  }, [decoratePosts, getPageSizeForCategory, selectedCategory])

  // 初始加载帖子
  useEffect(() => {
    loadPosts(1, true, selectedCategory)
  }, [loadPosts, selectedCategory])

  // 无限滚动：监听滚动到底部
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !isLoading && !isInitialLoading && !error) {
          loadPosts(currentPage + 1, false, selectedCategory)
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
  }, [hasMore, isLoading, isInitialLoading, currentPage, loadPosts, error])

  const filteredPosts = useMemo(() => {
    if (selectedCategory === 'all') return posts
    return posts.filter(post => post.category === selectedCategory)
  }, [posts, selectedCategory])

  // 1111
  return (
    <div className="grid lg:grid-cols-[60px_minmax(0,9fr)_300px] gap-3 items-start">
      <aside className="hidden lg:block w-[140px] lg:sticky lg:top-20 self-start lg:-ml-20">
        <div className="space-y-3">
          <div className="card p-3 space-y-3">
            <div className="space-y-2">
              {categories.map(cat => {
                const Icon = cat.icon
                return (
                  <button
                    key={cat.key}
                    onClick={() => setSelectedCategory(cat.key)}
                    className={`inline-flex w-full items-center justify-start gap-2 pl-3.5 pr-2.5 py-2 rounded-lg border transition-all hover:-translate-x-0.5 focus:outline-none focus:ring-2 focus:ring-primary-200 ${
                      selectedCategory === cat.key
                        ? 'bg-primary-50 border-primary-200 text-primary-700 shadow-sm ring-1 ring-primary-100'
                        : 'border-gray-200 bg-white hover:border-primary-200 hover:text-primary-700 hover:bg-primary-50/60'
                    }`}
                  >
                    <Icon className="h-4 w-4" />
                    <span>{cat.label}</span>
                  </button>
                )
              })}
            </div>
          </div>
        </div>
      </aside>

      <section className="flex flex-col gap-6">
        <div className="flex items-center gap-2 overflow-x-auto lg:hidden pb-2">
          {categories.map(cat => {
            const Icon = cat.icon
            return (
              <button
                key={cat.key}
                onClick={() => setSelectedCategory(cat.key)}
                className={`inline-flex items-center gap-1 pl-3.5 pr-3 py-1.5 rounded-full text-sm border whitespace-nowrap ${
                  selectedCategory === cat.key
                    ? 'bg-primary-50 border-primary-200 text-primary-700'
                    : 'bg-white border-gray-200 text-gray-700'
                }`}
              >
                <Icon className="h-4 w-4" />
                <span>{cat.label}</span>
              </button>
            )
          })}
        </div>

        {/* 初始加载状态 */}
        {isInitialLoading && (
          <div className="card text-center py-12">
            <Loader2 className="h-8 w-8 text-primary-600 animate-spin mx-auto mb-4" />
            <p className="text-gray-500">加载中...</p>
          </div>
        )}

        {error && !isInitialLoading && (
          <div className="card border border-red-200 bg-red-50 text-red-700">
            <div className="flex items-start justify-between space-x-3">
              <div>
                <p className="font-semibold">加载失败</p>
                <p className="text-sm text-red-600 break-words">{error}</p>
              </div>
              <button
                type="button"
                onClick={() => loadPosts(1, true)}
                className="btn-secondary bg-white text-red-700 hover:bg-red-100"
              >
                重试
              </button>
            </div>
          </div>
        )}

        {/* 帖子列表 */}
        {!isInitialLoading && (
          <>
            <div className="space-y-3">
              {filteredPosts.length === 0 && !isLoading && (
                <div className="card text-center py-10">
                  <p className="text-gray-700 font-medium mb-1">该分类下暂时没有帖子</p>
                  <p className="text-sm text-gray-500">试试切换到其他分类或稍后再来～</p>
                </div>
              )}
              {filteredPosts.map(post => (
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

                      {/* 标签 */}
                      <div className="flex flex-wrap items-center gap-2 mb-3">
                        {post.tags?.map(tag => (
                          <span
                            key={tag}
                            className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full border border-primary-100 bg-primary-50/70 text-primary-700 text-xs font-medium shadow-sm"
                          >
                            {tag}
                          </span>
                        ))}
                      </div>

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
            {posts.length === 0 && !isLoading && !error && (
              <div className="card text-center py-12">
                <MessageSquare className="h-16 w-16 text-gray-300 mx-auto mb-4" />
                <p className="text-gray-500 text-lg">暂无帖子</p>
              </div>
            )}
          </>
        )}
      </section>

      <aside className="w-full lg:ml-6 xl:ml-0">
        <div className="sticky top-24 space-y-4 max-w-[320px]">
          <div className="card bg-gradient-to-br from-primary-50 via-white to-white border-primary-100/60 relative overflow-hidden">
            <div className="absolute -right-10 -top-10 w-28 h-28 bg-primary-100/60 rounded-full blur-2xl" />
            <div className="absolute -left-8 bottom-0 w-24 h-24 bg-white/60 border border-primary-50 rounded-full blur-2xl" />
            <div className="relative space-y-4">
              <div className="flex items-center gap-2">
                <Gift className="h-5 w-5 text-primary-700" />
                <h2 className="text-lg font-semibold text-gray-900">今日抽奖</h2>
                <span className="text-xs text-primary-700 bg-primary-100 px-2 py-0.5 rounded-full border border-primary-200">模拟</span>
              </div>
              <p className="text-sm text-gray-600">每日签到即可抽奖，会员、积分、限定徽章等你拿～</p>
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={() => navigate('/lottery')}
                  className="btn-primary flex-1 flex items-center justify-center gap-2 shadow-sm"
                >
                  <Sparkles className="h-4 w-4" />
                  立即抽奖
                </button>
              </div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <Flame className="h-5 w-5 text-primary-600" />
                <h2 className="text-lg font-semibold text-gray-900">热门文章</h2>
              </div>
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

          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <UserPlus className="h-5 w-5 text-primary-600" />
                <h2 className="text-lg font-semibold text-gray-900">推荐关注</h2>
              </div>
            </div>
            <div className="space-y-3">
              {mockRecommendUsers.map((user) => (
                <div key={user.id} className="flex items-center space-x-2 p-1.5 rounded-lg hover:bg-gray-50 transition-colors">
                  <Link
                    to={`/users/${user.id}`}
                    state={{ username: user.name }}
                    className="flex-shrink-0"
                  >
                    <img
                      src={`https://api.dicebear.com/7.x/avataaars/svg?seed=recommend-${user.id}`}
                      alt={user.name}
                      className="w-10 h-10 rounded-full"
                    />
                  </Link>
                  <div className="flex-1 min-w-0 flex items-center">
                    <div className="flex-1 min-w-0">
                      <Link
                        to={`/users/${user.id}`}
                        state={{ username: user.name }}
                        className="text-sm font-medium text-gray-900 hover:text-primary-600 transition-colors line-clamp-1"
                      >
                        {user.name}
                      </Link>
                      <p className="text-xs text-gray-500 line-clamp-1">{user.title}</p>
                    </div>
                    <span className="text-xs text-primary-600 ml-3 w-12 text-right flex-shrink-0">
                      {user.followers}k
                    </span>
                  </div>
                  <button className="text-xs text-primary-600 font-medium hover:text-primary-700 flex-shrink-0">
                    关注
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div className="card text-center text-xs text-gray-500 leading-relaxed">
            <p className="font-semibold text-gray-700">© 2025 Go Postery</p>
            <p>内容版权归原作者所有，转载请注明出处。</p>
            <p className="text-gray-400">如有侵权或合作事宜，请联系管理员处理。</p>
          </div>
        </div>
      </aside>
    </div>
  )
}
