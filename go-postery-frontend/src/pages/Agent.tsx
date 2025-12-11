import { FormEvent, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Send, Sparkles } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

type ChatMessage = {
  id: string
  role: 'user' | 'ai'
  content: string
}

export default function Agent() {
  const { user } = useAuth()
  const [messages, setMessages] = useState<ChatMessage[]>([
    { id: 'welcome', role: 'ai', content: '嗨，我是你的 AI Agent，有什么可以帮你？' },
  ])
  const [input, setInput] = useState('')
  const [isThinking, setIsThinking] = useState(false)
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
    if (!trimmed) return

    const userMsg: ChatMessage = {
      id: `u-${Date.now()}`,
      role: 'user',
      content: trimmed,
    }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setIsThinking(true)

    // 简单模拟 AI 回复
    setTimeout(() => {
      setMessages(prev => [
        ...prev,
        {
          id: `a-${Date.now()}`,
          role: 'ai',
          content: `已收到：${trimmed}`,
        },
      ])
      setIsThinking(false)
    }, 600)
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
