import {
  useState,
  useMemo,
  useRef,
  useEffect,
  useCallback,
  FormEvent,
  type UIEvent,
  type WheelEvent,
  type MouseEvent,
} from 'react'
import { Link, useLocation } from 'react-router-dom'
import { ArrowLeft, Send, MessageCircle, Trash2 } from 'lucide-react'
import UserAvatar from '../components/UserAvatar'
import { useAuth } from '../contexts/AuthContext'
import { apiGet, apiPost, API_BASE_URL } from '../utils/api'
import { formatRelativeTime } from '../utils/date'
import { normalizeId } from '../utils/id'

const MESSAGE_PAGE_SIZE = 8
const SESSION_TYPE_PRIVATE = 1
const HISTORY_REVEAL_DELAY_MS = 500
const MESSAGE_TIME_GAP_MS = 5 * 60 * 1000
const MESSAGE_CONFIRM_TIMEOUT_MS = 15 * 1000
const MESSAGE_CONFIRM_MATCH_WINDOW_MS = 5 * 60 * 1000

type DeliveryStatus = 'pending' | 'failed'

type SessionResponse = {
  session_id?: string | number
  target_id?: string | number
  target_name?: string
  target_avatar?: string
  last_message?: string
  last_message_time?: string
  unread_count?: number
  session_type?: number
}

type Conversation = {
  id: string
  targetId: string
  name: string
  avatar: string
  lastMessage: string
  lastMessageTime?: string
  lastMessageDeliveryStatus?: DeliveryStatus
  lastLocalMessageId?: string
  unread: number
  sessionType: number
}

type ChatMessage = {
  id: string
  from: 'me' | 'other'
  content: string
  createdAt?: string
  deliveryStatus?: DeliveryStatus
}

type MessageResponse = {
  content?: string
  message_from?: string | number
  message_to?: string | number
  id?: string | number
  session_id?: string | number
  session_type?: number
  created_at?: string
  createdAt?: string
}

type MessageListResponse = {
  total?: number
  has_more?: boolean
  messages?: MessageResponse[]
}

type MessagePageState = {
  pageNo: number
  hasMore: boolean
  isLoading: boolean
  error: string | null
  total?: number
}

type LocationState = {
  userId?: string
  username?: string
}

const buildWsUrl = (apiBaseUrl: string) => {
  const base = new URL(apiBaseUrl, window.location.origin)
  base.protocol = base.protocol === 'https:' ? 'wss:' : 'ws:'
  const cleanPath = base.pathname.replace(/\/+$/, '')
  base.pathname = `${cleanPath}/ws/session`
  base.search = ''
  return base.toString()
}

const toTimestamp = (value?: string) => {
  if (!value) return 0
  const parsed = new Date(value).getTime()
  return Number.isNaN(parsed) ? 0 : parsed
}

const wait = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null

const extractWsMessages = (payload: unknown): MessageResponse[] => {
  if (!payload) return []
  if (Array.isArray(payload)) {
    return payload.filter(Boolean) as MessageResponse[]
  }
  if (!isRecord(payload)) return []

  // 新 WebSocket 网关使用 biz_type 区分业务模块，聊天消息位于 biz_data。
  const bizType = typeof payload.biz_type === 'string' ? payload.biz_type : ''
  if (bizType && bizType !== 'session') return []

  const type = typeof payload.type === 'string' ? payload.type : ''
  if (type && type !== 'message') return []

  const hasMessageFields =
    'session_id' in payload ||
    'message_from' in payload ||
    'message_to' in payload ||
    'content' in payload ||
    'created_at' in payload ||
    'createdAt' in payload
  if (hasMessageFields) {
    return [payload as MessageResponse]
  }

  const nested = payload.biz_data ?? payload.data ?? payload.message ?? payload.payload ?? payload.messages
  if (Array.isArray(nested)) {
    return nested.filter(Boolean) as MessageResponse[]
  }
  if (isRecord(nested)) {
    return [nested as MessageResponse]
  }
  return []
}

const sortMessages = (messages: ChatMessage[]) =>
  [...messages].sort((a, b) => {
    const diff = toTimestamp(a.createdAt) - toTimestamp(b.createdAt)
    if (diff !== 0) return diff
    return a.id.localeCompare(b.id, undefined, { numeric: true })
  })

type MessageConfirmation = {
  localId: string
  confirmedMessage: ChatMessage
}

type MessageMergeResult = {
  messages: ChatMessage[]
  confirmations: MessageConfirmation[]
}

const reconcileIncomingMessage = (
  messages: ChatMessage[],
  incoming: ChatMessage,
  options: { allowTimelessConfirmation?: boolean } = {}
): MessageMergeResult => {
  if (!incoming.id || messages.some((msg) => msg.id === incoming.id)) {
    return { messages, confirmations: [] }
  }
  if (incoming.from === 'me' && !incoming.deliveryStatus) {
    const incomingTimestamp = toTimestamp(incoming.createdAt)
    const awaitingIndex = messages.findIndex(
      (msg) => {
        if (
          msg.from !== 'me' ||
          !msg.deliveryStatus ||
          msg.content !== incoming.content
        ) {
          return false
        }
        const localTimestamp = toTimestamp(msg.createdAt)
        if (incomingTimestamp > 0 && localTimestamp > 0) {
          return Math.abs(incomingTimestamp - localTimestamp) <= MESSAGE_CONFIRM_MATCH_WINDOW_MS
        }
        return Boolean(options.allowTimelessConfirmation)
      }
    )
    if (awaitingIndex !== -1) {
      const localId = messages[awaitingIndex].id
      const next = [...messages]
      next[awaitingIndex] = { ...incoming }
      return {
        messages: next,
        confirmations: [{ localId, confirmedMessage: incoming }],
      }
    }
  }
  return { messages: [...messages, incoming], confirmations: [] }
}

