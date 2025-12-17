import { useState, useMemo, useRef, useEffect, FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Send, MessageCircle } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

type Conversation = {
  id: number
  name: string
  lastMessage: string
  unread?: number
}

type ChatMessage = {
  id: string
  from: 'me' | 'other'
  content: string
}

const mockConversations: Conversation[] = [
  { id: 1, name: '前端小能手', lastMessage: '这版 UI 很赞！', unread: 2 },
  { id: 2, name: 'Go 语言爱好者', lastMessage: '今晚一起讨论下网关？' },
  { id: 3, name: '设计灵感库', lastMessage: '发你几套配色参考' },
  { id: 4, name: '后端老王', lastMessage: '日志规范要补充下' },
]

const mockChat: Record<number, ChatMessage[]> = {
  1: [
    { id: '1', from: 'other', content: '这版 UI 很赞！' },
    { id: '2', from: 'me', content: '谢谢，后面我再调下行距。' },
  ],
  2: [
    { id: '1', from: 'other', content: '今晚一起讨论下网关？' },
    { id: '2', from: 'me', content: '可以，8 点后有空。' },
  ],
}

export default function Messages() {
  const { author } = useAuth()
  const [activeId, setActiveId] = useState<number>(mockConversations[0]?.id ?? 0)
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const chatList = useMemo(() => mockChat[activeId] || [], [activeId])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [chatList])

  const handleSend = (e: FormEvent) => {
    e.preventDefault()
    const trimmed = input.trim()
    if (!trimmed) return

    const newMsg: ChatMessage = {
      id: `${Date.now()}`,
      from: 'me',
      content: trimmed,
    }
    mockChat[activeId] = [...chatList, newMsg]
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
          {mockConversations.map((c) => {
            const isActive = c.id === activeId
            return (
              <button
                key={c.id}
                onClick={() => setActiveId(c.id)}
                className={`w-full flex items-center space-x-3 rounded-lg px-3 py-2 text-left transition-colors ${
                  isActive ? 'bg-primary-50 text-primary-700' : 'hover:bg-gray-50'
                }`}
              >
                <img
                  src={`https://api.dicebear.com/7.x/avataaars/svg?seed=msg-${c.id}`}
                  alt={c.name}
                  className="w-10 h-10 rounded-full"
                />
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-sm line-clamp-1">{c.name}</p>
                  <p className="text-xs text-gray-500 line-clamp-1">{c.lastMessage}</p>
                </div>
                {c.unread ? (
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-red-100 text-red-600">
                    {c.unread}
                  </span>
                ) : null}
              </button>
            )
          })}
        </div>

        <div className="flex flex-col">
          <div className="flex items-center justify-between pb-3 border-b border-gray-100">
            <div className="flex items-center space-x-3">
              <img
                src={`https://api.dicebear.com/7.x/avataaars/svg?seed=msg-${activeId}`}
                alt="chat-author"
                className="w-10 h-10 rounded-full"
              />
              <div>
                <p className="font-semibold text-gray-900">
                  {mockConversations.find(c => c.id === activeId)?.name || '会话'}
                </p>
                <p className="text-xs text-gray-500">示例会话</p>
              </div>
            </div>
            {author && (
              <div className="text-xs text-gray-500">我：{author.name}</div>
            )}
          </div>

          <div className="flex-1 overflow-y-auto space-y-3 py-3 pr-2">
            {chatList.map(msg => {
              const isMe = msg.from === 'me'
              return (
                <div key={msg.id} className={`flex ${isMe ? 'justify-end' : 'justify-start'}`}>
                  {!isMe && (
                    <img
                      src={`https://api.dicebear.com/7.x/avataaars/svg?seed=msg-${activeId}`}
                      alt="other"
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
                      src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${author?.name || 'me'}`}
                      alt="me"
                      className="w-8 h-8 rounded-full ml-2"
                    />
                  )}
                </div>
              )
            })}
            <div ref={messagesEndRef} />
          </div>

          <form onSubmit={handleSend} className="pt-3 border-t border-gray-100">
            <div className="flex items-center space-x-3">
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="输入消息..."
                rows={2}
                className="textarea flex-1 resize-none"
              />
              <button
                type="submit"
                disabled={!input.trim()}
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
