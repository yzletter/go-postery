import { FormEvent, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Send, Sparkles } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { apiPost } from '../utils/api'

type ChatMessage = {
  id: string
  role: 'user' | 'ai'
  content: string
  documents?: AgentDocument[]
}

type AgentDocument = {
  title?: string
  url?: string
  snippet?: string
  source?: string
  id?: string
}

const buildSessionId = () => {
  const now = BigInt(Date.now())
  const rand = BigInt(Math.floor(Math.random() * 1_000_000))
  return (now * 1_000_000n + rand).toString()
}

const asString = (value: unknown) => {
  if (typeof value === 'string') {
    const trimmed = value.trim()
    return trimmed ? trimmed : undefined
  }
  if (typeof value === 'number') return Number.isFinite(value) ? String(value) : undefined
  return undefined
}

const normalizeDocuments = (raw: unknown): AgentDocument[] => {
  if (!raw) return []
  const list = Array.isArray(raw) ? raw : [raw]
  const normalized = list.map((item) => {
    if (typeof item === 'string') {
      const title = item.trim()
      return title ? { title } : null
    }
    if (!item || typeof item !== 'object') {
      const title = asString(item)
      return title ? { title } : null
    }
    const record = item as Record<string, unknown>
    return {
      title:
        asString(record.title) ??
        asString(record.name) ??
        asString(record.doc_title) ??
        asString(record.document) ??
        asString(record.filename),
      url: asString(record.url) ?? asString(record.link) ?? asString(record.source_url),
      snippet:
        asString(record.snippet) ??
        asString(record.summary) ??
        asString(record.abstract) ??
        asString(record.content),
      source: asString(record.source) ?? asString(record.provider) ?? asString(record.origin),
      id: asString(record.id) ?? asString(record.document_id),
    }
  })

  return normalized.filter((doc): doc is AgentDocument => Boolean(doc))
}

const extractAgentPayload = (payload: unknown) => {
  if (typeof payload === 'string') {
    return { reply: payload, documents: [] as AgentDocument[] }
  }
  if (!payload || typeof payload !== 'object') {
    return { reply: '', documents: [] as AgentDocument[] }
  }
  const record = payload as Record<string, unknown>
  const candidates = [record.reply, record.answer, record.content, record.message, record.result]
  let reply = ''
  for (const candidate of candidates) {
    const value = asString(candidate)
    if (value) {
      reply = value
      break
    }
  }
  const documents = normalizeDocuments(
    record.documents ?? record.document ?? record.references ?? record.reference ?? record.sources ?? record.source
  )
  return { reply, documents }
}

const buildDocumentLabel = (doc: AgentDocument, index: number) => {
  return doc.title || doc.url || doc.id || `文献 ${index + 1}`
}

export default function Agent() {
  const { user } = useAuth()
  const [messages, setMessages] = useState<ChatMessage[]>([
    { id: 'welcome', role: 'ai', content: '嗨，我是你的 AI Agent，有什么可以帮你？' },
  ])
  const [input, setInput] = useState('')
  const [isThinking, setIsThinking] = useState(false)
  const sessionIdRef = useRef<string>(buildSessionId())
  const bottomRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const trimmed = input.trim()
    if (!trimmed || isThinking) return

    const userMsg: ChatMessage = {
      id: `u-${Date.now()}`,
      role: 'user',
      content: trimmed,
    }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setIsThinking(true)

    try {
      const { data } = await apiPost<unknown>('/agent/chat', {
        session_id: sessionIdRef.current,
        query: trimmed,
      })
      const { reply, documents } = extractAgentPayload(data)
      setMessages(prev => [
        ...prev,
        {
          id: `a-${Date.now()}`,
          role: 'ai',
          content: reply || '暂时没有返回内容。',
          documents,
        },
      ])
    } catch (error) {
      console.error('Agent chat failed:', error)
      setMessages(prev => [
        ...prev,
        {
          id: `a-${Date.now()}`,
          role: 'ai',
          content: '抱歉，Agent 暂时不可用，请稍后再试。',
        },
      ])
    } finally {
      setIsThinking(false)
    }
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
        <div className="inline-flex items-center space-x-2 text-sm text-gray-500">
          <Sparkles className="h-4 w-4 text-primary-600" />
          <span>Agent 聊天</span>
        </div>
      </div>

      <div className="card min-h-[75vh] flex flex-col">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">AI Agent</h1>
            <p className="text-sm text-gray-500">与智能助手快速沟通</p>
          </div>
          {user && (
            <div className="flex items-center space-x-2 text-sm text-gray-600">
              <img
                src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${user.name}`}
                alt={user.name}
                className="w-8 h-8 rounded-full"
              />
              <span>{user.name}</span>
            </div>
          )}
        </div>

        <div className="flex-1 overflow-y-auto space-y-4 pr-2">
          {messages.map(msg => {
            const isUser = msg.role === 'user'
            return (
              <div
                key={msg.id}
                className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}
              >
                {!isUser && (
                  <img
                    src="https://api.dicebear.com/7.x/bottts/svg?seed=ai-agent"
                    alt="AI"
                    className="w-9 h-9 rounded-full mr-3"
                  />
                )}
                <div
                  className={`max-w-[70%] rounded-2xl px-4 py-2 shadow-sm ${
                    isUser
                      ? 'bg-primary-600 text-white rounded-br-sm'
                      : 'bg-gray-100 text-gray-900 rounded-bl-sm'
                  }`}
                >
                  <p className="whitespace-pre-wrap break-words leading-relaxed">{msg.content}</p>
                  {!isUser && msg.documents && msg.documents.length > 0 && (
                    <div className="mt-3 border-t border-gray-200/70 pt-3 text-xs text-gray-600">
                      <p className="mb-2 text-[11px] uppercase tracking-wide text-gray-400">引用文献</p>
                      <ol className="list-decimal space-y-1 pl-4">
                        {msg.documents.map((doc, index) => {
                          const label = buildDocumentLabel(doc, index)
                          return (
                            <li key={`${label}-${index}`} className="leading-relaxed">
                              {doc.url ? (
                                <a
                                  href={doc.url}
                                  target="_blank"
                                  rel="noreferrer"
                                  className="text-primary-600 hover:text-primary-700 underline"
                                >
                                  {label}
                                </a>
                              ) : (
                                <span>{label}</span>
                              )}
                              {doc.source && <span className="text-gray-400"> · {doc.source}</span>}
                              {doc.snippet && (
                                <span className="block text-[11px] text-gray-400">
                                  {doc.snippet}
                                </span>
                              )}
                            </li>
                          )
                        })}
                      </ol>
                    </div>
                  )}
                </div>
                {isUser && (
                  <img
                    src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${user?.name || 'user'}`}
                    alt="Me"
                    className="w-9 h-9 rounded-full ml-3"
                  />
                )}
              </div>
            )
          })}
          {isThinking && (
            <div className="flex items-center space-x-2 text-sm text-gray-500">
              <div className="w-5 h-5 border-2 border-primary-600 border-t-transparent rounded-full animate-spin" />
              <span>Agent 正在思考...</span>
            </div>
          )}
          <div ref={bottomRef} />
        </div>

        <form onSubmit={handleSubmit} className="mt-4 border-t border-gray-100 pt-4">
          <div className="flex items-center space-x-3">
            <textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="输入你的问题或需求..."
              rows={1}
              className="textarea flex-1 resize-none min-h-[44px]"
            />
            <button
              type="submit"
              disabled={!input.trim() || isThinking}
              className="btn-primary flex items-center space-x-2"
            >
              <Send className="h-4 w-4" />
              <span>发送</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
