import { useState, useMemo, useRef, useEffect, FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Send, MessageCircle } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { apiGet } from '../utils/api'

type SessionResponse = {
  session_id?: string | number
  target_id?: string | number
  target_name?: string
  target_avatar?: string
  last_message?: string
  last_message_time?: string
  unread_count?: number
}

type Conversation = {
  id: string
  targetId: string
  name: string
  avatar: string
  lastMessage: string
  lastMessageTime?: string
  unread: number
}

type ChatMessage = {
  id: string
  from: 'me' | 'other'
  content: string
}

const normalizeSession = (item: SessionResponse): Conversation => {
  const name = typeof item.target_name === 'string' ? item.target_name.trim() : ''
  const avatar = typeof item.target_avatar === 'string' ? item.target_avatar.trim() : ''
  const lastMessage = typeof item.last_message === 'string' ? item.last_message.trim() : ''
  const lastMessageTime =
    typeof item.last_message_time === 'string' && item.last_message_time
      ? item.last_message_time
      : undefined
  const unreadCount = Number.isFinite(Number(item.unread_count)) ? Number(item.unread_count) : 0

  return {
    id: String(item.session_id ?? ''),
    targetId: String(item.target_id ?? ''),
    name,
    avatar,
    lastMessage,
    lastMessageTime,
    unread: unreadCount,
  }
}

export default function Messages() {
  const { user } = useAuth()
  const [sessions, setSessions] = useState<Conversation[]>([])
  const [activeId, setActiveId] = useState<string>('')
  const [messagesBySession, setMessagesBySession] = useState<Record<string, ChatMessage[]>>({})
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [refreshKey, setRefreshKey] = useState(0)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const activeSession = useMemo(
    () => sessions.find((session) => session.id === activeId) ?? null,
    [sessions, activeId]
  )
  const chatList = useMemo(() => {
    if (!activeId) return []
    return messagesBySession[activeId] || []
  }, [messagesBySession, activeId])

  useEffect(() => {
    if (!user) {
      setSessions([])
      setActiveId('')
      setMessagesBySession({})
      setIsLoading(false)
      setError(null)
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
          .map(normalizeSession)
          .filter((item) => item.id && item.name)

        setSessions(normalized)
        setActiveId((prev) => {
          if (prev && normalized.some((session) => session.id === prev)) {
            return prev
          }
          return normalized[0]?.id ?? ''
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
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [chatList])

  const handleSend = (e: FormEvent) => {
    e.preventDefault()
    const trimmed = input.trim()
    if (!trimmed) return

    if (!activeId) return

    const newMsg: ChatMessage = {
      id: `${Date.now()}`,
      from: 'me',
      content: trimmed,
    }

    setMessagesBySession((prev) => {
      const current = prev[activeId] ?? []
      return {
        ...prev,
        [activeId]: [...current, newMsg],
      }
    })
    setSessions((prev) =>
      prev.map((session) =>
        session.id === activeId
          ? {
              ...session,
              lastMessage: trimmed,
              lastMessageTime: new Date().toISOString(),
              unread: 0,
            }
          : session
      )
    )
    setInput('')
  }

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
        </div>
      </div>

      <div className="card grid md:grid-cols-[260px_1fr] gap-4 min-h-[70vh]">
        <div className="border-r border-gray-100 pr-2 space-y-2">
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
              const avatarUrl =
                session.avatar ||
                `https://api.dicebear.com/7.x/avataaars/svg?seed=msg-${session.targetId || session.name || session.id}`
              return (
                <button
                  key={session.id}
                  onClick={() => setActiveId(session.id)}
                  className={`w-full flex items-center space-x-3 rounded-lg px-3 py-2 text-left transition-colors ${
                    isActive ? 'bg-primary-50 text-primary-700' : 'hover:bg-gray-50'
                  }`}
                >
                  <img
                    src={avatarUrl}
                    alt={session.name}
                    className="w-10 h-10 rounded-full"
                  />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-sm line-clamp-1">{session.name}</p>
                    <p className="text-xs text-gray-500 line-clamp-1">
                      {session.lastMessage || '暂无消息'}
                    </p>
                  </div>
                  {session.unread > 0 ? (
                    <span className="text-[10px] px-2 py-0.5 rounded-full bg-red-100 text-red-600">
                      {session.unread}
                    </span>
                  ) : null}
                </button>
              )
            })}
        </div>

        <div className="flex flex-col">
          <div className="flex items-center justify-between pb-3 border-b border-gray-100">
            <div className="flex items-center space-x-3">
              {activeSession ? (
                <img
                  src={
                    activeSession.avatar ||
                    `https://api.dicebear.com/7.x/avataaars/svg?seed=msg-${activeSession.targetId || activeSession.name || activeSession.id}`
                  }
                  alt={activeSession.name}
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

          <div className="flex-1 overflow-y-auto space-y-3 py-3 pr-2">
            {!activeSession && (
              <div className="text-sm text-gray-500 text-center py-8">
                选择一个会话开始聊天
              </div>
            )}
            {activeSession && chatList.length === 0 && (
              <div className="text-sm text-gray-500 text-center py-8">暂无消息</div>
            )}
            {activeSession &&
              chatList.map((msg) => {
                const isMe = msg.from === 'me'
                const otherAvatar =
                  activeSession.avatar ||
                  `https://api.dicebear.com/7.x/avataaars/svg?seed=msg-${activeSession.targetId || activeSession.name || activeSession.id}`
                return (
                  <div key={msg.id} className={`flex ${isMe ? 'justify-end' : 'justify-start'}`}>
                    {!isMe && (
                      <img
                        src={otherAvatar}
                        alt={activeSession.name}
                        className="w-8 h-8 rounded-full mr-2"
                      />
                    )}
                    <div
                      className={`max-w-[70%] rounded-2xl px-3 py-2 text-sm ${
                        isMe ? 'bg-primary-600 text-white rounded-br-sm' : 'bg-gray-100 text-gray-900 rounded-bl-sm'
                      }`}
                    >
                      {msg.content}
                    </div>
                    {isMe && (
                      <img
                        src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${user?.name || 'me'}`}
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
                onChange={(e) => setInput(e.target.value)}
                placeholder={activeSession ? '输入消息...' : '请选择会话'}
                rows={2}
                className="textarea flex-1 resize-none"
              />
              <button
                type="submit"
                disabled={!input.trim() || !activeId || !user}
                className="btn-primary flex items-center space-x-2"
              >
                <Send className="h-4 w-4" />
                <span>发送</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
