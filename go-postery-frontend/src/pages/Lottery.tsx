import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  CheckCircle2,
  Clock3,
  Gift,
  RefreshCcw,
  RotateCcw,
  Sparkles,
  Ticket,
  Wallet,
  XCircle,
} from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { ApiError, apiGet, apiPost } from '../utils/api'
import { normalizeId } from '../utils/id'

type GiftItem = {
  id: string
  name: string
  avatar?: string
  description?: string
  prize?: number
}

type WheelItem = GiftItem & {
  color: string
  type?: 'lose'
}

type HistoryItem = {
  id: number
  prize: string
  status: string
  time: string
}

type LotteryOrder = {
  id: string
  user: {
    id: string
    name?: string
    email?: string
    avatar?: string
  }
  gift: GiftItem
  count: number
  created_at?: string
}

const DECISION_SECONDS = 600
const SPIN_TIMEOUT = 4400
const MAX_HISTORY = 8

const WHEEL_COLORS = [
  '#0ea5e9',
  '#14b8a6',
  '#22c55e',
  '#f59e0b',
  '#f97316',
  '#6366f1',
  '#a855f7',
  '#ec4899',
]

const DEFAULT_LOSE_GIFT: GiftItem = {
  id: '0',
  name: '谢谢参与',
  description: '下次再来',
  prize: 0,
}

const normalizeGift = (raw: any): GiftItem | null => {
  if (!raw) return null
  const id = normalizeId(raw.id ?? raw.Id ?? raw.ID)
  const name = typeof raw.name === 'string' ? raw.name.trim() : ''
  if (!id || !name) return null
  const avatar = typeof raw.avatar === 'string' ? raw.avatar.trim() : ''
  const description = typeof raw.description === 'string' ? raw.description.trim() : ''
  const prizeValue = Number(raw.prize)
  const prize = Number.isFinite(prizeValue) ? prizeValue : undefined
  return {
    id,
    name,
    avatar: avatar || undefined,
    description: description || undefined,
    prize,
  }
}

const normalizeOrder = (raw: any): LotteryOrder | null => {
  if (!raw) return null
  const id = normalizeId(raw.id ?? raw.Id ?? raw.ID)
  if (!id) return null
  const userRaw = raw.user ?? raw.User ?? {}
  const userId = normalizeId(userRaw.id ?? userRaw.Id ?? userRaw.ID)
  const userName = typeof userRaw.name === 'string' ? userRaw.name.trim() : ''
  const email = typeof userRaw.email === 'string' ? userRaw.email.trim() : ''
  const avatar = typeof userRaw.avatar === 'string' ? userRaw.avatar.trim() : ''
  const gift = normalizeGift(raw.gift ?? raw.Gift ?? raw.prize ?? raw.gift_info)
  if (!gift) return null
  const countValue = Number(raw.count)
  const count = Number.isFinite(countValue) ? countValue : 1
  const createdAt =
    typeof raw.created_at === 'string'
      ? raw.created_at
      : typeof raw.createdAt === 'string'
        ? raw.createdAt
        : ''
  return {
    id,
    user: {
      id: userId,
      name: userName || undefined,
      email: email || undefined,
      avatar: avatar || undefined,
    },
    gift,
    count,
    created_at: createdAt || undefined,
  }
}

const isLoseGift = (gift: GiftItem | null) => {
  if (!gift) return false
  const normalizedName = gift.name.trim()
  return normalizeId(gift.id) === '0' || normalizedName === '谢谢参与'
}

const formatCountdown = (value: number) => {
  if (!Number.isFinite(value)) return '00:00'
  const total = Math.max(0, Math.floor(value))
  const minutes = Math.floor(total / 60)
  const seconds = total % 60
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
}

const formatPrize = (value?: number) => {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return '—'
  return `¥${numeric}`
}

const formatDateTime = (value?: string) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN')
}

