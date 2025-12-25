import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { Search, Flame, Filter, Clock, Sparkles, Tag, ArrowUpRight, LayoutGrid, BarChart3 } from 'lucide-react'
import { formatRelativeTime } from '../utils/date'
import { Post } from '../types'

type SearchResultItem = Post & {
  summary?: string
  matchScore?: number
  badge?: string
}

const searchCategories = [
  { key: 'all', label: '全部' },
  { key: 'frontend', label: '前端' },
  { key: 'backend', label: '后端' },
  { key: 'go', label: 'Go' },
  { key: 'java', label: 'Java' },
  { key: 'python', label: 'Python' },
  { key: 'ai', label: 'AI' },
  { key: 'ops', label: '运维' },
]

const categoryLabelMap = searchCategories.reduce<Record<string, string>>((acc, cur) => {
  acc[cur.key] = cur.label
  return acc
}, {})

const mockSearchResults: SearchResultItem[] = [
  {
    id: '901',
    title: 'React 18 并发渲染落地与性能评估',
    summary: '记录首页 feed 重构后的性能指标，包含 Suspense 边界、搜索结果列表骨架屏以及用户行为埋点的设计取舍。',
    content: '我们在新版搜索结果页中启用了 React 18 的并发渲染与自动批处理，前后端协同优化首屏渲染时间，并通过骨架屏方案降低感知延迟。',
    author: { id: '21', name: '前端小能手' },
    createdAt: '2025-01-10T08:00:00Z',
    views: 2380,
    likes: 186,
    comments: 52,
    tags: ['React', '性能优化', '并发渲染'],
    category: 'frontend',
    matchScore: 96,
    badge: '精选',
  },
  {
    id: '902',
    title: 'Go 微服务中的链路追踪与可观测性最佳实践',
    summary: '用 OpenTelemetry 打通搜索接口、推荐接口的链路，结合 Jaeger、Prometheus 追踪慢查询并优化索引策略。',
    content: '搜索接口的 P99 延迟优化从慢查询定位开始，通过链路追踪找出瓶颈，并用缓存+异步索引刷新降低冷启动成本。',
    author: { id: '7', name: '后端老王' },
    createdAt: '2025-01-05T12:30:00Z',
    views: 1986,
    likes: 143,
    comments: 41,
    tags: ['Go', '微服务', '链路追踪'],
    category: 'backend',
    matchScore: 91,
    badge: '实战',
  },
  {
    id: '903',
    title: 'AI Agent 在客服场景的对话记忆设计',
    summary: '围绕搜索意图、召回、重排的多轮记忆方案，使用向量索引与摘要缓存让 Agent 回答更加稳定。',
    content: '我们为搜索结果页接入 Agent 推荐模块，利用短期记忆缓存和长期向量召回，提升用户后续点击的转化。',
    author: { id: '33', name: '产品拆解手册' },
    createdAt: '2024-12-28T09:10:00Z',
    views: 1650,
    likes: 132,
    comments: 28,
    tags: ['LLM', 'Agent', 'RAG'],
    category: 'ai',
    matchScore: 87,
    badge: '新',
  },
  {
    id: '904',
    title: '前端搜索体验设计：防抖、空状态与分词提示',
    summary: '拆解搜索框交互，涵盖输入节流、防抖、历史记录、快捷标签与「无结果」兜底的 UI 设计。',
    content: '为了让搜索结果页面更可感知，加入了查询词高亮、相关推荐卡片和热词榜，确保在无结果时仍有可探索路径。',
    author: { id: '4', name: '设计灵感库' },
    createdAt: '2025-01-11T15:00:00Z',
    views: 1420,
    likes: 118,
    comments: 24,
    tags: ['搜索体验', '交互设计', '防抖'],
    category: 'frontend',
    matchScore: 94,
    badge: '体验',
  },
  {
    id: '905',
    title: '数据库查询优化：索引失效与覆盖索引实战',
    summary: '通过 Explain 逐条分析搜索落库的慢 SQL，讲解索引失效的 5 个常见原因与解决方案。',
    content: '针对搜索历史统计的表设计，使用组合索引和覆盖索引提升查询性能，并利用延迟写入降低锁竞争。',
    author: { id: '18', name: '数据攻城狮' },
    createdAt: '2024-12-20T07:45:00Z',
    views: 1766,
    likes: 101,
    comments: 30,
    tags: ['数据库', '索引', '性能'],
    category: 'backend',
    matchScore: 83,
    badge: '案例',
  },
  {
    id: '906',
    title: 'Vue3 + Vite 落地企业级组件设计体系',
    summary: '在搜索结果页中复用卡片、标签、侧边栏等基础组件，输出约定式的 Token 与响应式栅格方案。',
    content: '通过设计 Token 保证搜索页主题一致性，利用原子化布局快速搭建候选项、筛选区、洞察面板的 UI。',
    author: { id: '11', name: '前端生信' },
    createdAt: '2025-01-03T11:15:00Z',
    views: 1214,
    likes: 92,
    comments: 19,
    tags: ['Vue3', '设计体系', '组件化'],
    category: 'frontend',
    matchScore: 82,
  },
  {
    id: '907',
    title: 'DevOps 实战：CI/CD 里如何预览搜索结果页',
    summary: '利用 Preview 环境 + Mock 数据的方式，在合并前就能验证搜索 UI、空状态、性能指标。',
    content: '针对搜索结果页的 PR，我们在流水线上自动生成预览链接，同时注入假数据与性能监测脚本，确保体验一致。',
    author: { id: '27', name: '运维之光' },
    createdAt: '2024-12-26T13:40:00Z',
    views: 980,
    likes: 74,
    comments: 16,
    tags: ['DevOps', 'CI/CD', '预览环境'],
    category: 'ops',
    matchScore: 76,
  },
  {
    id: '908',
    title: 'TypeScript 类型系统在搜索模块的约束设计',
    summary: '通过类型守卫约束搜索结果结构，减少后续重构风险，并给出前后端数据契约示例。',
    content: '为搜索结果卡片定义了稳定的类型接口，结合 Zod 校验与编译时提示，避免字段缺失导致的渲染异常。',
    author: { id: '5', name: 'Type极客' },
    createdAt: '2025-01-08T17:20:00Z',
    views: 1104,
    likes: 88,
    comments: 22,
    tags: ['TypeScript', '类型体操', '工程化'],
    category: 'frontend',
    matchScore: 85,
  },
  {
    id: '909',
    title: 'Python 数据分析：为搜索结果做点击率预测',
    summary: '用 Pandas + scikit-learn 训练 CTR 模型，配合特征重要性分析指导前端排版与曝光策略。',
    content: '我们对搜索结果的点击数据进行了特征工程，筛选出影响点击率的核心维度，并将结果回灌到排序策略。',
    author: { id: '30', name: '数据小白' },
    createdAt: '2025-01-02T10:05:00Z',
    views: 880,
    likes: 63,
    comments: 18,
    tags: ['Python', '数据分析', 'CTR'],
    category: 'python',
    matchScore: 78,
  },
  {
    id: '910',
    title: 'Java 高并发：限流与降级在搜索接口的落地',
    summary: '讲解搜索接口在高峰期的限流、熔断与降级策略，并给出网关层与应用层的配置示例。',
    content: '通过滑动窗口限流保护搜索接口，结合缓存兜底策略，保证搜索结果页在突发流量下仍可用。',
    author: { id: '16', name: 'Java手册' },
    createdAt: '2024-12-18T16:30:00Z',
    views: 1320,
    likes: 80,
    comments: 21,
    tags: ['Java', '高并发', '限流'],
    category: 'java',
    matchScore: 80,
  },
]