const mergeMessageList = (
  current: ChatMessage[],
  incoming: ChatMessage[],
  options: { allowTimelessConfirmation?: boolean } = {}
): MessageMergeResult => {
  let merged = current
  const confirmations: MessageConfirmation[] = []
  for (const message of incoming) {
    const result = reconcileIncomingMessage(merged, message, options)
    merged = result.messages
    confirmations.push(...result.confirmations)
  }
  return {
    messages: sortMessages(merged),
    confirmations,
  }
}

const summarizeMessage = (content: string) => Array.from(content).slice(0, 5).join('')

const normalizeSession = (
  item: SessionResponse,
  fallback?: { targetId?: string; name?: string }
): Conversation => {
  const name = typeof item.target_name === 'string' ? item.target_name.trim() : ''
  const fallbackName = typeof fallback?.name === 'string' ? fallback.name.trim() : ''
  const targetId = normalizeId(item.target_id ?? fallback?.targetId)
  const avatar = typeof item.target_avatar === 'string' ? item.target_avatar.trim() : ''
  const lastMessage = typeof item.last_message === 'string' ? item.last_message.trim() : ''
  const lastMessageTime =
    typeof item.last_message_time === 'string' && item.last_message_time
      ? item.last_message_time
      : undefined
  const unreadCount = Number.isFinite(Number(item.unread_count)) ? Number(item.unread_count) : 0
  const sessionType = Number.isFinite(Number(item.session_type)) ? Number(item.session_type) : SESSION_TYPE_PRIVATE

  return {
    id: normalizeId(item.session_id),
    targetId,
    name: name || fallbackName || (targetId ? `用户 ${targetId}` : '未知用户'),
    avatar,
    lastMessage,
    lastMessageTime,
    unread: unreadCount,
    sessionType,
  }
}

const normalizeMessage = (
  item: MessageResponse,
  currentUserId: string
): ChatMessage | null => {
  if (!item) return null
  const content = typeof item.content === 'string' ? item.content : String(item.content ?? '')
  const fromId = normalizeId(item.message_from)
  const createdAt =
    typeof item.created_at === 'string'
      ? item.created_at
      : typeof item.createdAt === 'string'
        ? item.createdAt
        : undefined
  const isMe = Boolean(currentUserId && fromId && currentUserId === fromId)
  const id = normalizeId(item.id)
  if (!id) return null

  return {
    id,
    from: isMe ? 'me' : 'other',
    content,
    createdAt,
  }
}

