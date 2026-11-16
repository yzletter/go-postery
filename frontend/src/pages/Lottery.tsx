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
    avatar?: string
  }
  gift: GiftItem
  count: number
  created_at?: string
  status?: number
  paid_at?: string
  expire_at?: string
}

type LotteryDrawStatus = 'won' | 'pending_order' | 'missed' | 'sold_out' | 'unknown'

type LotteryDrawResult = {
  gift: GiftItem
  tempOrderId: string
  success?: boolean
  description?: string
  userId?: string
  status: LotteryDrawStatus
}

const DECISION_SECONDS = 600
const SPIN_TIMEOUT = 4400
const MAX_HISTORY = 8
const ORDER_STATUS_PENDING = 0
const ORDER_STATUS_PAID = 1
const ORDER_STATUS_CANCELLED = 2
const ORDER_STATUS_EXPIRED = 3
const ORDER_NOT_AVAILABLE_CODE = 40002
const ORDER_NOT_FOUND_CODE = 40003
const LOTTERY_ENDED_CODE = 40005
const LOTTERY_ALREADY_PAID_CODE = 40006
const RESULT_RETRY_INTERVAL = 2000
const MAX_RESULT_RETRIES = 10

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

const RESULT_DESCRIPTION_SUCCESS = '抽奖成功'
const RESULT_DESCRIPTION_NO_GIFTS = '奖品已抽完'
const RESULT_DESCRIPTION_MISS = '很遗憾，未抽中奖品，谢谢参与'
const RESULT_DESCRIPTION_TEMP_ORDER = '当前已有订单'

const readString = (value: unknown) => (typeof value === 'string' ? value.trim() : '')

const readBoolean = (value: unknown): boolean | undefined => {
  if (typeof value === 'boolean') return value
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase()
    if (normalized === 'true') return true
    if (normalized === 'false') return false
  }
  return undefined
}

const normalizeGift = (raw: any): GiftItem | null => {
  if (!raw) return null
  const id = normalizeId(raw.id ?? raw.Id ?? raw.ID ?? raw.gift_id ?? raw.GiftID)
  const name = readString(raw.name ?? raw.Name)
  if (!id || !name) return null
  const avatar = readString(raw.avatar ?? raw.Avatar)
  const description = readString(raw.description ?? raw.Description)
  const prizeValue = Number(raw.prize ?? raw.Prize)
  const prize = Number.isFinite(prizeValue) ? prizeValue : undefined
  return {
    id,
    name,
    avatar: avatar || undefined,
    description: description || undefined,
    prize,
  }
}

const isLoseGift = (gift: GiftItem | null) => {
  if (!gift) return false
  const normalizedName = gift.name.trim()
  return normalizeId(gift.id) === '0' || normalizedName === '谢谢参与'
}

const fallbackGiftFromDescription = (description: string): GiftItem | null => {
  if (!description) return null
  if (description.includes(RESULT_DESCRIPTION_NO_GIFTS)) {
    return {
      ...DEFAULT_LOSE_GIFT,
      name: RESULT_DESCRIPTION_NO_GIFTS,
      description,
    }
  }
  if (description.includes('未抽中') || description.includes('谢谢参与')) {
    return {
      ...DEFAULT_LOSE_GIFT,
      description,
    }
  }
  return null
}