const quickTags = ['搜索体验', '性能优化', '并发', '微服务', '组件化', '类型系统', 'DevOps', 'AI Agent', '数据库']

const sortOptions = [
  { key: 'relevance', label: '相关度' },
  { key: 'latest', label: '最新' },
  { key: 'hot', label: '热度' },
] as const

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

const computeHeat = (post: SearchResultItem) =>
  (post.likes ?? 0) * 3 + (post.comments ?? 0) * 4 + (post.views ?? 0) * 0.4 + (post.matchScore ?? 0)

export default function SearchPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [keyword, setKeyword] = useState(searchParams.get('q')?.trim() ?? '')
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [sort, setSort] = useState<(typeof sortOptions)[number]['key']>('relevance')

  useEffect(() => {
    setKeyword(searchParams.get('q')?.trim() ?? '')
  }, [searchParams])

  const filteredResults = useMemo(() => {
    const query = keyword.trim().toLowerCase()
    const byCategory = mockSearchResults.filter((post) => {
      if (selectedCategory !== 'all' && post.category !== selectedCategory) return false
      if (!query) return true
      const pool = `${post.title} ${post.content} ${post.author.name} ${(post.tags || []).join(' ')}`
      return pool.toLowerCase().includes(query)
    })

    const sortBy = [...byCategory].sort((a, b) => {
      if (sort === 'latest') {
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      }
      if (sort === 'hot') {
        return computeHeat(b) - computeHeat(a)
      }
      const relevanceScore = (post: SearchResultItem) => {
        if (!query) {
          return (post.matchScore ?? 0) + computeHeat(post) / 10
        }
        let score = 0
        if (post.title.toLowerCase().includes(query)) score += 6
        if (post.content.toLowerCase().includes(query)) score += 3
        if ((post.tags || []).some((tag) => tag.toLowerCase().includes(query))) score += 2
        return score + (post.matchScore ?? 0) / 5 + computeHeat(post) / 500
      }

      return relevanceScore(b) - relevanceScore(a)
    })

    return sortBy
  }, [keyword, selectedCategory, sort])

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const next = keyword.trim()
    setSearchParams(next ? { q: next } : {})
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
              <p className="text-gray-600 mt-1">基于热门话题与模拟数据，快速预览社区里的高质量帖子。</p>
            </div>
            <div className="hidden sm:flex items-center gap-3 text-sm text-gray-600">
              <div className="flex items-center gap-2 bg-white/60 px-3 py-2 rounded-lg border border-primary-100 shadow-sm">
                <LayoutGrid className="h-4 w-4 text-primary-600" />
                <span>共 {mockSearchResults.length} 条示例数据</span>
              </div>
              <div className="flex items-center gap-2 bg-white/60 px-3 py-2 rounded-lg border border-primary-100 shadow-sm">
                <BarChart3 className="h-4 w-4 text-primary-600" />
                <span>实时筛选与排序</span>
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
              <span className="text-xs text-gray-500">快捷标签：</span>
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
              <div className="flex flex-wrap items-center gap-2">
                {sortOptions.map((option) => (
                  <button
                    key={option.key}
                    onClick={() => setSort(option.key)}
                    className={`px-3 py-1.5 text-sm rounded-lg border transition-colors ${
                      sort === option.key
                        ? 'bg-primary-50 border-primary-200 text-primary-700'
                        : 'bg-white border-gray-200 text-gray-700 hover:border-primary-200'
                    }`}
                  >
                    {option.label}
                  </button>
                ))}
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
            {filteredResults.length === 0 && (
              <div className="card text-center py-12">
                <Search className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                <p className="text-gray-800 font-semibold mb-1">还没有找到匹配的内容</p>
                <p className="text-sm text-gray-500">试试换个关键词，或者调整筛选条件</p>
              </div>
            )}

            {filteredResults.map((post) => (
              <article
                key={post.id}
                className="card p-4 sm:p-5 hover:-translate-y-0.5 transition-all cursor-pointer"
                onClick={() => handleOpenPost(post.id)}
              >
                <div className="flex flex-col gap-3">
                  <div className="flex items-center gap-2 flex-wrap">
                    {post.badge && (
                      <span className="inline-flex items-center px-2 py-0.5 text-[11px] rounded-full bg-primary-50 text-primary-700 border border-primary-100">
                        {post.badge}
                      </span>
                    )}
                    {post.category && (
                      <span className="inline-flex items-center px-2 py-0.5 text-[11px] rounded-full bg-gray-100 text-gray-700">
                        {categoryLabelMap[post.category] || '其他'}
                      </span>
                    )}
                    <span className="inline-flex items-center px-2 py-0.5 text-[11px] rounded-full bg-orange-50 text-orange-700 border border-orange-100">
                      匹配度 {post.matchScore ?? 72}%
                    </span>
                  </div>

                  <div className="flex items-start gap-3">
                    <div className="flex-shrink-0">
                      <img
                        src={`https://api.dicebear.com/7.x/avataaars/svg?seed=search-${post.author.id}`}
                        alt={post.author.name}
                        className="w-11 h-11 rounded-full border border-gray-200"
                      />
                    </div>
                    <div className="flex-1 min-w-0 space-y-2">
                      <div className="flex items-start justify-between gap-3">
                        <h3 className="text-lg font-semibold text-gray-900 leading-snug line-clamp-2">
                          {highlightText(post.title, keyword)}
                        </h3>
                        <ArrowUpRight className="h-5 w-5 text-gray-400 flex-shrink-0" />
                      </div>
                      <p className="text-sm text-gray-600 leading-relaxed line-clamp-2 sm:line-clamp-3">
                        {highlightText(post.summary || post.content, keyword)}
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
            <p className="text-3xl font-bold text-gray-900 mb-1">{mockSearchResults.length}</p>
            <p className="text-sm text-gray-600">条可用的模拟搜索结果，支持分类与排序。</p>
            <p className="mt-3 text-sm text-gray-500">
              当前页面演示了搜索结果的排版与筛选，后续可接入真实接口替换模拟数据。
            </p>
          </div>
        </aside>
      </div>
    </div>
  )
}
