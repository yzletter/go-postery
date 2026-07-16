import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Users, HeartHandshake } from 'lucide-react'
import UserAvatar from '../components/UserAvatar'
import { useAuth } from '../contexts/AuthContext'
import type { FollowRelation, FollowUser } from '../types'
import { followUser, getFollowRelation, isFollowing, listFollowers, listFollowees, unfollowUser } from '../utils/follow'

type TabKey = 'following' | 'followers'

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
  const [followeesTotal, setFolloweesTotal] = useState<number | null>(null)
  const [followersTotal, setFollowersTotal] = useState<number | null>(null)
  const [followeesHasMore, setFolloweesHasMore] = useState(false)
  const [followersHasMore, setFollowersHasMore] = useState(false)
  const [followeesLoading, setFolloweesLoading] = useState(false)
  const [followersLoading, setFollowersLoading] = useState(false)
  const [followeesError, setFolloweesError] = useState<string | null>(null)
  const [followersError, setFollowersError] = useState<string | null>(null)
  const [relationById, setRelationById] = useState<Record<string, FollowRelation>>({})
  const [relationErrorById, setRelationErrorById] = useState<Record<string, boolean>>({})
  const [actingId, setActingId] = useState<string | null>(null)

  const reloadFollowees = useCallback(async () => {
    setFolloweesLoading(true)
    setFolloweesError(null)
    try {
      const { users, total, hasMore } = await listFollowees()
      setFollowees(users)
      setFolloweesTotal(total)
      setFolloweesHasMore(hasMore)
    } catch (error) {
      setFollowees([])
      setFolloweesTotal(null)
      setFolloweesHasMore(false)
      setFolloweesError(error instanceof Error ? error.message : '获取关注列表失败')
    } finally {
      setFolloweesLoading(false)
    }
  }, [])

  const reloadFollowers = useCallback(async () => {
    setFollowersLoading(true)
    setFollowersError(null)
    try {
      const { users, total, hasMore } = await listFollowers()
      setFollowers(users)
      setFollowersTotal(total)
      setFollowersHasMore(hasMore)
    } catch (error) {
      setFollowers([])
      setFollowersTotal(null)
      setFollowersHasMore(false)
      setFollowersError(error instanceof Error ? error.message : '获取粉丝列表失败')
    } finally {
      setFollowersLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!user) return
    void reloadFollowees()
    void reloadFollowers()
  }, [reloadFollowers, reloadFollowees, user])

  // 两个列表本身已经能推导大多数关系，避免为每个用户额外请求一次关系接口。
  useEffect(() => {
    const followeeIds = new Set(followees.map((item) => item.id))
    const followerIds = new Set(followers.map((item) => item.id))
    const nextRelations: Record<string, FollowRelation> = {}
    const nextUnknown: Record<string, boolean> = {}

    followees.forEach((item) => {
      nextRelations[item.id] = followerIds.has(item.id) ? 3 : 1
    })

    const followeesIncomplete =
      followeesLoading || Boolean(followeesError) || followeesHasMore
    followers.forEach((item) => {
      if (followeeIds.has(item.id)) {
        nextRelations[item.id] = 3
      } else if (followeesIncomplete) {
        nextUnknown[item.id] = true
      } else {
        nextRelations[item.id] = 2
      }
    })

    setRelationById(nextRelations)
    setRelationErrorById(nextUnknown)
  }, [
    followees,
    followeesError,
    followeesHasMore,
    followeesLoading,
    followers,
  ])

  const current = useMemo(() => {
    if (activeTab === 'following') {
      return {
        label: '我关注的',
        icon: <HeartHandshake className="h-4 w-4" />,
        data: followees,
        isLoading: followeesLoading,
        error: followeesError,
        hasMore: followeesHasMore,
        total: followeesTotal,
      }
    }

    return {
      label: '关注我的',
      icon: <Users className="h-4 w-4" />,
      data: followers,
      isLoading: followersLoading,
      error: followersError,
      hasMore: followersHasMore,
      total: followersTotal,
    }
  }, [
    activeTab,
    followees,
    followeesError,
    followeesHasMore,
    followeesLoading,
    followeesTotal,
    followers,
    followersError,
    followersHasMore,
    followersLoading,
    followersTotal,
  ])

  const tabConfig = useMemo(
    () => ({
      following: { label: '我关注的', count: followeesTotal ?? '—', icon: <HeartHandshake className="h-4 w-4" /> },
      followers: { label: '关注我的', count: followersTotal ?? '—', icon: <Users className="h-4 w-4" /> },
    }),
    [followeesTotal, followersTotal]
  )

  const handleRetryRelation = useCallback(async (target: FollowUser) => {
    if (actingId) return
    setActingId(target.id)
    try {
      const relation = await getFollowRelation(target.id)
      setRelationById(prev => ({ ...prev, [target.id]: relation }))
      setRelationErrorById(prev => {
        const next = { ...prev }
        delete next[target.id]
        return next
      })
    } catch (error) {
      alert(error instanceof Error ? error.message : '获取关注关系失败')
    } finally {
      setActingId(null)
    }
  }, [actingId])

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
          setFolloweesTotal(prev => (prev == null ? prev : Math.max(0, prev - 1)))
          setRelationById(prev => ({
            ...prev,
            [target.id]: currentRelation === 3 ? 2 : 0,
          }))
          return
        }

        await followUser(target.id)
        if (!followees.some(u => u.id === target.id)) {
          setFollowees(prev => [target, ...prev])
          setFolloweesTotal(prev => (prev == null ? prev : prev + 1))
        }
        setRelationById(prev => ({
          ...prev,
          [target.id]: currentRelation === 2 ? 3 : 1,
        }))
      } catch (error) {
        console.error('更新关注关系失败:', error)
        alert(error instanceof Error ? error.message : '更新关注关系失败')
      } finally {
        setActingId(null)
      }
    },
    [actingId, activeTab, followees, relationById, user]
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
            {current.hasMore ? (
              <span className="text-xs text-amber-700">
                当前展示前 {current.data.length} / {current.total ?? '更多'} 人
              </span>
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
                const relationFailed = Boolean(relationErrorById[item.id])
                const label = relationFailed
                  ? '关注状态需单独确认'
                  : relationToShow !== undefined
                    ? relationLabelMap[relationToShow]
                    : '加载中...'
                const canToggle =
                  !relationFailed && (activeTab === 'following' ? true : relation !== undefined)
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
                      <UserAvatar
                        avatar={item.avatar}
                        name={item.name}
                        userId={item.id}
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
                    {relationFailed ? (
                      <button
                        type="button"
                        disabled={isActing}
                        onClick={() => void handleRetryRelation(item)}
                        className="text-xs text-primary-600 font-medium hover:text-primary-700 disabled:opacity-60 disabled:cursor-not-allowed"
                      >
                        {isActing ? '查询中...' : '查询状态'}
                      </button>
                    ) : (
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