const normalizeOrder = (raw: any): LotteryOrder | null => {
  if (!raw) return null
  const id = normalizeId(raw.id ?? raw.Id ?? raw.ID ?? raw.order_id ?? raw.OrderID)
  if (!id) return null
  const userRaw = raw.user ?? raw.User ?? {}
  const userId = normalizeId(userRaw.id ?? userRaw.Id ?? userRaw.ID)
  const userName =
    typeof userRaw.nickname === 'string'
      ? userRaw.nickname.trim()
      : typeof userRaw.Nickname === 'string'
        ? userRaw.Nickname.trim()
        : typeof userRaw.name === 'string'
          ? userRaw.name.trim()
          : ''
  const avatar = typeof userRaw.avatar === 'string' ? userRaw.avatar.trim() : ''
  const gift = normalizeGift(raw.gift ?? raw.Gift ?? raw.prize ?? raw.gift_info)
  if (!gift) return null
  const countValue = Number(raw.count ?? raw.Count)
  const count = Number.isFinite(countValue) ? countValue : 1
  const createdAt =
    typeof raw.created_at === 'string'
      ? raw.created_at
      : typeof raw.createdAt === 'string'
        ? raw.createdAt
        : typeof raw.CreatedAt === 'string'
          ? raw.CreatedAt
          : ''
  const paidAt =
    typeof raw.paid_at === 'string'
      ? raw.paid_at
      : typeof raw.paidAt === 'string'
        ? raw.paidAt
        : typeof raw.PaidAt === 'string'
          ? raw.PaidAt
          : ''
  const expireAt =
    typeof raw.expire_at === 'string'
      ? raw.expire_at
      : typeof raw.expireAt === 'string'
        ? raw.expireAt
        : typeof raw.ExpireAt === 'string'
          ? raw.ExpireAt
          : ''
  const statusValue = Number(raw.status ?? raw.Status)
  const status = Number.isFinite(statusValue) ? statusValue : undefined

  return {
    id,
    user: {
      id: userId,
      name: userName || undefined,
      avatar: avatar || undefined,
    },
    gift,
    count,
    created_at: createdAt || undefined,
    status,
    paid_at: paidAt || undefined,
    expire_at: expireAt || undefined,
  }
}

const getDrawStatus = (
  gift: GiftItem,
  tempOrderId: string,
  description: string
): LotteryDrawStatus => {
  if (description.includes(RESULT_DESCRIPTION_NO_GIFTS) || gift.name === RESULT_DESCRIPTION_NO_GIFTS) {
    return 'sold_out'
  }
  if (tempOrderId && description.includes(RESULT_DESCRIPTION_TEMP_ORDER) && !isLoseGift(gift)) {
    return 'pending_order'
  }
  if (description.includes('未抽中') || isLoseGift(gift)) {
    return 'missed'
  }
  if (tempOrderId && !isLoseGift(gift)) {
    return 'won'
  }
  return 'unknown'
}

const normalizeDrawResult = (raw: any): LotteryDrawResult | null => {
  if (!raw) return null
  const responseGift = raw.gift ?? raw.Gift
  const description = readString(
    raw.result_description ??
    raw.resultDescription ??
    raw.ResultDescription ??
    (responseGift ? raw.description ?? raw.Description : undefined)
  )
  const gift = normalizeGift(responseGift ?? raw) ?? fallbackGiftFromDescription(description)
  if (!gift) return null
  const tempOrderId = normalizeId(
    raw.temp_order_id ??
    raw.tempOrderID ??
    raw.TempOrderID ??
    raw.tempOrderId
  )
  const userId = normalizeId(raw.user_id ?? raw.userId ?? raw.UserID)
  return {
    gift,
    tempOrderId,
    success: readBoolean(raw.success ?? raw.Success),
    description: description || undefined,
    userId: userId || undefined,
    status: getDrawStatus(gift, tempOrderId, description),
  }
}

const isPayableDrawResult = (drawResult: LotteryDrawResult) => {
  return Boolean(drawResult.tempOrderId) && !isLoseGift(drawResult.gift)
}

const getDrawNotice = (drawResult: LotteryDrawResult) => {
  if (drawResult.status === 'pending_order') {
    return '检测到已有待处理订单，请继续确认领取或放弃。'
  }
  if (drawResult.status === 'sold_out') {
    return drawResult.description || RESULT_DESCRIPTION_NO_GIFTS
  }
  if (drawResult.status === 'missed') {
    return drawResult.description && drawResult.description !== RESULT_DESCRIPTION_MISS
      ? drawResult.description
      : null
  }
  if (drawResult.description && drawResult.description !== RESULT_DESCRIPTION_SUCCESS) {
    return drawResult.description
  }
  return null
}