export default function Messages() {
  const { user } = useAuth()
  const location = useLocation()
  const locationState = (location.state ?? {}) as LocationState
  const routeTargetId = normalizeId(locationState.userId)
  const routeTargetName = typeof locationState.username === 'string' ? locationState.username.trim() : ''
  const currentUserId = useMemo(() => normalizeId(user?.id), [user?.id])
  const [sessions, setSessions] = useState<Conversation[]>([])
  const [activeId, setActiveId] = useState<string>('')
  const [messagesBySession, setMessagesBySession] = useState<Record<string, ChatMessage[]>>({})
  const [messagePageBySession, setMessagePageBySession] = useState<Record<string, MessagePageState>>({})
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sendError, setSendError] = useState<string | null>(null)
  const [deletingSessionIds, setDeletingSessionIds] = useState<Record<string, boolean>>({})
  const [refreshKey, setRefreshKey] = useState(0)
  const [connectionStatus, setConnectionStatus] = useState<'disconnected' | 'connecting' | 'connected' | 'error'>('disconnected')
  const messagesContainerRef = useRef<HTMLDivElement>(null)
  const historyObserverTargetRef = useRef<HTMLDivElement>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const activeIdRef = useRef('')
  const sessionsRef = useRef<Conversation[]>([])
  const messagesBySessionRef = useRef<Record<string, ChatMessage[]>>({})
  const messagePageBySessionRef = useRef<Record<string, MessagePageState>>({})
  const confirmationTimersRef = useRef<Map<string, number>>(new Map())
  const localMessageSequenceRef = useRef(0)
  const handledRouteTargetRef = useRef(false)
  const pendingScrollAdjustmentRef = useRef<{
    sessionId: string
    previousHeight: number
    previousTop: number
  } | null>(null)
  const isFetchingMoreRef = useRef(false)
  const hasUserScrolledRef = useRef(false)
  const lastScrollTopRef = useRef(0)

  const activeSession = useMemo(
    () => sessions.find((session) => session.id === activeId) ?? null,
    [sessions, activeId]
  )
  const activeSessionId = activeSession?.id ?? ''
  const activeSessionTargetId = activeSession?.targetId ?? ''
  const chatList = useMemo(() => {
    if (!activeId) return []
    return messagesBySession[activeId] || []
  }, [messagesBySession, activeId])
  const activeMessageState = activeId ? messagePageBySession[activeId] : undefined
  const connectionBadge = useMemo(() => {
    if (!user) {
      return {
        label: '请先登录',
        dot: 'bg-gray-300',
        text: 'text-gray-500',
        bg: 'bg-gray-50',
        ring: 'ring-gray-200',
        pulse: false,
      }
    }
    if (connectionStatus === 'connected') {
      return {
        label: '已连接',
        dot: 'bg-emerald-500',
        text: 'text-emerald-700',
        bg: 'bg-emerald-50',
        ring: 'ring-emerald-200',
        pulse: false,
      }
    }
    if (connectionStatus === 'connecting') {
      return {
        label: '连接中',
        dot: 'bg-amber-500',
        text: 'text-amber-700',
        bg: 'bg-amber-50',
        ring: 'ring-amber-200',
        pulse: true,
      }
    }
    if (connectionStatus === 'error') {
      return {
        label: '连接异常',
        dot: 'bg-red-500',
        text: 'text-red-700',
        bg: 'bg-red-50',
        ring: 'ring-red-200',
        pulse: false,
      }
    }
    return {
      label: '已断开',
      dot: 'bg-red-500',
      text: 'text-red-700',
      bg: 'bg-red-50',
      ring: 'ring-red-200',
      pulse: false,
    }
  }, [connectionStatus, user])

  const clearConfirmationTimer = useCallback((localId: string) => {
    const timerId = confirmationTimersRef.current.get(localId)
    if (timerId !== undefined) {
      window.clearTimeout(timerId)
      confirmationTimersRef.current.delete(localId)
    }
  }, [])

  const mergeIncomingMessages = useCallback(
    (
      sessionId: string,
      incoming: ChatMessage[],
      options: { allowTimelessConfirmation?: boolean } = {}
    ) => {
      const currentState = messagesBySessionRef.current
      const result = mergeMessageList(currentState[sessionId] ?? [], incoming, options)
      const nextState = {
        ...currentState,
        [sessionId]: result.messages,
      }
      messagesBySessionRef.current = nextState
      setMessagesBySession(nextState)
      result.confirmations.forEach(({ localId }) => clearConfirmationTimer(localId))
      return result.confirmations
    },
    [clearConfirmationTimer]
  )

  const applyConfirmationsToSession = useCallback(
    (sessionId: string, confirmations: MessageConfirmation[]) => {
      if (confirmations.length === 0) return
      setSessions((prev) =>
        prev.map((session) => {
          if (session.id !== sessionId || !session.lastLocalMessageId) return session
          const confirmation = confirmations.find(
            ({ localId }) => localId === session.lastLocalMessageId
          )
          if (!confirmation) return session
          return {
            ...session,
            lastMessage: summarizeMessage(confirmation.confirmedMessage.content),
            lastMessageTime: confirmation.confirmedMessage.createdAt,
            lastMessageDeliveryStatus: undefined,
            lastLocalMessageId: undefined,
          }
        })
      )
    },
    []
  )

  const markMessageUnconfirmed = useCallback((sessionId: string, localId: string) => {
    confirmationTimersRef.current.delete(localId)
    const currentState = messagesBySessionRef.current
    const currentMessages = currentState[sessionId] ?? []
    let didMarkFailed = false
    const nextMessages = currentMessages.map((message) => {
      if (message.id !== localId || message.deliveryStatus !== 'pending') return message
      didMarkFailed = true
      return { ...message, deliveryStatus: 'failed' as const }
    })
    if (!didMarkFailed) return

    const nextState = {
      ...currentState,
      [sessionId]: nextMessages,
    }
    messagesBySessionRef.current = nextState
    setMessagesBySession(nextState)
    setSessions((prev) =>
      prev.map((session) =>
        session.id === sessionId && session.lastLocalMessageId === localId
          ? { ...session, lastMessageDeliveryStatus: 'failed' }
          : session
      )
    )
    if (activeIdRef.current === sessionId) {
      setSendError('消息已写入连接，但未收到后端确认，请检查后再决定是否重发')
    }
  }, [])

  useEffect(() => {
    activeIdRef.current = activeId
  }, [activeId])

  useEffect(() => {
    sessionsRef.current = sessions
  }, [sessions])

  useEffect(() => {
    messagesBySessionRef.current = messagesBySession
  }, [messagesBySession])

  useEffect(() => {
    messagePageBySessionRef.current = messagePageBySession
  }, [messagePageBySession])

  useEffect(
    () => () => {
      confirmationTimersRef.current.forEach((timerId) => window.clearTimeout(timerId))
      confirmationTimersRef.current.clear()
    },
    []
  )

  useEffect(() => {
    pendingScrollAdjustmentRef.current = null
    isFetchingMoreRef.current = false
    hasUserScrolledRef.current = false
    lastScrollTopRef.current = 0
  }, [activeSessionId])

  useEffect(() => {
    handledRouteTargetRef.current = false
  }, [routeTargetId, user?.id])

  useEffect(() => {
    setSendError(null)
  }, [activeId])

  const sendReadAck = useCallback((sessionId: string) => {
    const normalizedSessionId = normalizeId(sessionId)
    const ws = wsRef.current
    if (!normalizedSessionId || !ws || ws.readyState !== WebSocket.OPEN) return
    const payload = {
      biz_type: 'session',
      biz_data: {
        type: 'read_ack',
        session_id: normalizedSessionId,
      },
    }
    ws.send(JSON.stringify(payload))
  }, [])

  const fetchMessages = useCallback(
    async (
      sessionId: string,
      targetId: string,
      pageNo: number,
      options: { signal?: AbortSignal; minDelayMs?: number } = {}
    ) => {
      const { signal, minDelayMs = 0 } = options
      const normalizedSessionId = normalizeId(sessionId)
      const normalizedTargetId = normalizeId(targetId)
      if (!normalizedSessionId || !normalizedTargetId || !currentUserId) return

      setMessagePageBySession((prev) => {
        const previous = prev[normalizedSessionId] ?? { pageNo: 0, hasMore: false, isLoading: false, error: null }
        return {
          ...prev,
          [normalizedSessionId]: { ...previous, isLoading: true, error: null },
        }
      })

      try {
        const delayPromise = minDelayMs > 0 ? wait(minDelayMs) : Promise.resolve()
        const [{ data }] = await Promise.all([
          apiGet<MessageListResponse>(
            `/sessions/target/${encodeURIComponent(normalizedTargetId)}/messages?pageNo=${pageNo}&pageSize=${MESSAGE_PAGE_SIZE}`,
            { signal }
          ),
          delayPromise,
        ])
        const rawMessages = Array.isArray(data?.messages) ? data.messages : []
        const normalized = rawMessages
          .map((item) => normalizeMessage(item, currentUserId))
          .filter((item): item is ChatMessage => Boolean(item))
        const hasMoreValue = Boolean(
          data?.has_more ?? (data as { hasMore?: boolean } | null)?.hasMore
        )

        const confirmations = mergeIncomingMessages(normalizedSessionId, normalized)
        applyConfirmationsToSession(normalizedSessionId, confirmations)

        setMessagePageBySession((prev) => {
          const previous = prev[normalizedSessionId] ?? { pageNo: 0, hasMore: false, isLoading: false, error: null }
          return {
            ...prev,
            [normalizedSessionId]: {
              ...previous,
              pageNo,
              hasMore: hasMoreValue,
              total: typeof data?.total === 'number' ? data.total : previous.total,
              isLoading: false,
              error: null,
            },
          }
        })

        if (activeIdRef.current === normalizedSessionId) {
          const hadUnread =
            sessionsRef.current.find((item) => item.id === normalizedSessionId)?.unread ?? 0
          setSessions((prev) =>
            prev.map((item) => (item.id === normalizedSessionId ? { ...item, unread: 0 } : item))
          )
          if (hadUnread > 0) {
            sendReadAck(normalizedSessionId)
          }
        }
      } catch (err) {
        if ((err as { name?: string })?.name === 'AbortError') return
        const message = err instanceof Error ? err.message : '获取聊天记录失败'
        setMessagePageBySession((prev) => {
          const previous = prev[normalizedSessionId] ?? { pageNo: 0, hasMore: false, isLoading: false, error: null }
          return {
            ...prev,
            [normalizedSessionId]: {
              ...previous,
              isLoading: false,
              error: message,
            },
          }
        })
      }
    },
    [
      applyConfirmationsToSession,
      currentUserId,
      mergeIncomingMessages,
      sendReadAck,
    ]
  )

  const fetchOrCreateSession = useCallback(
    async (
      targetId: string,
      fallbackName?: string,
      options: { activate?: boolean; notify?: boolean } = {}
    ) => {
      const normalizedTargetId = normalizeId(targetId)
      if (!normalizedTargetId) return null
      const { activate = true, notify = true } = options

      try {
        const { data } = await apiGet<SessionResponse>(
          `/sessions/target/${encodeURIComponent(normalizedTargetId)}`
        )
        const session = normalizeSession(data ?? {}, { targetId: normalizedTargetId, name: fallbackName })
        if (!session.id) return null

        setSessions((prev) => {
          const index = prev.findIndex((item) => item.id === session.id)
          if (index === -1) {
            return [session, ...prev]
          }
          const existing = prev[index]
          const merged = {
            ...existing,
            ...session,
            unread: Math.max(existing.unread, session.unread),
          }
          const next = [...prev]
          next.splice(index, 1)
          return [merged, ...next]
        })

        if (activate) {
          setActiveId(session.id)
        }
        return session
      } catch (err) {
        if (notify) {
          const message = err instanceof Error ? err.message : '创建会话失败'
          setError(message)
        }
        return null
      }
    },
    []
  )

  useEffect(() => {
    if (!user) {
      confirmationTimersRef.current.forEach((timerId) => window.clearTimeout(timerId))
      confirmationTimersRef.current.clear()
      sessionsRef.current = []
      messagesBySessionRef.current = {}
      setSessions([])
      setActiveId('')
      setMessagesBySession({})
      setMessagePageBySession({})
      setIsLoading(false)
      setError(null)
      setConnectionStatus('disconnected')
      return
    }

    let isMounted = true
    const controller = new AbortController()
    setIsLoading(true)
    setError(null)

    const loadSessions = async () => {
      try {
        const { data } = await apiGet<SessionResponse[]>('/sessions', { signal: controller.signal })
        if (!isMounted) return

        const list = Array.isArray(data) ? data : []
        const normalized = list
          .map((item) => normalizeSession(item))
          .filter((item) => item.id && item.targetId)

        setSessions(normalized)
        setActiveId((prev) => {
          if (prev && normalized.some((session) => session.id === prev)) {
            return prev
          }
          return ''
        })
      } catch (err) {
        if (!isMounted) return
        if ((err as { name?: string })?.name === 'AbortError') return
        console.error('Failed to fetch sessions:', err)
        const message = err instanceof Error ? err.message : '获取会话列表失败'
        setError(message)
        setSessions([])
        setActiveId('')
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    void loadSessions()

    return () => {
      isMounted = false
      controller.abort()
    }
  }, [user, refreshKey])

  useEffect(() => {
    if (!user || !routeTargetId || handledRouteTargetRef.current) return
    if (isLoading) return

    const existing = sessions.find((session) => session.targetId === routeTargetId)
    if (existing) {
      setActiveId(existing.id)
      handledRouteTargetRef.current = true
      return
    }

    void fetchOrCreateSession(routeTargetId, routeTargetName, { activate: true, notify: true })
      .finally(() => {
        handledRouteTargetRef.current = true
      })
  }, [fetchOrCreateSession, isLoading, routeTargetId, routeTargetName, sessions, user])

  useEffect(() => {
    if (!activeSessionId || !activeSessionTargetId || !currentUserId) return
    const state = messagePageBySessionRef.current[activeSessionId]
    if (state?.pageNo || state?.isLoading) return

    const controller = new AbortController()
    void fetchMessages(activeSessionId, activeSessionTargetId, 1, {
      signal: controller.signal,
      minDelayMs: 0,
    })

    return () => controller.abort()
  }, [activeSessionId, activeSessionTargetId, currentUserId, fetchMessages])

  const loadMoreMessages = useCallback(async () => {
    if (!activeSessionId || !activeSessionTargetId) return
    const state = messagePageBySessionRef.current[activeSessionId]
    if (!state?.hasMore || state.isLoading || isFetchingMoreRef.current) return

    const container = messagesContainerRef.current
    if (container) {
      pendingScrollAdjustmentRef.current = {
        sessionId: activeSessionId,
        previousHeight: container.scrollHeight,
        previousTop: container.scrollTop,
      }
    }

    isFetchingMoreRef.current = true
    try {
      const nextPage = (state.pageNo ?? 0) + 1
      await fetchMessages(activeSessionId, activeSessionTargetId, nextPage, {
        minDelayMs: HISTORY_REVEAL_DELAY_MS,
      })
    } finally {
      isFetchingMoreRef.current = false
    }
  }, [activeSessionId, activeSessionTargetId, fetchMessages])

  const triggerLoadMoreIfAtTop = useCallback(() => {
    const container = messagesContainerRef.current
    if (!container || container.scrollTop > 0) return
    void loadMoreMessages()
  }, [loadMoreMessages])

  const handleMessagesScroll = useCallback((event: UIEvent<HTMLDivElement>) => {
    const currentTop = event.currentTarget.scrollTop
    if (currentTop < lastScrollTopRef.current) {
      hasUserScrolledRef.current = true
    }
    lastScrollTopRef.current = currentTop
  }, [])

  const handleMessagesWheel = useCallback(
    (event: WheelEvent<HTMLDivElement>) => {
      if (event.deltaY < 0) {
        hasUserScrolledRef.current = true
        triggerLoadMoreIfAtTop()
      }
    },
    [triggerLoadMoreIfAtTop]
  )

  const handleMessagesTouchMove = useCallback(() => {
    hasUserScrolledRef.current = true
    triggerLoadMoreIfAtTop()
  }, [triggerLoadMoreIfAtTop])

  useEffect(() => {
    const container = messagesContainerRef.current
    const target = historyObserverTargetRef.current
    if (!container || !target) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (!entries[0]?.isIntersecting) return
        if (!hasUserScrolledRef.current) return
        if (!activeSessionId || !activeSessionTargetId) return
        if (!activeMessageState?.hasMore || activeMessageState?.isLoading || activeMessageState?.error) return
        void loadMoreMessages()
      },
      { root: container, rootMargin: '120px 0px 0px 0px', threshold: 0.1 }
    )

    observer.observe(target)

    return () => observer.disconnect()
  }, [
    activeMessageState?.error,
    activeMessageState?.hasMore,
    activeMessageState?.isLoading,
    activeSessionId,
    activeSessionTargetId,
    loadMoreMessages,
  ])

  useEffect(() => {
    const container = messagesContainerRef.current
    if (!container) return
    if (!hasUserScrolledRef.current) return
    if (container.scrollHeight <= container.clientHeight + 1) {
      void loadMoreMessages()
    }
  }, [chatList, loadMoreMessages])

  useEffect(() => {
    const container = messagesContainerRef.current
    if (!container) return
    const pending = pendingScrollAdjustmentRef.current
    if (pending && pending.sessionId === activeSessionId) {
      const nextTop = pending.previousTop + (container.scrollHeight - pending.previousHeight)
      container.scrollTop = nextTop
      pendingScrollAdjustmentRef.current = null
      return
    }
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [activeSessionId, chatList])

  useEffect(() => {
    if (!user || !currentUserId) {
      return
    }

    const wsUrl = buildWsUrl(API_BASE_URL)
    if (!wsUrl) return

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws
    setConnectionStatus('connecting')

    ws.onopen = () => setConnectionStatus('connected')
    ws.onclose = () => setConnectionStatus('disconnected')
    ws.onerror = () => setConnectionStatus('error')

    ws.onmessage = (event) => {
      let payload: unknown = null
      try {
        payload = JSON.parse(event.data)
      } catch {
        return
      }

      const incoming = extractWsMessages(payload)
      if (incoming.length === 0) return

      incoming.forEach((raw) => {
        const sessionId = normalizeId(raw.session_id)
        if (!sessionId) return

        const message = normalizeMessage(raw, currentUserId)
        if (!message) return

        const fromId = normalizeId(raw.message_from)
        const toId = normalizeId(raw.message_to)
        const otherId = fromId && fromId === currentUserId ? toId : fromId
        const sessionType = Number.isFinite(Number(raw.session_type)) ? Number(raw.session_type) : SESSION_TYPE_PRIVATE
        const isActive = activeIdRef.current === sessionId

        const confirmations = mergeIncomingMessages(sessionId, [message], {
          allowTimelessConfirmation: true,
        })
        const confirmedLocalId = confirmations[0]?.localId

        setSessions((prev) => {
          const index = prev.findIndex((session) => session.id === sessionId)
          const baseSession =
            index >= 0
              ? prev[index]
              : {
                  id: sessionId,
                  targetId: otherId,
                  name: otherId ? `用户 ${otherId}` : '新会话',
                  avatar: '',
                  lastMessage: '',
                  lastMessageTime: undefined,
                  unread: 0,
                  sessionType,
                }
          const nextUnread = isActive
            ? 0
            : message.from === 'other'
              ? baseSession.unread + 1
              : baseSession.unread
          const keepNewerLocalSummary =
            message.from === 'me' &&
            Boolean(confirmedLocalId) &&
            Boolean(baseSession.lastLocalMessageId) &&
            baseSession.lastLocalMessageId !== confirmedLocalId
          const updated = {
            ...baseSession,
            targetId: baseSession.targetId || otherId,
            lastMessage: keepNewerLocalSummary
              ? baseSession.lastMessage
              : summarizeMessage(message.content),
            lastMessageTime: keepNewerLocalSummary
              ? baseSession.lastMessageTime
              : message.createdAt,
            lastMessageDeliveryStatus: keepNewerLocalSummary
              ? baseSession.lastMessageDeliveryStatus
              : undefined,
            lastLocalMessageId: keepNewerLocalSummary
              ? baseSession.lastLocalMessageId
              : undefined,
            unread: nextUnread,
            sessionType,
          }
          const next = index >= 0 ? prev.filter((_, idx) => idx !== index) : [...prev]
          return [updated, ...next]
        })

        if (message.from === 'other' && isActive) {
          sendReadAck(sessionId)
        }

        const hasSession = sessionsRef.current.some((session) => session.id === sessionId)
        if (!hasSession && otherId) {
          void fetchOrCreateSession(otherId, undefined, { activate: false, notify: false })
        }
      })
    }

    return () => {
      ws.close()
      wsRef.current = null
    }
  }, [currentUserId, fetchOrCreateSession, mergeIncomingMessages, sendReadAck, user])

  const handleSelectSession = useCallback((sessionId: string) => {
    const hadUnread = sessionsRef.current.find((session) => session.id === sessionId)?.unread ?? 0
    setActiveId(sessionId)
    setSessions((prev) =>
      prev.map((session) => (session.id === sessionId ? { ...session, unread: 0 } : session))
    )
    if (hadUnread > 0) {
      sendReadAck(sessionId)
    }
  }, [sendReadAck])

  const handleSend = (e: FormEvent) => {
    e.preventDefault()
    setSendError(null)
    const trimmed = input.trim()
    if (!trimmed) return

    if (!activeId || !activeSession || !currentUserId) return
    const ws = wsRef.current
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setSendError('实时连接未就绪，请稍后再试')
      return
    }

    const sessionId = normalizeId(activeId)
    if (!sessionId) return

    const payload = {
      biz_type: 'session',
      biz_data: {
        type: 'message',
        session_id: sessionId,
        session_type: activeSession.sessionType,
        message_to: activeSession.targetId,
        content: trimmed,
      },
    }

    try {
      ws.send(JSON.stringify(payload))
    } catch (err) {
      setSendError('消息发送失败，请稍后重试')
      return
    }

    const createdAt = new Date().toISOString()
    localMessageSequenceRef.current += 1
    const newMsg: ChatMessage = {
      id: `local-${Date.now()}-${localMessageSequenceRef.current}`,
      from: 'me',
      content: trimmed,
      createdAt,
      deliveryStatus: 'pending',
    }

    mergeIncomingMessages(sessionId, [newMsg])
    setSessions((prev) => {
      const index = prev.findIndex((session) => session.id === sessionId)
      if (index === -1) return prev
      const updated = {
        ...prev[index],
        lastMessage: summarizeMessage(trimmed),
        lastMessageTime: createdAt,
        lastMessageDeliveryStatus: 'pending' as const,
        lastLocalMessageId: newMsg.id,
        unread: 0,
      }
      const next = [...prev]
      next.splice(index, 1)
      return [updated, ...next]
    })
    const timerId = window.setTimeout(
      () => markMessageUnconfirmed(sessionId, newMsg.id),
      MESSAGE_CONFIRM_TIMEOUT_MS
    )
    confirmationTimersRef.current.set(newMsg.id, timerId)
    setInput('')
  }

  const handleDeleteSession = useCallback(
    async (session: Conversation, event?: MouseEvent<HTMLButtonElement>) => {
      event?.stopPropagation()
      if (!session.id || deletingSessionIds[session.id]) return
      const displayName = session.name || '该用户'
      const confirmed = window.confirm(`确定删除与 ${displayName} 的会话吗？`)
      if (!confirmed) return

      setDeletingSessionIds((prev) => ({ ...prev, [session.id]: true }))
      setError(null)

      try {
        await apiPost(`/sessions/${encodeURIComponent(session.id)}/delete`, null)
        setSessions((prev) => prev.filter((item) => item.id !== session.id))
        const currentMessages = messagesBySessionRef.current[session.id] ?? []
        currentMessages.forEach((message) => clearConfirmationTimer(message.id))
        const nextMessagesBySession = { ...messagesBySessionRef.current }
        delete nextMessagesBySession[session.id]
        messagesBySessionRef.current = nextMessagesBySession
        setMessagesBySession(nextMessagesBySession)
        setMessagePageBySession((prev) => {
          if (!(session.id in prev)) return prev
          const next = { ...prev }
          delete next[session.id]
          return next
        })
        setActiveId((prev) => (prev === session.id ? '' : prev))
      } catch (err) {
        const message = err instanceof Error ? err.message : '删除会话失败'
        setError(message)
      } finally {
        setDeletingSessionIds((prev) => {
          const next = { ...prev }
          delete next[session.id]
          return next
        })
      }
    },
    [clearConfirmationTimer, deletingSessionIds]
  )

  let lastShownTimestamp = 0

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <Link
          to="/"
          className="inline-flex items-center space-x-2 text-gray-600 hover:text-primary-600 transition-colors"
        >
          <ArrowLeft className="h-5 w-5" />
          <span>返回首页</span>
        </Link>
        <div className="text-sm text-gray-500 flex items-center space-x-2">
          <MessageCircle className="h-4 w-4 text-primary-600" />
          <span>本服务由 Go-Chatery 提供</span>
          <span
            className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium ring-1 ${connectionBadge.bg} ${connectionBadge.text} ${connectionBadge.ring}`}
          >
            <span
              className={`h-1.5 w-1.5 rounded-full ${connectionBadge.dot} ${
                connectionBadge.pulse ? 'animate-pulse' : ''
              }`}
            />
            {connectionBadge.label}
          </span>
        </div>
      </div>

      <div className="card grid md:grid-cols-[260px_1fr] gap-4 h-[70vh] overflow-hidden">
        <div className="border-r border-gray-100 pr-2 space-y-2 min-h-0 overflow-y-auto">
          {isLoading && (
            <div className="text-sm text-gray-500 px-3 py-2">加载会话中...</div>
          )}
          {!isLoading && !user && (
            <div className="text-sm text-gray-500 px-3 py-2">登录后查看私信</div>
          )}
          {!isLoading && user && error && (
            <div className="text-sm text-red-600 px-3 py-2 flex items-center justify-between gap-2">
              <span className="flex-1 break-words">{error}</span>
              <button
                type="button"
                onClick={() => setRefreshKey((prev) => prev + 1)}
                className="text-xs text-primary-600 hover:text-primary-700"
              >
                重试
              </button>
            </div>
          )}
          {!isLoading && user && !error && sessions.length === 0 && (
            <div className="text-sm text-gray-500 px-3 py-2">暂无会话</div>
          )}
          {!isLoading &&
            !error &&
            sessions.map((session) => {
              const isActive = session.id === activeId
              const isDeleting = Boolean(deletingSessionIds[session.id])
              const lastMessageStatus =
                session.lastMessageDeliveryStatus === 'pending'
                  ? '发送中'
                  : session.lastMessageDeliveryStatus === 'failed'
                    ? '未确认'
                    : ''
              const lastMessageTime = session.lastMessage
                ? formatRelativeTime(session.lastMessageTime, '时间未提供')
                : ''
              return (
                <div
                  key={session.id}
                  className={`w-full flex items-center gap-2 rounded-lg px-3 py-2 transition-colors ${
                    isActive ? 'bg-primary-50 text-primary-700' : 'hover:bg-gray-50'
                  }`}
                >
                  <button
                    type="button"
                    onClick={() => handleSelectSession(session.id)}
                    className="flex flex-1 min-w-0 items-center space-x-3 text-left"
                  >
                    <UserAvatar
                      avatar={session.avatar}
                      name={session.name}
                      userId={session.targetId || session.id}
                      fallbackSeed={`msg-${session.targetId || session.name || session.id}`}
                      className="w-10 h-10 rounded-full"
                    />
                    <div className="flex-1 min-w-0">
                    <div className="flex items-center">
                      <p className="font-medium text-sm line-clamp-1">{session.name}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <p
                        className={`text-xs line-clamp-1 flex-1 ${
                          session.lastMessageDeliveryStatus === 'failed'
                            ? 'text-red-600'
                            : session.lastMessageDeliveryStatus === 'pending'
                              ? 'text-amber-600'
                              : 'text-gray-500'
                        }`}
                      >
                        {session.lastMessage || '暂无消息'}
                        {lastMessageStatus ? ` · ${lastMessageStatus}` : ''}
                      </p>
                      <div className="flex items-center gap-2 shrink-0">
                        {lastMessageTime && (
                          <span className="text-[10px] text-gray-400">
                            {lastMessageTime}
                          </span>
                        )}
                        {session.unread > 0 ? (
                          <span
                            className="w-2 h-2 rounded-full bg-red-500"
                            aria-label="未读"
                            title="未读"
                          />
                        ) : null}
                      </div>
                    </div>
                  </div>
                </button>
                  <button
                    type="button"
                    onClick={(event) => handleDeleteSession(session, event)}
                    disabled={isDeleting}
                    aria-label={`删除与 ${session.name || '该用户'} 的会话`}
                    title="删除会话"
                    className={`shrink-0 rounded-md p-1 text-gray-400 transition-colors hover:text-red-500 ${
                      isDeleting ? 'cursor-not-allowed opacity-50' : ''
                    }`}
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              )
            })}
        </div>

        <div className="flex flex-col min-h-0">
          <div className="flex items-center justify-between pb-3 border-b border-gray-100">
            <div className="flex items-center space-x-3">
              {activeSession ? (
                <UserAvatar
                  avatar={activeSession.avatar}
                  name={activeSession.name}
                  userId={activeSession.targetId || activeSession.id}
                  fallbackSeed={`msg-${activeSession.targetId || activeSession.name || activeSession.id}`}
                  className="w-10 h-10 rounded-full"
                />
              ) : (
                <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
                  <MessageCircle className="h-5 w-5 text-gray-400" />
                </div>
              )}
              <div>
                <p className="font-semibold text-gray-900">
                  {activeSession?.name || '请选择会话'}
                </p>
                <p className="text-xs text-gray-500">
                  {activeSession ? '会话进行中' : '从左侧选择一个会话'}
                </p>
              </div>
            </div>
            {user && (
              <div className="text-xs text-gray-500">我：{user.name}</div>
            )}
          </div>

          <div
            ref={messagesContainerRef}
            onScroll={handleMessagesScroll}
            onWheel={handleMessagesWheel}
            onTouchMove={handleMessagesTouchMove}
            className="flex-1 overflow-y-auto overscroll-y-contain space-y-3 py-3 pr-2 min-h-0"
          >
            <div ref={historyObserverTargetRef} className="h-1" />
            {!activeSession && (
              <div className="text-sm text-gray-500 text-center py-8">
                选择一个会话开始聊天
              </div>
            )}
            {activeSession && activeMessageState?.error && (
              <div className="text-sm text-red-600 text-center py-2">
                {activeMessageState.error}
              </div>
            )}
            {activeSession && activeMessageState?.isLoading && chatList.length === 0 && (
              <div className="text-sm text-gray-500 text-center py-8">加载消息中...</div>
            )}
            {activeSession && activeMessageState?.hasMore && (
              <div className="flex justify-center text-xs text-gray-500">
                {activeMessageState.isLoading ? '加载中...' : '上滑加载更多'}
              </div>
            )}
            {activeSession &&
              activeMessageState &&
              !activeMessageState.hasMore &&
              !activeMessageState.isLoading &&
              !activeMessageState.error &&
              chatList.length > 0 && (
                <div className="flex justify-center text-xs text-gray-400">没有更多历史记录</div>
              )}
            {activeSession && chatList.length === 0 && (
              <div className="text-sm text-gray-500 text-center py-8">暂无聊天记录</div>
            )}
            {activeSession &&
              chatList.map((msg) => {
                const isMe = msg.from === 'me'
                const messageTimestamp = toTimestamp(msg.createdAt)
                const shouldShowTime =
                  messageTimestamp > 0 &&
                  (lastShownTimestamp === 0 || messageTimestamp - lastShownTimestamp >= MESSAGE_TIME_GAP_MS)
                if (shouldShowTime) {
                  lastShownTimestamp = messageTimestamp
                }
                const messageTime = shouldShowTime ? formatRelativeTime(msg.createdAt, '') : ''
                const deliveryLabel =
                  msg.deliveryStatus === 'pending'
                    ? '等待后端确认'
                    : msg.deliveryStatus === 'failed'
                      ? '未获后端确认'
                      : ''
                const metadataLabel =
                  deliveryLabel || (messageTimestamp <= 0 ? '时间未提供' : messageTime)
                const metadataColor =
                  msg.deliveryStatus === 'failed'
                    ? 'text-red-500'
                    : msg.deliveryStatus === 'pending'
                      ? 'text-amber-500'
                      : 'text-gray-400'
                return (
                  <div key={msg.id} className={`flex ${isMe ? 'justify-end' : 'justify-start'}`}>
                    {!isMe && (
                      <UserAvatar
                        avatar={activeSession.avatar}
                        name={activeSession.name}
                        userId={activeSession.targetId || activeSession.id}
                        fallbackSeed={`msg-${activeSession.targetId || activeSession.name || activeSession.id}`}
                        className="w-8 h-8 rounded-full mr-2"
                      />
                    )}
                    <div className={`flex flex-col ${isMe ? 'items-end' : 'items-start'}`}>
                      <div
                        className={`max-w-[30ch] break-words rounded-2xl px-3 py-2 text-sm ${
                          isMe ? 'bg-primary-600 text-white rounded-br-sm' : 'bg-gray-100 text-gray-900 rounded-bl-sm'
                        } ${msg.deliveryStatus === 'failed' ? 'ring-2 ring-red-300' : ''}`}
                      >
                        <span className={msg.deliveryStatus === 'pending' ? 'opacity-70' : ''}>
                          {msg.content}
                        </span>
                      </div>
                      {metadataLabel ? (
                        <span className={`mt-1 text-[10px] ${metadataColor}`}>
                          {metadataLabel}
                        </span>
                      ) : null}
                    </div>
                    {isMe && (
                      <UserAvatar
                        avatar={user?.avatar}
                        name={user?.name || '我'}
                        userId={user?.id}
                        fallbackSeed={user?.name || 'me'}
                        alt="me"
                        className="w-8 h-8 rounded-full ml-2"
                      />
                    )}
                  </div>
                )
              })}
            {activeSession && <div ref={messagesEndRef} />}
          </div>

          <form onSubmit={handleSend} className="pt-3 border-t border-gray-100">
            <div className="flex items-center space-x-3">
              <textarea
                value={input}
                onChange={(e) => {
                  setInput(e.target.value)
                  setSendError(null)
                }}
                placeholder={activeSession ? '输入消息...' : '请选择会话'}
                rows={2}
                className="textarea flex-1 resize-none"
              />
              <button
                type="submit"
                disabled={!input.trim() || !activeId || !user || connectionStatus !== 'connected'}
                className="btn-primary flex items-center space-x-2"
              >
                <Send className="h-4 w-4" />
                <span>发送</span>
              </button>
            </div>
            {sendError && (
              <div className="text-xs text-red-600 mt-2">{sendError}</div>
            )}
          </form>
        </div>
      </div>
    </div>
  )
}