export default function Lottery() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const userId = useMemo(() => normalizeId(user?.id), [user?.id])
  const isLoggedIn = Boolean(userId)

  const [rotation, setRotation] = useState(0)
  const [isSpinning, setIsSpinning] = useState(false)
  const [result, setResult] = useState<GiftItem | null>(null)
  const [decisionState, setDecisionState] = useState<'idle' | 'pending' | 'paid' | 'abandoned' | 'missed'>('idle')
  const [countdown, setCountdown] = useState(0)
  const [history, setHistory] = useState<HistoryItem[]>([])
  const [gifts, setGifts] = useState<GiftItem[]>([])
  const [isGiftsLoading, setIsGiftsLoading] = useState(false)
  const [giftsError, setGiftsError] = useState<string | null>(null)
  const [drawError, setDrawError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState<'pay' | 'giveup' | null>(null)
  const [latestOrder, setLatestOrder] = useState<LotteryOrder | null>(null)
  const [orderStatus, setOrderStatus] = useState<'idle' | 'loading' | 'empty' | 'error' | 'ready'>('idle')
  const [orderError, setOrderError] = useState<string | null>(null)

  const activeRecordId = useRef<number | null>(null)
  const decisionDeadlineRef = useRef<number | null>(null)
  const spinTimerRef = useRef<number | null>(null)
  const countdownTimerRef = useRef<number | null>(null)
  const decisionStateRef = useRef(decisionState)
  const isSubmittingRef = useRef<'pay' | 'giveup' | null>(null)
  const activeGiftRef = useRef<GiftItem | null>(null)

  useEffect(() => {
    decisionStateRef.current = decisionState
  }, [decisionState])

  useEffect(() => {
    isSubmittingRef.current = isSubmitting
  }, [isSubmitting])

  useEffect(() => {
    activeGiftRef.current = result
  }, [result])

  const fetchGifts = useCallback(async () => {
    setIsGiftsLoading(true)
    setGiftsError(null)
    try {
      const { data } = await apiGet<unknown>('/gifts')
      const payload = data as any
      const rawList = Array.isArray(payload)
        ? payload
        : Array.isArray(payload?.gifts)
          ? payload.gifts
          : []
      const normalized = rawList
        .map((item) => normalizeGift(item))
        .filter((item): item is GiftItem => Boolean(item))
      const uniqueMap = new Map<string, GiftItem>()
      normalized.forEach(item => {
        if (!uniqueMap.has(item.id)) {
          uniqueMap.set(item.id, item)
        }
      })
      setGifts(Array.from(uniqueMap.values()))
    } catch (error) {
      console.error('Failed to fetch gifts:', error)
      setGifts([])
      setGiftsError(error instanceof Error ? error.message : '获取奖品失败')
    } finally {
      setIsGiftsLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchGifts()
  }, [fetchGifts])

  const fetchLatestOrder = useCallback(async () => {
    if (!userId) {
      setLatestOrder(null)
      setOrderStatus('idle')
      setOrderError(null)
      return
    }

    setOrderStatus('loading')
    setOrderError(null)
    try {
      const { data } = await apiGet<LotteryOrder>('/lottery/result')
      const normalized = normalizeOrder(data)
      if (normalized) {
        setLatestOrder(normalized)
        setOrderStatus('ready')
      } else {
        setLatestOrder(null)
        setOrderStatus('empty')
      }
    } catch (error) {
      console.error('Failed to fetch lottery result:', error)
      if (error instanceof ApiError && error.status === 200) {
        setLatestOrder(null)
        setOrderStatus('empty')
      } else {
        setLatestOrder(null)
        setOrderStatus('error')
        setOrderError(error instanceof Error ? error.message : '获取中奖结果失败')
      }
    }
  }, [userId])

  useEffect(() => {
    if (!userId) {
      setLatestOrder(null)
      setOrderStatus('idle')
      setOrderError(null)
      return
    }
    void fetchLatestOrder()
  }, [fetchLatestOrder, userId])

  const wheelItems = useMemo(() => {
    const items: WheelItem[] = gifts.map((gift, index) => ({
      ...gift,
      color: WHEEL_COLORS[index % WHEEL_COLORS.length],
      type: isLoseGift(gift) ? 'lose' : undefined,
    }))
    const hasThanks = items.some(item => isLoseGift(item))
    if (!hasThanks) {
      items.push({ ...DEFAULT_LOSE_GIFT, color: '#e5e7eb', type: 'lose' })
    }
    return items
  }, [gifts])

  const wheelCount = wheelItems.length || 1
  const angle = 360 / wheelCount

  const wheelGradient = useMemo(() => {
    if (!wheelItems.length) {
      return '#e5e7eb 0deg 360deg'
    }
    return wheelItems
      .map((item, index) => {
        const start = index * angle
        const end = start + angle
        return `${item.color} ${start}deg ${end}deg`
      })
      .join(', ')
  }, [angle, wheelItems])

  const giftColorById = useMemo(() => {
    return wheelItems.reduce<Record<string, string>>((acc, item) => {
      acc[item.id] = item.color
      return acc
    }, {})
  }, [wheelItems])

  const decisionMinutes = Math.round(DECISION_SECONDS / 60)

  const resetDecisionTimer = useCallback(() => {
    if (countdownTimerRef.current) {
      clearInterval(countdownTimerRef.current)
    }
    countdownTimerRef.current = null
    decisionDeadlineRef.current = null
    setCountdown(0)
  }, [])

  const updateHistoryStatus = useCallback((status: string) => {
    if (!activeRecordId.current) return
    setHistory(prev =>
      prev.map(item => (item.id === activeRecordId.current ? { ...item, status } : item))
    )
  }, [])

  const submitGiveup = useCallback(async (mode: 'manual' | 'timeout') => {
    if (isSubmittingRef.current) return
    const gift = activeGiftRef.current
    if (!gift || isLoseGift(gift)) return
    if (decisionStateRef.current !== 'pending') return
    if (!userId) {
      if (mode === 'manual') {
        alert('请先登录后再操作')
        navigate('/login')
      }
      return
    }

    setIsSubmitting('giveup')
    setActionError(null)

    if (mode === 'timeout') {
      setDecisionState('abandoned')
      updateHistoryStatus('超时自动放弃')
    }

    try {
      await apiPost('/lottery/giveup', { user_id: userId, gift_id: gift.id })
      if (mode === 'manual') {
        setDecisionState('abandoned')
        resetDecisionTimer()
        updateHistoryStatus('已主动放弃')
      }
    } catch (error) {
      console.error('放弃支付失败:', error)
      const message = error instanceof Error ? error.message : '放弃支付失败'
      setActionError(message)
    } finally {
      setIsSubmitting(null)
    }
  }, [navigate, resetDecisionTimer, updateHistoryStatus, userId])

  const startDecisionWindow = useCallback((gift: GiftItem) => {
    if (isLoseGift(gift)) {
      setDecisionState('missed')
      resetDecisionTimer()
      updateHistoryStatus('未中奖')
      return
    }

    setDecisionState('pending')
    setCountdown(DECISION_SECONDS)
    decisionDeadlineRef.current = Date.now() + DECISION_SECONDS * 1000

    countdownTimerRef.current = window.setInterval(() => {
      if (!decisionDeadlineRef.current) return
      const remaining = Math.max(
        0,
        Math.ceil((decisionDeadlineRef.current - Date.now()) / 1000)
      )
      setCountdown(remaining)
      if (remaining <= 0) {
        if (isSubmittingRef.current) return
        resetDecisionTimer()
        void submitGiveup('timeout')
      }
    }, 250)
  }, [resetDecisionTimer, submitGiveup, updateHistoryStatus])

  const findWheelIndex = useCallback((gift: GiftItem) => {
    if (!wheelItems.length) return 0
    const targetId = normalizeId(gift.id)
    let index = wheelItems.findIndex(item => normalizeId(item.id) === targetId)
    if (index < 0 && gift.name) {
      index = wheelItems.findIndex(item => item.name === gift.name)
    }
    if (index < 0) {
      index = Math.floor(Math.random() * wheelItems.length)
    }
    return index
  }, [wheelItems])

  const handleSpin = async () => {
    if (isSpinning || isGiftsLoading || gifts.length === 0) return
    if (!userId) {
      alert('请先登录后再抽奖')
      navigate('/login')
      return
    }

    if (spinTimerRef.current) {
      clearTimeout(spinTimerRef.current)
    }

    setIsSpinning(true)
    setDecisionState('idle')
    setResult(null)
    setDrawError(null)
    setActionError(null)
    resetDecisionTimer()

    try {
      const { data } = await apiGet<GiftItem>('/lottery/lucky')
      const rawGift = (data as any)?.gift ?? data
      const gift = normalizeGift(rawGift)
      if (!gift) {
        throw new Error('抽奖失败，请稍后重试')
      }

      const targetIndex = findWheelIndex(gift)
      const extraTurns = 6 + Math.floor(Math.random() * 3)
      const targetAngle = 360 - targetIndex * angle - angle / 2

      setRotation(prev => {
        const current = ((prev % 360) + 360) % 360
        const normalizedTarget = ((targetAngle % 360) + 360) % 360
        const delta = extraTurns * 360 + ((normalizedTarget - current + 360) % 360)
        return prev + delta
      })

      spinTimerRef.current = window.setTimeout(() => {
        setResult(gift)
        setIsSpinning(false)
        const recordId = Date.now()
        activeRecordId.current = recordId
        setHistory(prev => [
          {
            id: recordId,
            prize: gift.name,
            status: isLoseGift(gift) ? '未中奖' : '等待确认',
            time: new Date().toLocaleTimeString(),
          },
          ...prev,
        ].slice(0, MAX_HISTORY))

        startDecisionWindow(gift)
      }, SPIN_TIMEOUT)
    } catch (error) {
      console.error('Failed to draw lottery:', error)
      setIsSpinning(false)
      setDrawError(error instanceof Error ? error.message : '抽奖失败，请稍后重试')
    }
  }

  const handlePay = async () => {
    if (decisionState !== 'pending' || isSubmitting) return
    if (!result || isLoseGift(result)) return
    if (!userId) {
      alert('请先登录后再操作')
      navigate('/login')
      return
    }

    setIsSubmitting('pay')
    setActionError(null)

    try {
      await apiPost('/lottery/pay', { user_id: userId, gift_id: result.id })
      setDecisionState('paid')
      resetDecisionTimer()
      updateHistoryStatus('已支付领取')
      void fetchLatestOrder()
    } catch (error) {
      console.error('支付失败:', error)
      setActionError(error instanceof Error ? error.message : '支付失败，请稍后重试')
    } finally {
      setIsSubmitting(null)
    }
  }

  const handleReset = () => {
    resetDecisionTimer()
    setDecisionState('idle')
    setResult(null)
    setDrawError(null)
    setActionError(null)
    activeRecordId.current = null
  }

  useEffect(() => {
    return () => {
      if (spinTimerRef.current) clearTimeout(spinTimerRef.current)
      resetDecisionTimer()
    }
  }, [resetDecisionTimer])

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      <div className="card bg-gradient-to-r from-primary-50 via-white to-white border-primary-100/80 overflow-hidden relative">
        <div className="absolute -right-16 -top-16 w-56 h-56 bg-primary-100/70 rounded-full blur-3xl" />
        <div className="absolute -left-10 bottom-0 w-40 h-40 bg-white/70 border border-primary-50 rounded-full blur-3xl" />
        <div className="relative flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div className="space-y-2">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white border border-primary-100 text-primary-700 text-sm">
              <Sparkles className="h-4 w-4" />
              今日幸运轮盘
            </div>
            <h1 className="text-2xl lg:text-3xl font-bold text-gray-900">抽出今天的专属福利</h1>
            <p className="text-gray-600 max-w-2xl">
              奖品池来自后台奖品列表，中奖后需在 {decisionMinutes} 分钟内完成支付或放弃，超时系统将自动释放订单。
            </p>
            <div className="flex flex-wrap gap-3 text-sm text-gray-600">
              <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white border border-gray-200">
                <Clock3 className="h-4 w-4 text-primary-600" />
                {decisionMinutes} 分钟支付时限
              </span>
              <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white border border-gray-200">
                <Wallet className="h-4 w-4 text-primary-600" />
                中奖后需确认支付
              </span>
              <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white border border-gray-200">
                <Ticket className="h-4 w-4 text-primary-600" />
                登录后即可抽奖
              </span>
            </div>
          </div>
          <div className="flex flex-col items-start gap-2">
            <div className="flex gap-2 items-center">
              <button
                type="button"
                onClick={handleSpin}
                disabled={isSpinning || isGiftsLoading || gifts.length === 0}
                className="btn-primary flex items-center gap-2 shadow-md disabled:cursor-not-allowed"
              >
                <Sparkles className="h-5 w-5" />
                {isSpinning ? '抽奖中...' : '启动抽奖'}
              </button>
              <button
                type="button"
                onClick={handleReset}
                className="btn-secondary flex items-center gap-2"
              >
                <RotateCcw className="h-4 w-4" />
                重置提示
              </button>
            </div>
            {!isLoggedIn && (
              <div className="text-xs text-orange-600 bg-orange-50 border border-orange-100 px-3 py-1.5 rounded-full">
                请先登录后参与抽奖
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="grid lg:grid-cols-[1.1fr_0.9fr] gap-6">
        <div className="space-y-6">
          <div className="card relative overflow-hidden">
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(14,165,233,0.08),transparent_45%),radial-gradient(circle_at_80%_10%,rgba(99,102,241,0.08),transparent_40%)]" />
            <div className="relative flex flex-col items-center gap-6">
              <div className="text-center space-y-1">
                <p className="text-gray-500 text-sm">
                  {isGiftsLoading ? '奖品加载中...' : giftsError ? '奖品加载失败' : `奖品池共 ${gifts.length} 款`}
                </p>
                <p className="text-lg font-semibold text-gray-900">
                  {isSpinning ? '正在抽取...' : result ? result.name : '点击上方按钮开始抽奖'}
                </p>
              </div>

              <div className="relative w-full max-w-[480px] aspect-square">
                <div
                  className="absolute inset-0 rounded-full shadow-lg border-8 border-white transition-transform duration-[4200ms] ease-out"
                  style={{
                    background: `conic-gradient(${wheelGradient})`,
                    transform: `rotate(${rotation}deg)`,
                  }}
                >
                  {wheelItems.map((item, index) => {
                    const prizeRotate = angle * index + angle / 2
                    return (
                      <div
                        key={`${item.id}-${item.name}`}
                        className="absolute inset-0 flex items-start"
                        style={{ transform: `rotate(${prizeRotate}deg)` }}
                      >
                        <div
                          className="origin-center translate-y-2 flex flex-col items-center text-center w-1/2 ml-auto mr-auto"
                          style={{ transform: `rotate(${-prizeRotate}deg)` }}
                        >
                          <span className="text-xs font-semibold text-gray-900 drop-shadow-sm">
                            {item.name}
                          </span>
                          <span className="text-[11px] text-gray-700 opacity-80">
                            {item.description || (item.type === 'lose' ? '下次好运' : '惊喜好礼')}
                          </span>
                        </div>
                      </div>
                    )
                  })}
                </div>

                <div className="absolute left-1/2 -top-3 -translate-x-1/2 z-20">
                  <div className="w-0 h-0 border-l-6 border-r-6 border-b-10 border-l-transparent border-r-transparent border-b-primary-600 drop-shadow" />
                  <div className="w-0 h-0 border-l-4 border-r-4 border-b-8 border-l-transparent border-r-transparent border-b-primary-300 -mt-0.5 ml-1" />
                </div>

                <div className="absolute inset-[18%] rounded-full bg-white shadow-lg flex flex-col items-center justify-center text-center px-6">
                  <div className="text-xs text-gray-500">奖品池</div>
                  <div className="text-3xl font-bold text-gray-900 flex items-baseline gap-1">
                    {isGiftsLoading ? '...' : gifts.length.toLocaleString()}
                    <span className="text-sm text-gray-500">款</span>
                  </div>
                  <p className="text-[13px] text-gray-500 mt-1">
                    {giftsError ? '奖品列表加载失败' : '以后台奖品列表为准'}
                  </p>
                </div>
              </div>

              <div className="grid gap-3 w-full">
                <div className="rounded-xl bg-primary-50 border border-primary-100 px-4 py-3 text-sm text-primary-800">
                  中奖需在 {decisionMinutes} 分钟内选择支付或放弃，超时自动释放
                </div>
                {giftsError && (
                  <div className="rounded-xl bg-red-50 border border-red-100 px-4 py-3 text-xs text-red-700 flex items-center justify-between">
                    <span>奖品加载失败：{giftsError}</span>
                    <button
                      type="button"
                      onClick={fetchGifts}
                      className="text-red-700 hover:text-red-600 font-semibold"
                    >
                      重新加载
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="card space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Gift className="h-5 w-5 text-primary-600" />
                <h3 className="text-lg font-semibold text-gray-900">奖品列表</h3>
              </div>
              <button
                type="button"
                onClick={fetchGifts}
                disabled={isGiftsLoading}
                className="btn-secondary !py-1.5 !px-3 text-xs flex items-center gap-1"
              >
                <RefreshCcw className="h-3.5 w-3.5" />
                刷新
              </button>
            </div>

            {isGiftsLoading ? (
              <p className="text-sm text-gray-600">奖品加载中...</p>
            ) : gifts.length === 0 ? (
              <p className="text-sm text-gray-600">暂无可展示奖品</p>
            ) : (
              <div className="grid sm:grid-cols-2 gap-3">
                {gifts.map((gift, index) => {
                  const accent = WHEEL_COLORS[index % WHEEL_COLORS.length]
                  return (
                    <div
                      key={gift.id}
                      className="flex items-start gap-3 p-3 rounded-lg border border-gray-200 bg-white"
                    >
                      {gift.avatar ? (
                        <img
                          src={gift.avatar}
                          alt={gift.name}
                          className="w-10 h-10 rounded-full object-cover"
                          loading="lazy"
                        />
                      ) : (
                        <div
                          className="w-10 h-10 rounded-full flex items-center justify-center text-white text-xs font-semibold"
                          style={{ backgroundColor: accent }}
                        >
                          {gift.name.slice(0, 2)}
                        </div>
                      )}
                      <div className="flex-1">
                        <div className="flex items-center justify-between gap-2">
                          <p className="text-sm font-semibold text-gray-900">{gift.name}</p>
                          <span className="text-xs text-primary-700 bg-primary-50 border border-primary-100 rounded-full px-2 py-0.5">
                            {formatPrize(gift.prize)}
                          </span>
                        </div>
                        <p className="text-xs text-gray-600 mt-1">
                          {gift.description || '暂无描述'}
                        </p>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}

            {giftsError && !isGiftsLoading && (
              <p className="text-xs text-red-600">{giftsError}</p>
            )}
          </div>
        </div>

        <div className="space-y-4">
          <div className="card space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Wallet className="h-5 w-5 text-primary-600" />
                <h2 className="text-lg font-semibold text-gray-900">支付或放弃</h2>
              </div>
              {decisionState === 'pending' && (
                <div className="flex items-center gap-2 text-xs text-primary-700 bg-primary-50 border border-primary-100 px-3 py-1.5 rounded-full">
                  <Clock3 className="h-4 w-4" />
                  剩余 {formatCountdown(countdown)}
                </div>
              )}
            </div>

            {!result && (
              <div className="text-sm text-gray-600">
                {isLoggedIn
                  ? '先启动抽奖，抽中后会在这里展示奖品并进入支付/放弃倒计时。'
                  : '登录后即可参与抽奖，中奖结果会显示在这里。'}
              </div>
            )}

            {drawError && (
              <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 border border-red-100 px-3 py-2 rounded-lg">
                <XCircle className="h-4 w-4" />
                {drawError}
              </div>
            )}

            {result && (
              <div className="space-y-3">
                <div className="flex items-start gap-3 p-3 rounded-lg bg-gray-50 border border-gray-200">
                  {result.avatar ? (
                    <img
                      src={result.avatar}
                      alt={result.name}
                      className="w-10 h-10 rounded-full object-cover"
                      loading="lazy"
                    />
                  ) : (
                    <div
                      className="w-10 h-10 rounded-full flex items-center justify-center text-white font-semibold"
                      style={{ backgroundColor: giftColorById[result.id] || '#94a3b8' }}
                    >
                      {result.name.slice(0, 2)}
                    </div>
                  )}
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <p className="font-semibold text-gray-900">{result.name}</p>
                      {isLoseGift(result) && (
                        <span className="text-xs text-gray-500 bg-white border border-gray-200 rounded-full px-2 py-0.5">未中奖</span>
                      )}
                    </div>
                    {result.description && (
                      <p className="text-sm text-gray-600">{result.description}</p>
                    )}
                    {!isLoseGift(result) && (
                      <p className="text-sm text-gray-700 mt-1">
                        奖品价值 <span className="font-semibold text-primary-700">{formatPrize(result.prize)}</span>
                      </p>
                    )}
                  </div>
                </div>

                {isLoseGift(result) && (
                  <div className="flex items-center gap-2 text-sm text-gray-600">
                    <XCircle className="h-4 w-4 text-gray-400" />
                    很遗憾未中奖，欢迎再试一次。
                  </div>
                )}

                {!isLoseGift(result) && (
                  <div className="flex flex-col gap-3">
                    <div className="flex items-center gap-2 text-sm">
                      <Clock3 className="h-4 w-4 text-primary-600" />
                      <span className="text-gray-700">
                        请在倒计时结束前确认支付或放弃。
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <button
                        type="button"
                        onClick={handlePay}
                        disabled={decisionState !== 'pending' || Boolean(isSubmitting)}
                        className="btn-primary flex-1 flex items-center justify-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
                      >
                        <Wallet className="h-4 w-4" />
                        {isSubmitting === 'pay' ? '支付中...' : '支付并领取'}
                      </button>
                      <button
                        type="button"
                        onClick={() => void submitGiveup('manual')}
                        disabled={decisionState !== 'pending' || Boolean(isSubmitting)}
                        className="btn-secondary flex-1 flex items-center justify-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
                      >
                        <XCircle className="h-4 w-4" />
                        {isSubmitting === 'giveup' ? '放弃中...' : '放弃'}
                      </button>
                    </div>
                    {actionError && (
                      <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 border border-red-100 px-3 py-2 rounded-lg">
                        <XCircle className="h-4 w-4" />
                        {actionError}
                      </div>
                    )}
                    {decisionState === 'paid' && (
                      <div className="flex items-center gap-2 text-sm text-green-600 bg-green-50 border border-green-100 px-3 py-2 rounded-lg">
                        <CheckCircle2 className="h-4 w-4" />
                        支付完成，奖品已锁定。
                      </div>
                    )}
                    {decisionState === 'abandoned' && (
                      <div className="flex items-center gap-2 text-sm text-orange-600 bg-orange-50 border border-orange-100 px-3 py-2 rounded-lg">
                        <XCircle className="h-4 w-4" />
                        奖品已放弃，可继续抽取其他奖励。
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="card space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="h-5 w-5 text-primary-600" />
                <h3 className="text-lg font-semibold text-gray-900">最新中奖结果</h3>
              </div>
              {isLoggedIn && (
                <button
                  type="button"
                  onClick={() => void fetchLatestOrder()}
                  disabled={orderStatus === 'loading'}
                  className="btn-secondary !py-1.5 !px-3 text-xs flex items-center gap-1"
                >
                  <RefreshCcw className="h-3.5 w-3.5" />
                  刷新
                </button>
              )}
            </div>

            {!isLoggedIn && (
              <p className="text-sm text-gray-600">登录后查看支付成功的中奖结果。</p>
            )}

            {isLoggedIn && orderStatus === 'loading' && (
              <p className="text-sm text-gray-600">正在加载中奖结果...</p>
            )}

            {isLoggedIn && orderStatus === 'empty' && (
              <p className="text-sm text-gray-600">暂无中奖记录，支付成功后会展示在这里。</p>
            )}

            {isLoggedIn && orderStatus === 'error' && (
              <div className="text-sm text-red-600 bg-red-50 border border-red-100 px-3 py-2 rounded-lg">
                {orderError || '获取中奖结果失败'}
              </div>
            )}

            {isLoggedIn && orderStatus === 'ready' && latestOrder && (
              <div className="flex items-start gap-3 p-3 rounded-lg bg-gray-50 border border-gray-200">
                {latestOrder.gift.avatar ? (
                  <img
                    src={latestOrder.gift.avatar}
                    alt={latestOrder.gift.name}
                    className="w-10 h-10 rounded-full object-cover"
                    loading="lazy"
                  />
                ) : (
                  <div
                    className="w-10 h-10 rounded-full flex items-center justify-center text-white font-semibold"
                    style={{ backgroundColor: giftColorById[latestOrder.gift.id] || '#94a3b8' }}
                  >
                    {latestOrder.gift.name.slice(0, 2)}
                  </div>
                )}
                <div className="flex-1 space-y-1">
                  <div className="flex items-center justify-between gap-2">
                    <p className="font-semibold text-gray-900">{latestOrder.gift.name}</p>
                    <span className="text-xs text-gray-500">订单 #{latestOrder.id}</span>
                  </div>
                  {latestOrder.gift.description && (
                    <p className="text-sm text-gray-600">{latestOrder.gift.description}</p>
                  )}
                  <div className="flex flex-wrap gap-3 text-xs text-gray-500">
                    <span>数量 {latestOrder.count}</span>
                    {latestOrder.created_at && (
                      <span>支付时间 {formatDateTime(latestOrder.created_at)}</span>
                    )}
                    <span>价值 {formatPrize(latestOrder.gift.prize)}</span>
                  </div>
                </div>
              </div>
            )}
          </div>

          <div className="card">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <Clock3 className="h-5 w-5 text-primary-600" />
                <h3 className="text-lg font-semibold text-gray-900">抽奖记录</h3>
              </div>
            </div>
            {history.length === 0 ? (
              <p className="text-sm text-gray-600">暂无记录，开始抽奖吧。</p>
            ) : (
              <ul className="space-y-2">
                {history.map(item => (
                  <li key={item.id} className="flex items-center justify-between rounded-lg border border-gray-200 px-3 py-2 bg-white">
                    <div>
                      <p className="text-sm font-semibold text-gray-900">{item.prize}</p>
                      <p className="text-xs text-gray-500">{item.time}</p>
                    </div>
                    <span className="text-xs px-3 py-1 rounded-full border bg-gray-50 text-gray-700">
                      {item.status}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="card space-y-2">
            <div className="flex items-center gap-2">
              <Sparkles className="h-5 w-5 text-primary-600" />
              <h3 className="text-lg font-semibold text-gray-900">规则说明</h3>
            </div>
            <ul className="list-disc list-inside space-y-1 text-sm text-gray-700">
              <li>奖品列表实时同步后台接口，库存以系统数据为准。</li>
              <li>抽奖需登录，中奖后请在 {decisionMinutes} 分钟内完成支付或放弃。</li>
              <li>抽到「谢谢参与」或奖品已抽完不会生成订单。</li>
              <li>支付成功后可在「最新中奖结果」查看记录。</li>
              <li>抽奖记录仅保留本地最近 {MAX_HISTORY} 条。</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
