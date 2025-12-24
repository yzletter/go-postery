import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { MessageSquare, Clock, Loader2, Eye, Heart, Flame, UserPlus, Gift, Sparkles } from 'lucide-react'
import type { Post } from '../types'
import { buildIdSeed } from '../utils/id'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { CATEGORY_PAGE_SIZE, DEFAULT_PAGE_SIZE, categories, mockRecommendUsers } from './home/constants'
import { fetchPosts } from './home/fetchPosts'
import { fetchTopPosts, type TopPost } from './home/fetchTopPosts'

export default function Home() {
  const [posts, setPosts] = useState<Post[]>([])
  const [currentPage, setCurrentPage] = useState(0)
  const [isLoading, setIsLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [isInitialLoading, setIsInitialLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [hotPosts, setHotPosts] = useState<TopPost[]>([])
  const [isHotLoading, setIsHotLoading] = useState(false)
  const [hotError, setHotError] = useState<string | null>(null)
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

  const loadHotPosts = useCallback(async () => {
    setIsHotLoading(true)
    setHotError(null)

    try {
      const topPosts = await fetchTopPosts()
      setHotPosts(topPosts)
    } catch (error) {
      console.error('Failed to load top posts:', error)
      setHotError(error instanceof Error ? error.message : '加载热门帖子失败')
    } finally {
      setIsHotLoading(false)
    }
  }, [fetchTopPosts])

  // 初始加载帖子
  useEffect(() => {
    loadPosts(1, true, selectedCategory)
  }, [loadPosts, selectedCategory])

  useEffect(() => {
    void loadHotPosts()
  }, [loadHotPosts])

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
  }, [hasMore, isLoading, isInitialLoading, currentPage, loadPosts, error, selectedCategory])

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
              {isHotLoading && hotPosts.length === 0 && (
                <li className="text-sm text-gray-500 flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>加载中...</span>
                </li>
              )}
              {!isHotLoading && hotError && hotPosts.length === 0 && (
                <li className="text-sm text-red-600 flex items-center justify-between gap-2">
                  <span className="flex-1 break-words">{hotError}</span>
                  <button
                    type="button"
                    onClick={loadHotPosts}
                    className="text-xs text-primary-600 hover:text-primary-700"
                  >
                    重试
                  </button>
                </li>
              )}
              {!isHotLoading && !hotError && hotPosts.length === 0 && (
                <li className="text-sm text-gray-500">暂无热门文章</li>
              )}
              {hotPosts.map((hot, index) => (
                <li key={hot.id}>
                  <button
                    type="button"
                    onClick={() => navigate(`/post/${hot.id}`)}
                    className="group flex items-start space-x-3 w-full text-left rounded-lg focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-200"
                  >
                    <div className="w-7 h-7 rounded-md bg-primary-50 text-primary-700 text-xs font-semibold flex items-center justify-center">
                      {index + 1}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-[11px] font-semibold text-gray-900 group-hover:text-primary-600 transition-colors line-clamp-2">
                        {hot.title}
                      </p>
                      <p className="text-xs text-gray-500 mt-1">热度 {hot.score}</p>
                    </div>
                  </button>
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
