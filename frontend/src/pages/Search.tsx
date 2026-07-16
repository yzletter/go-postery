import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { Search, Flame, Filter, Clock, Sparkles, Tag, ArrowUpRight, LayoutGrid, BarChart3, Loader2 } from 'lucide-react'
import UserAvatar from '../components/UserAvatar'
import { formatRelativeTime } from '../utils/date'
import { ApiError, apiPost } from '../utils/api'
import { normalizePost } from '../utils/post'
import { Post } from '../types'

type SearchResultItem = Post & {
  summary?: string
}

const SEARCH_TIMEOUT_MS = 8000
const SUMMARY_LIMIT = 160

const buildSummary = (content: string) => {
  if (!content) return ''
  const normalized = content.replace(/\s+/g, ' ').trim()
  if (!normalized) return ''
  if (normalized.length <= SUMMARY_LIMIT) return normalized
  return `${normalized.slice(0, SUMMARY_LIMIT)}...`
}

const normalizeSearchItem = (raw: any): SearchResultItem | null => {
  const normalized = normalizePost(raw)
  if (!normalized.id || !normalized.title) return null
  const summary = buildSummary(normalized.content)
  return {
    ...normalized,
    views: normalized.views ?? 0,
    likes: normalized.likes ?? 0,
    comments: normalized.comments ?? 0,
    summary: summary || undefined,
  }
}

const quickTags = ['搜索体验', '性能优化', '并发', '微服务', '组件化', '类型系统', 'DevOps', 'AI Agent', '数据库']

const escapeRegExp = (value: string) => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const highlightText = (text: string, keyword: string) => {
  if (!keyword.trim()) return text
  const escaped = escapeRegExp(keyword.trim())
  if (!escaped) return text
  const regex = new RegExp(`(${escaped})`, 'gi')
  const lower = keyword.trim().toLowerCase()

  return text.split(regex).map((part, index) => {
    const isMatch = part.toLowerCase() === lower
    return isMatch ? (
      <mark key={`${part}-${index}`} className="bg-primary-100 text-primary-700 rounded px-1 py-0.5">
        {part}
      </mark>
    ) : (
      <span key={`${part}-${index}`}>{part}</span>
    )
  })
}