const getHistoryStatus = (drawResult: LotteryDrawResult) => {
  switch (drawResult.status) {
    case 'pending_order':
      return '已有待确认订单'
    case 'missed':
      return '未中奖'
    case 'sold_out':
      return RESULT_DESCRIPTION_NO_GIFTS
    default:
      return isPayableDrawResult(drawResult) ? '等待确认' : '未中奖'
  }
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

const getOrderDeadline = (order: LotteryOrder): number | null => {
  if (order.expire_at) {
    const expireAt = new Date(order.expire_at).getTime()
    if (Number.isFinite(expireAt)) return expireAt
  }

  return null
}

const getOrderStatusText = (order: LotteryOrder) => {
  switch (order.status) {
    case ORDER_STATUS_PAID:
      return '已确认领取'
    case ORDER_STATUS_CANCELLED:
      return '已放弃'
    case ORDER_STATUS_EXPIRED:
      return '已超时'
    case ORDER_STATUS_PENDING:
      return (getOrderDeadline(order) ?? Infinity) <= Date.now() ? '已超时' : '待确认领取'
    default:
      return '状态待确认'
  }
}

const requestLatestOrder = async (signal?: AbortSignal): Promise<LotteryOrder> => {
  const { data } = await apiGet<LotteryOrder>('/lottery/result', { signal })
  const order = normalizeOrder(data)
  if (!order) throw new Error('订单信息暂不可用，请稍后刷新')
  return order
}

const isMissingOrderError = (error: unknown) => {
  return error instanceof ApiError && error.code === ORDER_NOT_FOUND_CODE
}

const isUncertainResultError = (error: unknown) => {
  return error instanceof ApiError ? error.status >= 500 : error instanceof TypeError
}

export default function Lottery() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const userId = useMemo(() => normalizeId(user?.id), [user?.id])
  const isLoggedIn = Boolean(userId)

  const [rotation, setRotation] = useState(0)
  const [isSpinning, setIsSpinning] = useState(false)
  const [result, setResult] = useState<GiftItem | null>(null)
  const [tempOrderId, setTempOrderId] = useState('')
  const [decisionState, setDecisionState] = useState<'idle' | 'pending' | 'checking' | 'paid' | 'abandoned' | 'missed'>('idle')
  const [countdown, setCountdown] = useState(0)
  const [history, setHistory] = useState<HistoryItem[]>([])
  const [gifts, setGifts] = useState<GiftItem[]>([])
  const [isGiftsLoading, setIsGiftsLoading] = useState(false)
  const [giftsError, setGiftsError] = useState<string | null>(null)
  const [drawError, setDrawError] = useState<string | null>(null)
  const [drawNotice, setDrawNotice] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState<'pay' | 'giveup' | null>(null)
  const [latestOrder, setLatestOrder] = useState<LotteryOrder | null>(null)
  const [pendingOrder, setPendingOrder] = useState<LotteryOrder | null>(null)
  const [orderStatus, setOrderStatus] = useState<'idle' | 'loading' | 'empty' | 'error' | 'ready'>('idle')
  const [orderError, setOrderError] = useState<string | null>(null)
  const [isCheckingResult, setIsCheckingResult] = useState(false)

  const activeRecordId = useRef<number | null>(null)
  const decisionDeadlineRef = useRef<number | null>(null)
  const spinTimerRef = useRef<number | null>(null)
  const countdownTimerRef = useRef<number | null>(null)
  const decisionStateRef = useRef(decisionState)
  const isSubmittingRef = useRef<'pay' | 'giveup' | null>(null)
  const activeGiftRef = useRef<GiftItem | null>(null)
  const activeTempOrderIdRef = useRef('')
  const userIdRef = useRef(userId)

  useEffect(() => {
    userIdRef.current = userId
  }, [userId])

  useEffect(() => {
    decisionStateRef.current = decisionState
  }, [decisionState])

  useEffect(() => {
    isSubmittingRef.current = isSubmitting
  }, [isSubmitting])

  useEffect(() => {
    activeGiftRef.current = result
  }, [result])

  useEffect(() => {
    activeTempOrderIdRef.current = tempOrderId
  }, [tempOrderId])

  const fetchGifts = useCallback(async () => {
    setIsGiftsLoading(true)
    setGiftsError(null)
    try {
      const { data } = await apiGet<unknown>('/lottery/gifts')
      const payload = data as any
      const rawList: unknown[] = Array.isArray(payload)
        ? payload
        : Array.isArray(payload?.gifts)
          ? payload.gifts
          : []
      const normalized = rawList
        .map((item: unknown) => normalizeGift(item))
        .filter((item): item is GiftItem => Boolean(item))
      const uniqueMap = new Map<string, GiftItem>()
      normalized.forEach((item: GiftItem) => {
        if (!uniqueMap.has(item.id)) {
          uniqueMap.set(item.id, item)
        }
      })
      setGifts(Array.from(uniqueMap.values()))
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) {
        setGifts([])
        setGiftsError(null)
        return
      }
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

  const fetchLatestOrder = useCallback(async (signal?: AbortSignal) => {
    if (!userId) {
      setLatestOrder(null)
      setPendingOrder(null)
      setOrderStatus('idle')
      setOrderError(null)
      return null
    }

    setOrderStatus('loading')
    setOrderError(null)
    try {
      const normalized = await requestLatestOrder(signal)
      if (signal?.aborted || userIdRef.current !== userId) return
      setLatestOrder(normalized)
      setPendingOrder(
        normalized.status === ORDER_STATUS_PENDING &&
        (getOrderDeadline(normalized) ?? 0) > Date.now()
          ? normalized
          : null
      )
      setOrderStatus('ready')
      setIsCheckingResult(false)
      return normalized
    } catch (error) {
      if (signal?.aborted || userIdRef.current !== userId) return
      if (isMissingOrderError(error)) {
        setLatestOrder(null)
        setPendingOrder(null)
        setOrderStatus('empty')
        return null
      } else {
        console.error('Failed to fetch lottery result:', error)
        setOrderStatus('error')
        setOrderError(error instanceof Error ? error.message : '获取中奖结果失败')
      }
    }
  }, [userId])

  useEffect(() => {
    if (!isCheckingResult || !userId) return

    const controller = new AbortController()
    let attempts = 0
    let timer: number | undefined
    const checkResult = async () => {
      attempts++
      const order = await fetchLatestOrder(controller.signal)
      if (controller.signal.aborted || order) return
      if (attempts >= MAX_RESULT_RETRIES) {
        setIsCheckingResult(false)
        setDrawNotice('结果暂未确认，请稍后刷新结果或重试抽奖。')
        return
      }
      timer = window.setTimeout(checkResult, RESULT_RETRY_INTERVAL)
    }

    timer = window.setTimeout(checkResult, RESULT_RETRY_INTERVAL)
    return () => {
      controller.abort()
      window.clearTimeout(timer)
    }
  }, [fetchLatestOrder, isCheckingResult, userId])

  useEffect(() => {
    if (!userId) {
      setLatestOrder(null)
      setPendingOrder(null)
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

  useEffect(() => {
    if (spinTimerRef.current) clearTimeout(spinTimerRef.current)
    resetDecisionTimer()
    setIsSpinning(false)
    setIsCheckingResult(false)
    setDecisionState('idle')
    setResult(null)
    setTempOrderId('')
    setLatestOrder(null)
    setPendingOrder(null)
    setDrawError(null)
    setDrawNotice(null)
    setActionError(null)
    setIsSubmitting(null)
    setHistory([])
    activeRecordId.current = null
    activeGiftRef.current = null
    activeTempOrderIdRef.current = ''
    isSubmittingRef.current = null
  }, [resetDecisionTimer, userId])

  const submitGiveup = useCallback(async (mode: 'manual' | 'release') => {
    const isRelease = mode === 'release'
    if (isSubmittingRef.current) return
    const gift = activeGiftRef.current
    const activeTempOrderId = activeTempOrderIdRef.current
    if (!gift) return
    if (!isRelease && isLoseGift(gift)) return
    if (!isRelease && decisionStateRef.current !== 'pending') return
    if (!userId) {
      if (mode === 'manual') {
        alert('请先登录后再操作')
        navigate('/login')
      }
      return
    }
    if (!activeTempOrderId) {
      if (mode === 'manual') {
        setActionError('临时订单信息缺失，请重新抽奖')
      }
      return
    }

    isSubmittingRef.current = 'giveup'
    setIsSubmitting('giveup')
    setActionError(null)

    try {
      await apiPost('/lottery/giveup', {
        user_id: userId,
        gift_id: gift.id,
        temp_order_id: activeTempOrderId,
      })
      if (userIdRef.current !== userId) return
      if (mode === 'manual') {
        setDecisionState('abandoned')
        resetDecisionTimer()
        updateHistoryStatus('已主动放弃')
      }
      if (!isRelease) {
        setDrawNotice('本次抽奖已放弃，无法再次抽取。')
      }
      setTempOrderId('')
      activeTempOrderIdRef.current = ''
      void fetchLatestOrder()
    } catch (error) {
      if (userIdRef.current !== userId) return
      console.error('放弃领取失败:', error)
      const order = await fetchLatestOrder()
      if (userIdRef.current !== userId) return
      if (order && (order.status === ORDER_STATUS_PAID || order.status === ORDER_STATUS_CANCELLED ||
        order.status === ORDER_STATUS_EXPIRED || (getOrderDeadline(order) ?? Infinity) <= Date.now())) return
      const message = error instanceof Error ? error.message : '放弃领取失败'
      if (isRelease) {
        setDrawError(message)
      } else {
        setActionError(message)
      }
    } finally {
      if (userIdRef.current === userId) {
        isSubmittingRef.current = null
        setIsSubmitting(null)
      }
    }
  }, [fetchLatestOrder, navigate, resetDecisionTimer, updateHistoryStatus, userId])

  const startDecisionWindow = useCallback((
    drawResult: LotteryDrawResult,
    orderDeadline?: number | null
  ) => {
    resetDecisionTimer()

    if (!isPayableDrawResult(drawResult)) {
      setDecisionState('missed')
      updateHistoryStatus(getHistoryStatus(drawResult))
      if (drawResult.tempOrderId) {
        void submitGiveup('release')
      }
      return
    }

    const deadline = typeof orderDeadline === 'number' ? orderDeadline : null
    if (!deadline) {
      setDecisionState('checking')
      setIsCheckingResult(true)
      setDrawNotice('订单有效期暂未确认，正在查询结果。')
      return
    }

    const initialRemaining = Math.max(0, Math.ceil((deadline - Date.now()) / 1000))
    if (initialRemaining <= 0) {
      setDecisionState('abandoned')
      setDrawNotice('该订单已超过确认时限，本次抽奖已结束。')
      updateHistoryStatus('订单已超时')
      return
    }

    setActionError(null)
    setDecisionState('pending')
    setCountdown(initialRemaining)
    decisionDeadlineRef.current = deadline

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
        setDecisionState('abandoned')
        setDrawNotice('订单已超过确认时限，系统将自动释放。')
        updateHistoryStatus('订单已超时')
        void fetchLatestOrder()
      }
    }, 250)
  }, [fetchLatestOrder, resetDecisionTimer, submitGiveup, updateHistoryStatus])

  useEffect(() => {
    if (!latestOrder || isSpinning) return
    const paid = latestOrder.status === ORDER_STATUS_PAID
    const ended = latestOrder.status === ORDER_STATUS_CANCELLED ||
      latestOrder.status === ORDER_STATUS_EXPIRED ||
      (latestOrder.status === ORDER_STATUS_PENDING && (getOrderDeadline(latestOrder) ?? Infinity) <= Date.now())
    if (!paid && !ended) return

    resetDecisionTimer()
    setResult(latestOrder.gift)
    activeGiftRef.current = latestOrder.gift
    setTempOrderId('')
    activeTempOrderIdRef.current = ''
    setDecisionState(paid ? 'paid' : isLoseGift(latestOrder.gift) ? 'missed' : 'abandoned')
    setDrawError(null)
    setActionError(null)
    setIsCheckingResult(false)
    const notice = paid
      ? '领取已确认，可在下方查看中奖结果。'
      : latestOrder.status === ORDER_STATUS_CANCELLED
        ? '本次抽奖已放弃，无法再次抽取。'
        : '订单已超过确认时限，本次抽奖已结束。'
    setDrawNotice(notice)
    updateHistoryStatus(paid ? '已确认领取' : latestOrder.status === ORDER_STATUS_CANCELLED ? '已放弃' : '订单已超时')
  }, [isSpinning, latestOrder, resetDecisionTimer, updateHistoryStatus])

  useEffect(() => {
    if (!pendingOrder || isSpinning) return

    const deadline = getOrderDeadline(pendingOrder)
    if (!deadline || deadline <= Date.now()) return

    const restoredResult: LotteryDrawResult = {
      gift: pendingOrder.gift,
      tempOrderId: pendingOrder.id,
      success: false,
      description: RESULT_DESCRIPTION_TEMP_ORDER,
      userId: pendingOrder.user.id || userId,
      status: 'pending_order',
    }

    setResult(pendingOrder.gift)
    setTempOrderId(pendingOrder.id)
    activeGiftRef.current = pendingOrder.gift
    activeTempOrderIdRef.current = pendingOrder.id
    setDrawError(null)
    setActionError(null)
    setDrawNotice(getDrawNotice(restoredResult))
    startDecisionWindow(restoredResult, deadline)
    setPendingOrder(null)
  }, [isSpinning, pendingOrder, startDecisionWindow, userId])

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
    if (isSpinning || isCheckingResult || isSubmittingRef.current || isGiftsLoading || gifts.length === 0) return
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
    setTempOrderId('')
    activeTempOrderIdRef.current = ''
    activeGiftRef.current = null
    activeRecordId.current = null
    setDrawError(null)
    setDrawNotice(null)
    setActionError(null)
    resetDecisionTimer()

    try {
      const { data } = await apiGet<unknown>('/lottery/lucky')
      if (userIdRef.current !== userId) return
      const drawResult = normalizeDrawResult(data)
      if (!drawResult) {
        throw new Error('抽奖失败，请稍后重试')
      }
      const { gift } = drawResult
      let orderDeadline: number | null | undefined

      if (isPayableDrawResult(drawResult)) {
        try {
          const order = await requestLatestOrder()
          if (userIdRef.current !== userId) return
          if (order.id !== drawResult.tempOrderId) {
            throw new Error('订单信息暂未确认')
          }
          setLatestOrder(order)
          setOrderStatus('ready')
          setOrderError(null)
          orderDeadline = getOrderDeadline(order)
          if (order.status !== ORDER_STATUS_PENDING || (orderDeadline ?? Infinity) <= Date.now()) {
            setIsSpinning(false)
            return
          }
        } catch (error) {
          if (userIdRef.current !== userId) return
          console.error('Failed to fetch lottery order deadline:', error)
        }
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
        if (userIdRef.current !== userId) return
        setResult(gift)
        setTempOrderId(drawResult.tempOrderId)
        activeGiftRef.current = gift
        activeTempOrderIdRef.current = drawResult.tempOrderId
        setDrawNotice(getDrawNotice(drawResult))
        setIsSpinning(false)
        const recordId = Date.now()
        activeRecordId.current = recordId
        setHistory(prev => [
          {
            id: recordId,
            prize: gift.name,
            status: getHistoryStatus(drawResult),
            time: new Date().toLocaleTimeString(),
          },
          ...prev,
        ].slice(0, MAX_HISTORY))

        startDecisionWindow(drawResult, orderDeadline)
      }, SPIN_TIMEOUT)
    } catch (error) {
      if (userIdRef.current !== userId) return
      console.error('Failed to draw lottery:', error)
      setIsSpinning(false)
      if (error instanceof ApiError &&
        (error.code === LOTTERY_ENDED_CODE || error.code === LOTTERY_ALREADY_PAID_CODE)) {
        setDrawError(error.message)
        await fetchLatestOrder()
      } else if (isUncertainResultError(error)) {
        setIsCheckingResult(true)
        setDrawNotice('抽奖结果暂未确认，正在查询，请稍候。')
      } else {
        setDrawError(error instanceof Error ? error.message : '抽奖失败，请稍后重试')
      }
    }
  }

  const handlePay = async () => {
    if (decisionState !== 'pending' || isSubmittingRef.current) return
    if (!result || isLoseGift(result)) return
    if (!tempOrderId) {
      setActionError('临时订单信息缺失，请重新抽奖')
      return
    }
    if (!userId) {
      alert('请先登录后再操作')
      navigate('/login')
      return
    }

    isSubmittingRef.current = 'pay'
    setIsSubmitting('pay')
    setActionError(null)

    try {
      await apiPost('/lottery/pay', {
        user_id: userId,
        gift_id: result.id,
        temp_order_id: tempOrderId,
      })
      if (userIdRef.current !== userId) return
      setDecisionState('paid')
      resetDecisionTimer()
      setTempOrderId('')
      activeTempOrderIdRef.current = ''
      setDrawNotice(null)
      updateHistoryStatus('已确认领取')
      void fetchLatestOrder()
    } catch (error) {
      if (userIdRef.current !== userId) return
      console.error('确认领取失败:', error)
      const order = await fetchLatestOrder()
      if (userIdRef.current !== userId) return
      if (order && (order.status === ORDER_STATUS_PAID || order.status === ORDER_STATUS_CANCELLED ||
        order.status === ORDER_STATUS_EXPIRED || (getOrderDeadline(order) ?? Infinity) <= Date.now())) return
      if (isUncertainResultError(error) && !order) {
        resetDecisionTimer()
        setDecisionState('checking')
        setIsCheckingResult(true)
        setActionError('领取结果暂未确认，正在查询，请勿重复操作。')
      } else if (error instanceof ApiError && error.code === ORDER_NOT_AVAILABLE_CODE && order === null) {
        resetDecisionTimer()
        setDecisionState('abandoned')
        setActionError(error.message)
      } else {
        setActionError(error instanceof Error ? error.message : '确认领取失败，请稍后重试')
      }
    } finally {
      if (userIdRef.current === userId) {
        isSubmittingRef.current = null
        setIsSubmitting(null)
      }
    }
  }

  const handleReset = () => {
    resetDecisionTimer()
    setDecisionState('idle')
    setResult(null)
    setTempOrderId('')
    activeTempOrderIdRef.current = ''
    setDrawError(null)
    setDrawNotice(null)
    setActionError(null)
    activeRecordId.current = null
  }

  const hasFinishedOrder = Boolean(latestOrder && (
    latestOrder.status === ORDER_STATUS_PAID ||
    latestOrder.status === ORDER_STATUS_CANCELLED ||
    latestOrder.status === ORDER_STATUS_EXPIRED ||
    (latestOrder.status === ORDER_STATUS_PENDING && (getOrderDeadline(latestOrder) ?? Infinity) <= Date.now())
  ))

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
              奖品池来自后台奖品列表，中奖后需在 {decisionMinutes} 分钟内确认领取或放弃，超时系统将自动释放订单。
            </p>
            <div className="flex flex-wrap gap-3 text-sm text-gray-600">
              <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white border border-gray-200">
                <Clock3 className="h-4 w-4 text-primary-600" />
                {decisionMinutes} 分钟确认时限
              </span>
              <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white border border-gray-200">
                <Wallet className="h-4 w-4 text-primary-600" />
                中奖后需确认领取
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
                disabled={isSpinning || isCheckingResult || Boolean(isSubmitting) || isGiftsLoading || gifts.length === 0 ||
                  hasFinishedOrder || decisionState === 'pending' || decisionState === 'checking'}
                className="btn-primary flex items-center gap-2 shadow-md disabled:cursor-not-allowed"
              >
                <Sparkles className="h-5 w-5" />
                {isSpinning ? '抽奖中...' : isCheckingResult || decisionState === 'checking' ? '结果确认中' :
                  hasFinishedOrder ? '本次抽奖已结束' : decisionState === 'pending' ? '已有待确认订单' : '启动抽奖'}
              </button>
              <button
                type="button"
                onClick={handleReset}
                disabled={isSpinning || isCheckingResult || Boolean(isSubmitting)}
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
                  中奖需在 {decisionMinutes} 分钟内选择确认领取或放弃，超时自动释放
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
                <h2 className="text-lg font-semibold text-gray-900">确认领取或放弃</h2>
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
                  ? '先启动抽奖，抽中后会在这里展示奖品并进入确认领取/放弃倒计时。'
                  : '登录后即可参与抽奖，中奖结果会显示在这里。'}
              </div>
            )}

            {drawError && (
              <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 border border-red-100 px-3 py-2 rounded-lg">
                <XCircle className="h-4 w-4" />
                {drawError}
              </div>
            )}

            {drawNotice && (
              <div className="flex items-center gap-2 text-sm text-primary-700 bg-primary-50 border border-primary-100 px-3 py-2 rounded-lg">
                <Clock3 className="h-4 w-4" />
                <span className="flex-1">{drawNotice}</span>
                <button
                  type="button"
                  onClick={() => void fetchLatestOrder()}
                  disabled={orderStatus === 'loading'}
                  className="font-semibold whitespace-nowrap disabled:opacity-60"
                >
                  查询结果
                </button>
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
                    很遗憾未中奖，谢谢参与。
                  </div>
                )}

                {!isLoseGift(result) && (
                  <div className="flex flex-col gap-3">
                    <div className="flex items-center gap-2 text-sm">
                      <Clock3 className="h-4 w-4 text-primary-600" />
                      <span className="text-gray-700">
                        {decisionState === 'paid' ? '该订单已确认领取。' :
                          decisionState === 'abandoned' ? '本次抽奖已结束。' :
                          decisionState === 'checking' ? '正在确认订单状态，请稍候。' :
                          '请在倒计时结束前确认领取或放弃。'}
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
                        {isSubmitting === 'pay' ? '确认中...' : '确认领取'}
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
                        领取已确认，奖品已锁定。
                      </div>
                    )}
                    {decisionState === 'abandoned' && (
                      <div className="flex items-center gap-2 text-sm text-orange-600 bg-orange-50 border border-orange-100 px-3 py-2 rounded-lg">
                        <XCircle className="h-4 w-4" />
                        本次抽奖已结束，不能再次抽取。
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
                <h3 className="text-lg font-semibold text-gray-900">抽奖订单</h3>
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
              <p className="text-sm text-gray-600">登录后查看抽奖订单和领取状态。</p>
            )}

            {isLoggedIn && orderStatus === 'loading' && (
              <p className="text-sm text-gray-600">正在查询抽奖订单...</p>
            )}

            {isLoggedIn && orderStatus === 'empty' && (
              <p className="text-sm text-gray-600">
                {isCheckingResult ? '暂未查询到订单，正在确认抽奖结果。' : '暂无抽奖订单。'}
              </p>
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
                  <p className="text-sm text-primary-700">{getOrderStatusText(latestOrder)}</p>
                  {latestOrder.gift.description && (
                    <p className="text-sm text-gray-600">{latestOrder.gift.description}</p>
                  )}
                  <div className="flex flex-wrap gap-3 text-xs text-gray-500">
                    <span>数量 {latestOrder.count}</span>
                    {(latestOrder.paid_at || latestOrder.created_at) && (
                      <span>{latestOrder.status === ORDER_STATUS_PAID ? '确认时间' : '下单时间'} {formatDateTime(latestOrder.paid_at || latestOrder.created_at)}</span>
                    )}
                    {latestOrder.status === ORDER_STATUS_PENDING && latestOrder.expire_at && (
                      <span>确认截止 {formatDateTime(latestOrder.expire_at)}</span>
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
              <li>抽奖需登录，中奖后请在 {decisionMinutes} 分钟内确认领取或放弃。</li>
              <li>每位用户参与一次，已有订单会复用，放弃或超时后不能重新抽取。</li>
              <li>抽到「谢谢参与」或奖品已抽完不会进入领取流程。</li>
              <li>领取状态以「抽奖订单」为准，结果暂未确认时可稍后查询。</li>
              <li>抽奖记录仅保留本地最近 {MAX_HISTORY} 条。</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
