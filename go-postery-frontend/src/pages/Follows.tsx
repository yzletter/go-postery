import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Users, HeartHandshake, ShieldAlert } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import type { FollowRelation, FollowUser } from '../types'
import { followUser, getFollowRelation, isFollowing, listFollowers, listFollowees, unfollowUser } from '../utils/follow'

type TabKey = 'following' | 'followers' | 'blocked'

const relationLabelMap: Record<FollowRelation, string> = {
  0: '互不关注',
  1: '已关注',
  2: '关注你',
  3: '互相关注',
}

export default function Follows() {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState<TabKey>('following')
  const [followees, setFollowees] = useState<FollowUser[]>([])
  const [followers, setFollowers] = useState<FollowUser[]>([])
  const [followeesLoading, setFolloweesLoading] = useState(false)
  const [followersLoading, setFollowersLoading] = useState(false)
  const [followeesError, setFolloweesError] = useState<string | null>(null)
  const [followersError, setFollowersError] = useState<string | null>(null)
  const [relationById, setRelationById] = useState<Record<string, FollowRelation>>({})
  const [actingId, setActingId] = useState<string | null>(null)

  const hydrateRelations = useCallback(async (users: FollowUser[]) => {
    if (!users.length) return

    const entries = await Promise.all(
      users.map(async (u) => {
        try {
          const relation = await getFollowRelation(u.id)
          return [u.id, relation] as const
        } catch {
          return [u.id, 0 as FollowRelation] as const
        }
      })
    )

    setRelationById(prev => ({
      ...prev,
      ...Object.fromEntries(entries),
    }))
  }, [])

  const reloadFollowees = useCallback(async () => {
    setFolloweesLoading(true)
    setFolloweesError(null)
    try {
      const list = await listFollowees()
      setFollowees(list)
      await hydrateRelations(list)
    } catch (error) {
      setFollowees([])
      setFolloweesError(error instanceof Error ? error.message : '获取关注列表失败')
    } finally {
      setFolloweesLoading(false)
    }
  }, [hydrateRelations])

  const reloadFollowers = useCallback(async () => {
    setFollowersLoading(true)
    setFollowersError(null)
    try {
      const list = await listFollowers()
      setFollowers(list)
      await hydrateRelations(list)
    } catch (error) {
      setFollowers([])
      setFollowersError(error instanceof Error ? error.message : '获取粉丝列表失败')
    } finally {
      setFollowersLoading(false)
    }
  }, [hydrateRelations])

  useEffect(() => {
    if (!user) return
    void reloadFollowees()
    void reloadFollowers()
  }, [reloadFollowers, reloadFollowees, user])

  const current = useMemo(() => {
    const blocked: FollowUser[] = []
    const blockedHint = '后端暂未提供黑名单接口'
    const isBlockedTab = activeTab === 'blocked'

    if (activeTab === 'following') {
      return {
        label: '我关注的',
        icon: <HeartHandshake className="h-4 w-4" />,
        data: followees,
        isLoading: followeesLoading,
        error: followeesError,
        hint: '',
        isReadOnly: false,
      }
    }

    if (activeTab === 'followers') {
      return {
        label: '关注我的',
        icon: <Users className="h-4 w-4" />,
        data: followers,
        isLoading: followersLoading,
        error: followersError,
        hint: '',
        isReadOnly: false,
      }
    }

    return {
      label: '黑名单',
      icon: <ShieldAlert className="h-4 w-4" />,
      data: blocked,
      isLoading: false,
      error: null as string | null,
      hint: isBlockedTab ? blockedHint : '',
      isReadOnly: true,
    }
  }, [activeTab, followees, followeesError, followeesLoading, followers, followersError, followersLoading])

  const tabConfig = useMemo(
    () => ({
      following: { label: '我关注的', count: followees.length, icon: <HeartHandshake className="h-4 w-4" /> },
      followers: { label: '关注我的', count: followers.length, icon: <Users className="h-4 w-4" /> },
      blocked: { label: '黑名单', count: 0, icon: <ShieldAlert className="h-4 w-4" /> },
    }),
    [followees.length, followers.length]
  )

  const handleToggleFollow = useCallback(
    async (target: FollowUser) => {
      if (!user) return
      if (actingId) return

      setActingId(target.id)
      try {
        const currentRelation = relationById[target.id] ?? 0
        const shouldUnfollow = isFollowing(currentRelation) || activeTab === 'following'

        if (shouldUnfollow) {
          await unfollowUser(target.id)
          setFollowees(prev => prev.filter(u => u.id !== target.id))
          const nextRelation = await getFollowRelation(target.id)
          setRelationById(prev => ({ ...prev, [target.id]: nextRelation }))
          return
        }

        await followUser(target.id)
        setFollowees(prev => (prev.some(u => u.id === target.id) ? prev : [target, ...prev]))
        const nextRelation = await getFollowRelation(target.id)
        setRelationById(prev => ({ ...prev, [target.id]: nextRelation }))
      } catch (error) {
        console.error('更新关注关系失败:', error)
        alert(error instanceof Error ? error.message : '更新关注关系失败')
      } finally {
        setActingId(null)
      }
    },
    [actingId, activeTab, relationById, user]
  )

  const handleRetry = useCallback(() => {
    if (activeTab === 'following') {
      void reloadFollowees()
      return
    }
    if (activeTab === 'followers') {
      void reloadFollowers()
    }
  }, [activeTab, reloadFollowers, reloadFollowees])

  if (!user) {
    return null
  }

  return (
    <div className="max-w-6xl mx-auto space-y-6 lg:-ml-1">
      <div className="flex items-center justify-between">
        <Link
          to="/"
          className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
        >
          <ArrowLeft className="h-5 w-5" />
          <span>返回首页</span>
        </Link>
        <div className="text-sm text-gray-500">
          {user.name} · 关注关系
        </div>
      </div>

      <div className="grid md:grid-cols-[220px_1fr] gap-4">
        <div className="card h-fit p-3 space-y-1">
          {Object.entries(tabConfig).map(([key, item]) => {
            const isActive = key === activeTab
            return (
              <button
                key={key}
                onClick={() => setActiveTab(key as TabKey)}
                className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-primary-50 text-primary-700 font-semibold'
                    : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                <span className="flex items-center space-x-2">
                  {item.icon}
                  <span>{item.label}</span>
                </span>
                <span className="text-xs text-gray-500">{item.count}</span>
              </button>
            )
          })}
        </div>

        <div className="card space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              {current.icon}
              <h2 className="text-lg font-semibold text-gray-900">{current.label}</h2>
            </div>
            {current.hint ? (
              <span className="text-xs text-gray-500">{current.hint}</span>
            ) : null}
          </div>
          {current.isLoading ? (
            <div className="text-sm text-gray-500">加载中...</div>
          ) : current.error ? (
            <div className="space-y-2">
              <p className="text-sm text-red-600">{current.error}</p>
              <button type="button" onClick={handleRetry} className="btn-secondary text-sm">
                重试
              </button>
            </div>
          ) : current.data.length === 0 ? (
            <div className="text-sm text-gray-500">暂无数据</div>
          ) : (
            <div className="space-y-3">
              {current.data.map((item) => {
                const relation = relationById[item.id]
                const relationToShow =
                  relation ?? (activeTab === 'following' ? (1 as FollowRelation) : undefined)
                const label = relationToShow !== undefined ? relationLabelMap[relationToShow] : '加载中...'
                const canToggle =
                  !current.isReadOnly && (activeTab === 'following' ? true : relation !== undefined)
                const isActing = actingId === item.id
                const followButtonText = canToggle
                  ? isFollowing(relationToShow ?? 0) || activeTab === 'following'
                    ? '取消关注'
                    : '关注'
                  : '...'

                return (
                  <div
                    key={item.id}
                    className="flex items-center space-x-3 p-2 rounded-lg hover:bg-gray-50 transition-colors"
                  >
                    <Link
                      to={`/users/${item.id}`}
                      state={{ username: item.name }}
                      className="flex-shrink-0"
                    >
                      <img
                        src={item.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=follow-${item.id}`}
                        alt={item.name}
                        className="w-10 h-10 rounded-full"
                      />
                    </Link>
                    <div className="flex-1 min-w-0">
                      <Link
                        to={`/users/${item.id}`}
                        state={{ username: item.name }}
                        className="font-medium text-gray-900 hover:text-primary-600 transition-colors line-clamp-1"
                      >
                        {item.name}
                      </Link>
                      <p className="text-xs text-gray-500 line-clamp-1">{label}</p>
                    </div>
                    {!current.isReadOnly && (
                      <button
                        type="button"
                        disabled={!canToggle || isActing}
                        onClick={() => void handleToggleFollow(item)}
                        className="text-xs text-primary-600 font-medium hover:text-primary-700 disabled:opacity-60 disabled:cursor-not-allowed"
                      >
                        {isActing ? '处理中...' : followButtonText}
                      </button>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