export default function SearchPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const searchQuery = searchParams.get('q')?.trim() ?? ''
  const [keyword, setKeyword] = useState(searchQuery)
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [results, setResults] = useState<SearchResultItem[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasSearched, setHasSearched] = useState(false)

  useEffect(() => {
    setKeyword(searchQuery)
    setSelectedCategory('all')
  }, [searchQuery])

  useEffect(() => {
    let isActive = true
    const controller = new AbortController()
    let timeoutId: ReturnType<typeof setTimeout> | null = null

    if (!searchQuery) {
      setResults([])
      setError(null)
      setIsLoading(false)
      setHasSearched(false)
      return () => {
        controller.abort()
        if (timeoutId) clearTimeout(timeoutId)
      }
    }

    setIsLoading(true)
    setError(null)
    setHasSearched(true)
    setResults([])
    timeoutId = setTimeout(() => controller.abort(), SEARCH_TIMEOUT_MS)

    const runSearch = async () => {
      try {
        const { data } = await apiPost<any[]>('/search', { query: searchQuery }, { signal: controller.signal })
        if (!isActive) return
        if (!Array.isArray(data)) {
          throw new Error('搜索响应数据格式错误')
        }

        const normalized = data
          .map((item) => normalizeSearchItem(item))
          .filter(Boolean) as SearchResultItem[]

        setResults(normalized)
      } catch (error) {
        if (!isActive) return
        if ((error as { name?: string })?.name === 'AbortError') {
          setError('请求超时，请检查后端服务状态')
        } else if (error instanceof ApiError) {
          setError(error.message)
        } else {
          setError(error instanceof Error ? error.message : '搜索失败')
        }
        setResults([])
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
        if (timeoutId) clearTimeout(timeoutId)
      }
    }

    void runSearch()

    return () => {
      isActive = false
      controller.abort()
      if (timeoutId) clearTimeout(timeoutId)
    }
  }, [searchQuery])

  const searchCategories = useMemo(() => {
    const labelsByKey = new Map<string, string>()

    results.forEach((post) => {
      post.tags?.forEach((tag) => {
        const label = tag.trim()
        const key = label.toLocaleLowerCase()
        if (label && !labelsByKey.has(key)) {
          labelsByKey.set(key, label)
        }
      })
    })

    return [
      { key: 'all', label: '全部' },
      ...Array.from(labelsByKey, ([key, label]) => ({ key, label })),
    ]
  }, [results])

  const filteredResults = useMemo(() => {
    return results.filter((post) => {
      if (
        selectedCategory !== 'all' &&
        !post.tags?.some((tag) => tag.trim().toLocaleLowerCase() === selectedCategory)
      ) {
        return false
      }
      return true
    })
  }, [results, selectedCategory])

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const next = keyword.trim()
    setSearchParams(next ? { q: next } : {})
    setSelectedCategory('all')
  }

  const handleQuickTag = (tag: string) => {
    setKeyword(tag)
    setSelectedCategory('all')
    setSearchParams({ q: tag })
  }

  const handleOpenPost = (id: string) => {
    navigate(`/post/${id}`)
  }

  const totalCount = filteredResults.length
  const rawCount = results.length
  const emptyStateTitle = hasSearched ? '还没有找到匹配的内容' : '输入关键词开始搜索'
  const emptyStateHint = hasSearched ? '试试换个关键词，或者调整标签筛选' : '支持标题与正文关键词检索'

  return (
    <div className="space-y-6">
      <section className="card relative overflow-hidden bg-gradient-to-r from-primary-50 via-white to-white border-primary-100/60">
        <div className="absolute -right-12 -top-12 w-40 h-40 bg-primary-100/60 rounded-full blur-3xl" />
        <div className="absolute -left-16 bottom-0 w-48 h-48 bg-white/40 border border-primary-50 rounded-full blur-3xl" />
        <div className="relative space-y-4">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-sm text-primary-700 font-semibold inline-flex items-center gap-2 px-2.5 py-1 rounded-full bg-primary-100 border border-primary-200">
                <Sparkles className="h-4 w-4" />
                搜索结果
              </p>
              <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 mt-2">找到你想要的内容</h1>
              <p className="text-gray-600 mt-1">基于标题和正文关键词检索社区帖子。</p>
            </div>
            <div className="hidden sm:flex items-center gap-3 text-sm text-gray-600">
              <div className="flex items-center gap-2 bg-white/60 px-3 py-2 rounded-lg border border-primary-100 shadow-sm">
                <LayoutGrid className="h-4 w-4 text-primary-600" />
                <span>共 {rawCount} 条匹配结果</span>
              </div>
              <div className="flex items-center gap-2 bg-white/60 px-3 py-2 rounded-lg border border-primary-100 shadow-sm">
                <BarChart3 className="h-4 w-4 text-primary-600" />
                <span>后端实时搜索</span>
              </div>
            </div>
          </div>

          <form onSubmit={handleSearchSubmit} className="space-y-3">
            <div className="flex flex-col sm:flex-row gap-3">
              <div className="relative flex-1">
                <Search className="h-5 w-5 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
                <input
                  value={keyword}
                  onChange={(e) => setKeyword(e.target.value)}
                  placeholder="输入关键词，如「搜索体验」「并发」「Agent」"
                  className="input h-12 pl-10 pr-4 bg-white/80 border-primary-100 focus:border-primary-300 focus:ring-primary-200 shadow-sm"
                />
              </div>
              <button type="submit" className="btn-primary h-12 px-6 shadow-sm">
                开始搜索
              </button>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs text-gray-500">快捷搜索：</span>
              {quickTags.map((tag) => (
                <button
                  key={tag}
                  type="button"
                  onClick={() => handleQuickTag(tag)}
                  className="px-3 py-1.5 text-xs rounded-full border border-primary-100 bg-white text-primary-700 hover:bg-primary-50 transition-colors"
                >
                  #{tag}
                </button>
              ))}
            </div>
          </form>
        </div>
      </section>

      <div className="grid lg:grid-cols-[minmax(0,3fr)_minmax(260px,1fr)] gap-6">
        <section className="space-y-4">
          <div className="card p-4 sm:p-5">
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-primary-50 border border-primary-100 flex items-center justify-center text-primary-700 font-semibold">
                  {totalCount}
                </div>
                <div>
                  <p className="text-sm text-gray-500">符合当前筛选条件</p>
                  <p className="text-base font-semibold text-gray-900">搜索结果</p>
                </div>
              </div>
            </div>

            <div className="mt-4 flex flex-wrap gap-2">
              {searchCategories.map((cat) => (
                <button
                  key={cat.key}
                  onClick={() => setSelectedCategory(cat.key)}
                  className={`px-3 py-1.5 text-sm rounded-full border transition-colors ${
                    selectedCategory === cat.key
                      ? 'bg-primary-600 text-white border-primary-600 shadow-sm'
                      : 'bg-white border-gray-200 text-gray-700 hover:border-primary-200'
                  }`}
                >
                  {cat.label}
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-3">
            {error && (
              <div className="card border border-red-100 bg-red-50 text-sm text-red-700">
                {error}
              </div>
            )}

            {isLoading && (
              <div className="card flex items-center justify-center gap-2 py-10 text-sm text-gray-500">
                <Loader2 className="h-4 w-4 animate-spin" />
                正在搜索内容...
              </div>
            )}

            {!isLoading && !error && filteredResults.length === 0 && (
              <div className="card text-center py-12">
                <Search className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                <p className="text-gray-800 font-semibold mb-1">{emptyStateTitle}</p>
                <p className="text-sm text-gray-500">{emptyStateHint}</p>
              </div>
            )}

            {filteredResults.map((post) => (
              <article
                key={post.id}
                className="card p-4 sm:p-5 hover:-translate-y-0.5 transition-all cursor-pointer"
                onClick={() => handleOpenPost(post.id)}
              >
                <div className="flex flex-col gap-3">
                  <div className="flex items-start gap-3">
                    <div className="flex-shrink-0">
                      <UserAvatar
                        avatar={post.author.avatar}
                        name={post.author.name}
                        userId={post.author.id}
                        className="w-11 h-11 rounded-full border border-gray-200"
                      />
                    </div>
                    <div className="flex-1 min-w-0 space-y-2">
                      <div className="flex items-start justify-between gap-3">
                        <h3 className="text-lg font-semibold text-gray-900 leading-snug line-clamp-2">
                          {highlightText(post.title, searchQuery)}
                        </h3>
                        <ArrowUpRight className="h-5 w-5 text-gray-400 flex-shrink-0" />
                      </div>
                      <p className="text-sm text-gray-600 leading-relaxed line-clamp-2 sm:line-clamp-3">
                        {highlightText(post.summary || post.content, searchQuery)}
                      </p>
                      <div className="flex flex-wrap items-center gap-2">
                        {post.tags?.map((tag) => (
                          <span
                            key={tag}
                            className="inline-flex items-center gap-1 px-2 py-1 text-[11px] rounded-full bg-gray-100 text-gray-700"
                          >
                            <Tag className="h-3 w-3" />
                            {tag}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-xs text-gray-500 pt-1">
                    <div className="flex items-center gap-3">
                      <Link
                        to={`/users/${post.author.id}`}
                        state={{ username: post.author.name }}
                        onClick={(e) => e.stopPropagation()}
                        className="font-medium text-gray-800 hover:text-primary-600"
                      >
                        {post.author.name}
                      </Link>
                      <span className="flex items-center gap-1">
                        <Clock className="h-4 w-4" />
                        {formatRelativeTime(post.createdAt)}
                      </span>
                      <span className="flex items-center gap-1">
                        <Flame className="h-4 w-4" />
                        {post.views ?? 0} 浏览
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="flex items-center gap-1">
                        <Sparkles className="h-4 w-4" />
                        {post.likes ?? 0} 赞
                      </span>
                      <span className="flex items-center gap-1">
                        <Filter className="h-4 w-4" />
                        {post.comments ?? 0} 讨论
                      </span>
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation()
                          handleOpenPost(post.id)
                        }}
                        className="text-primary-700 hover:text-primary-800 font-medium inline-flex items-center gap-1"
                      >
                        查看详情
                        <ArrowUpRight className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                </div>
              </article>
            ))}
          </div>
        </section>

        <aside className="space-y-4">
          <div className="card bg-gradient-to-br from-white via-primary-50/70 to-white border-primary-100">
            <p className="text-sm text-gray-600">总计</p>
            <p className="text-3xl font-bold text-gray-900 mb-1">{hasSearched ? rawCount : 0}</p>
            <p className="text-sm text-gray-600">
              {hasSearched ? '条匹配结果，可按返回的标签筛选。' : '输入关键词即可发起搜索。'}
            </p>
            <p className="mt-3 text-sm text-gray-500">
              {hasSearched ? `当前关键词：${searchQuery}` : '搜索将通过 /api/v1/search 返回匹配帖子。'}
            </p>
          </div>
        </aside>
      </div>
    </div>
  )
}
