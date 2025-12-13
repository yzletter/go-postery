import { Bot, Braces, Code2, Coffee, Goal, Grid, Pi, Server, Star } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

export type CategoryItem = { key: string; label: string; icon: LucideIcon; tag?: string }

export const categories: CategoryItem[] = [
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

export const getRequestTag = (categoryKey: string): string => {
  const match = categories.find(cat => cat.key === categoryKey)
  return match?.tag ?? match?.label ?? categoryKey
}

export const mockHotPosts = [
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

export const mockRecommendUsers = [
  { id: 101, name: '前端小能手', title: '分享 React / TS 实战', followers: 12.4 },
  { id: 102, name: 'Go 语言爱好者', title: 'Go / 微服务 / 云原生', followers: 8.6 },
  { id: 103, name: '设计灵感库', title: 'UI/UX 灵感与案例', followers: 15.2 },
  { id: 104, name: '后端老王', title: '性能调优与架构实践', followers: 6.8 },
  { id: 105, name: '产品拆解手册', title: '产品思考与需求分析', followers: 9.1 },
  { id: 106, name: '测试小白进阶', title: '自动化测试 / 质量保障', followers: 5.4 },
]

export const FETCH_TIMEOUT_MS = 8000
export const DEFAULT_PAGE_SIZE = 10
export const CATEGORY_PAGE_SIZE = 10

