import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ChangeEvent,
  type DragEvent,
  type FormEvent,
} from 'react'
import { Link } from 'react-router-dom'
import {
  AlertCircle,
  ArrowLeft,
  BookOpenCheck,
  Bot,
  BrainCircuit,
  CheckCircle2,
  FileText,
  FileUp,
  GraduationCap,
  Loader2,
  Play,
  RotateCcw,
  Send,
  Sparkles,
  StopCircle,
  Target,
  UploadCloud,
  User,
  Wifi,
  WifiOff,
  X,
} from 'lucide-react'
import UserAvatar from '../components/UserAvatar'
import { useAuth } from '../contexts/AuthContext'
import { getAuthToken } from '../utils/api'
import {
  buildInterviewEnvelope,
  buildInterviewWsUrl,
  parseInterviewServerEvent,
  uploadQuestionBank,
} from './interview/api'
import InterviewMessageItem from './interview/InterviewMessageItem'
import InterviewStageBar from './interview/InterviewStageBar'
import type {
  InterviewClientEvent,
  InterviewConnectionStatus,
  InterviewMessage,
  InterviewServerEvent,
} from './interview/types'

const MAX_SOURCE_CHARS = 120_000
const MAX_SOURCE_FILE_SIZE = 512 * 1024
const MAX_QUESTION_FILE_SIZE = 5 * 1024 * 1024
const START_RESPONSE_TIMEOUT_MS = 3 * 60 * 1000
const ANSWER_RESPONSE_TIMEOUT_MS = 3 * 60 * 1000
const CANCEL_RESPONSE_TIMEOUT_MS = 2 * 60 * 1000
const TEXT_FILE_EXTENSIONS = new Set(['md', 'markdown', 'txt'])

type UploadState = {
  status: 'idle' | 'uploading' | 'success' | 'error'
  message: string
}

type PendingAction = 'answer' | 'cancel' | 'start'

type PendingActionRequest = {
  action: PendingAction
  sessionId: string
  messageId?: string
}

type SourceFieldProps = {
  id: string
  title: string
  description: string
  placeholder: string
  value: string
  fileName: string
  disabled: boolean
  onChange: (value: string) => void
  onFileSelect: (file: File) => void
  onClearFile: () => void
}

const getFileExtension = (fileName: string) => {
  const segments = fileName.toLowerCase().split('.')
  return segments.length > 1 ? segments.pop() || '' : ''
}

const readSourceFile = async (file: File) => {
  if (!TEXT_FILE_EXTENSIONS.has(getFileExtension(file.name))) {
    throw new Error('仅支持导入 TXT 或 Markdown 文本文件')
  }
  if (file.size > MAX_SOURCE_FILE_SIZE) {
    throw new Error('文本文件不能超过 512KB')
  }

  const content = (await file.text()).replace(/^\uFEFF/, '')
  if (!content.trim() || content.includes('\u0000')) {
    throw new Error('文件不是有效的纯文本内容')
  }
  if (content.length > MAX_SOURCE_CHARS) {
    throw new Error(`文本内容不能超过 ${MAX_SOURCE_CHARS.toLocaleString()} 个字符`)
  }
  return content
}

function SourceField({
  id,
  title,
  description,
  placeholder,
  value,
  fileName,
  disabled,
  onChange,
  onFileSelect,
  onClearFile,
}: SourceFieldProps) {
  const [isDragging, setIsDragging] = useState(false)

  const handleDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setIsDragging(false)
    if (disabled) return
    const file = event.dataTransfer.files[0]
    if (file) onFileSelect(file)
  }

  return (
    <div className="rounded-2xl border border-gray-200 bg-white/80 p-4 shadow-sm">
      <div className="mb-3 flex items-start justify-between gap-3">
        <div>
          <label htmlFor={`${id}-textarea`} className="text-sm font-semibold text-gray-900">
            {title}
            <span className="ml-1 text-rose-500">*</span>
          </label>
          <p className="mt-0.5 text-xs text-gray-500">{description}</p>
        </div>
        <FileText className="h-5 w-5 flex-shrink-0 text-primary-500" />
      </div>

      <div
        onDragOver={(event) => {
          event.preventDefault()
          if (!disabled) setIsDragging(true)
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        className={`mb-3 flex items-center gap-2 rounded-xl border border-dashed px-3 py-2.5 transition-colors ${
          isDragging
            ? 'border-primary-400 bg-primary-50'
            : 'border-gray-200 bg-gray-50/80 hover:border-primary-200'
        }`}
      >
        <UploadCloud className="h-4 w-4 flex-shrink-0 text-gray-400" />
        {fileName ? (
          <>
            <span className="min-w-0 flex-1 truncate text-xs font-medium text-emerald-700">
              已导入 {fileName}
            </span>
            <button
              type="button"
              onClick={onClearFile}
              disabled={disabled}
              className="rounded-md p-1 text-gray-400 hover:bg-white hover:text-rose-500 disabled:cursor-not-allowed"
              aria-label={`移除${title}文件`}
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </>
        ) : (
          <>
            <span className="flex-1 text-xs text-gray-500">拖入 TXT / MD，或</span>
            <label
              htmlFor={`${id}-file`}
              className={`text-xs font-semibold text-primary-600 ${
                disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer hover:text-primary-700'
              }`}
            >
              选择文件
            </label>
            <input
              id={`${id}-file`}
              type="file"
              accept=".txt,.md,.markdown,text/plain,text/markdown"
              disabled={disabled}
              className="hidden"
              onChange={(event) => {
                const file = event.target.files?.[0]
                event.target.value = ''
                if (file) onFileSelect(file)
              }}
            />
          </>
        )}
      </div>

      <textarea
        id={`${id}-textarea`}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        rows={8}
        maxLength={MAX_SOURCE_CHARS}
        disabled={disabled}
        className="textarea min-h-[180px] resize-y bg-white"
      />
      <div className="mt-2 flex items-center justify-between text-[11px] text-gray-400">
        <span>也可以直接粘贴纯文本</span>
        <span>{value.length.toLocaleString()} / {MAX_SOURCE_CHARS.toLocaleString()}</span>
      </div>
    </div>
  )
}

const connectionMeta: Record<
  InterviewConnectionStatus,
  { label: string; detail: string; tone: string; dot: string }
> = {
  connecting: {
    label: '正在连接',
    detail: '正在建立实时面试通道',
    tone: 'bg-amber-50 text-amber-700 ring-amber-200',
    dot: 'bg-amber-500 animate-pulse',
  },
  connected: {
    label: '服务已连接',
    detail: '实时通道可用；断线后不能恢复原面试',
    tone: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
    dot: 'bg-emerald-500',
  },
  disconnected: {
    label: '连接已断开',
    detail: '正在重连实时通道；已中断的面试无法自动恢复',
    tone: 'bg-gray-50 text-gray-600 ring-gray-200',
    dot: 'bg-gray-400',
  },
  error: {
    label: '连接异常',
    detail: '正在重试连接；进行中的面试无法自动恢复',
    tone: 'bg-red-50 text-red-700 ring-red-200',
    dot: 'bg-red-500',
  },
}

export default function InterviewAgent() {
  const { user } = useAuth()
  const currentUserId = user?.id
  const [connectionStatus, setConnectionStatus] = useState<InterviewConnectionStatus>('connecting')
  const [messages, setMessages] = useState<InterviewMessage[]>([])
  const [currentStage, setCurrentStage] = useState('')
  const [sessionId, setSessionId] = useState('')
  const [isInterviewing, setIsInterviewing] = useState(false)
  const [isStarting, setIsStarting] = useState(false)
  const [isAwaitingResponse, setIsAwaitingResponse] = useState(false)
  const [isStopping, setIsStopping] = useState(false)
  const [hasFinished, setHasFinished] = useState(false)
  const [hasInterrupted, setHasInterrupted] = useState(false)
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(null)
  const [showSetup, setShowSetup] = useState(false)
  const [candidateName, setCandidateName] = useState(user?.name || '')
  const [jdText, setJdText] = useState('')
  const [resumeText, setResumeText] = useState('')
  const [jdFileName, setJdFileName] = useState('')
  const [resumeFileName, setResumeFileName] = useState('')
  const [setupError, setSetupError] = useState('')
  const [input, setInput] = useState('')
  const [sendError, setSendError] = useState('')
  const [questionUpload, setQuestionUpload] = useState<UploadState>({ status: 'idle', message: '' })
  const wsRef = useRef<WebSocket | null>(null)
  const messageIdRef = useRef(0)
  const reconnectTimerRef = useRef<number | null>(null)
  const reconnectAttemptRef = useRef(0)
  const responseTimerRef = useRef<number | null>(null)
  const pendingActionRef = useRef<PendingActionRequest | null>(null)
  const roundActiveRef = useRef(false)
  const activeSessionIdRef = useRef('')
  const interruptedSessionIdsRef = useRef(new Set<string>())
  const messageListRef = useRef<HTMLDivElement>(null)
  const isComposingRef = useRef(false)

  useEffect(() => {
    if (!candidateName && user?.name) setCandidateName(user.name)
  }, [candidateName, user?.name])

  const makeMessage = useCallback(
    (kind: InterviewMessage['kind'], content: string, extra: Partial<InterviewMessage> = {}): InterviewMessage => {
      messageIdRef.current += 1
      return {
        id: `${kind}-${Date.now()}-${messageIdRef.current}`,
        kind,
        content,
        ...extra,
      }
    },
    []
  )

  const appendMessage = useCallback(
    (kind: InterviewMessage['kind'], content: string, extra: Partial<InterviewMessage> = {}) => {
      setMessages((prev) => [...prev, makeMessage(kind, content, extra)])
    },
    [makeMessage]
  )

  const clearPendingAction = useCallback(() => {
    if (responseTimerRef.current !== null) {
      window.clearTimeout(responseTimerRef.current)
      responseTimerRef.current = null
    }
    pendingActionRef.current = null
    setPendingAction(null)
  }, [])

  const updateDeliveryState = useCallback(
    (
      messageId: string | undefined,
      deliveryLabel: string,
      deliveryState: NonNullable<InterviewMessage['deliveryState']>
    ) => {
      if (!messageId) return
      setMessages((prev) =>
        prev.map((message) =>
          message.id === messageId
            ? { ...message, deliveryLabel, deliveryState }
            : message
        )
      )
    },
    []
  )

  const interruptRound = useCallback(
    (reason: string, closeConnection = false) => {
      const pending = pendingActionRef.current
      const interruptedSessionId = activeSessionIdRef.current
      if (interruptedSessionId) {
        interruptedSessionIdsRef.current.add(interruptedSessionId)
      }
      if (pending?.action === 'start') {
        updateDeliveryState(pending.messageId, '未收到服务端启动确认', 'failed')
      } else if (pending?.action === 'answer') {
        updateDeliveryState(pending.messageId, '未收到服务端响应', 'failed')
      }

      clearPendingAction()
      roundActiveRef.current = false
      activeSessionIdRef.current = ''
      setSessionId('')
      setCurrentStage('')
      setIsInterviewing(false)
      setIsStarting(false)
      setIsAwaitingResponse(false)
      setIsStopping(false)
      setHasFinished(false)
      setHasInterrupted(true)
      setInput('')
      setSendError('')
      appendMessage('error', reason)

      if (closeConnection) {
        const socket = wsRef.current
        if (socket && socket.readyState < WebSocket.CLOSING) {
          socket.close(4000, 'interview response timeout')
        }
      }
    },
    [appendMessage, clearPendingAction, updateDeliveryState]
  )

  const armResponseTimeout = useCallback(
    (request: PendingActionRequest) => {
      clearPendingAction()
      pendingActionRef.current = request
      setPendingAction(request.action)

      const timeout =
        request.action === 'start'
          ? START_RESPONSE_TIMEOUT_MS
          : request.action === 'answer'
            ? ANSWER_RESPONSE_TIMEOUT_MS
            : CANCEL_RESPONSE_TIMEOUT_MS

      responseTimerRef.current = window.setTimeout(() => {
        if (pendingActionRef.current !== request) return

        const reason =
          request.action === 'start'
            ? '启动请求已发送，但 3 分钟内未收到服务端响应，本轮未确认开始。请重新开始。'
            : request.action === 'answer'
              ? '回答已发送，但 3 分钟内未收到服务端响应。本轮已中断，请重新开始。'
              : '提前结束请求已发送，但 2 分钟内未收到服务端响应。本轮状态无法确认，请重新开始。'
        interruptRound(reason, true)
      }, timeout)
    },
    [clearPendingAction, interruptRound]
  )

  const handleServerEvent = useCallback(
    (event: InterviewServerEvent) => {
      if (!roundActiveRef.current) return

      if (event.sessionId && interruptedSessionIdsRef.current.has(event.sessionId)) {
        return
      }
      const activeSessionId = activeSessionIdRef.current
      if (event.sessionId && activeSessionId && event.sessionId !== activeSessionId) {
        return
      }
      if (event.sessionId && !activeSessionId) {
        activeSessionIdRef.current = event.sessionId
        setSessionId(event.sessionId)
      }

      const pending = pendingActionRef.current
      const confirmsPendingAction =
        pending &&
        ((pending.action === 'start' && Boolean(event.sessionId)) ||
          (pending.action !== 'start' &&
            Boolean(event.sessionId) &&
            event.sessionId === pending.sessionId))

      if (confirmsPendingAction) {
        if (pending.action === 'start') {
          updateDeliveryState(pending.messageId, '服务端已响应', 'confirmed')
        } else if (pending.action === 'answer') {
          updateDeliveryState(pending.messageId, '服务端已响应', 'confirmed')
        }
        clearPendingAction()
      }

      setHasInterrupted(false)

      switch (event.type) {
        case 'stage_change': {
          if (event.stage) setCurrentStage(event.stage)
          if (event.message) appendMessage('stage', event.message)

          if (event.stage === 'completed') {
            clearPendingAction()
            roundActiveRef.current = false
            activeSessionIdRef.current = ''
            setIsInterviewing(false)
            setIsStarting(false)
            setIsAwaitingResponse(false)
            setIsStopping(false)
            setHasFinished(true)
          }
          break
        }
        case 'question':
          if (event.questionNum <= 0) {
            setCurrentStage('review_weak')
            setIsStarting(false)
            setIsAwaitingResponse(true)
            appendMessage('question', event.content || '请复习当前知识点。', {
              questionNum: event.questionNum,
            })
            break
          }
          setCurrentStage('interview')
          setIsInterviewing(true)
          setIsStarting(false)
          setIsAwaitingResponse(false)
          setIsStopping(false)
          appendMessage('question', event.content || '请回答当前问题。', {
            questionNum: event.questionNum,
          })
          break
        case 'score':
          setIsStarting(false)
          setIsAwaitingResponse(true)
          appendMessage('score', event.feedback, {
            score: event.score,
            keyPointsHit: event.keyPointsHit,
            keyPointsMissed: event.keyPointsMissed,
          })
          break
        case 'report':
          appendMessage('report', event.content)
          break
        case 'review_plan':
          appendMessage('review_plan', event.content)
          break
      }
    },
    [appendMessage, clearPendingAction, updateDeliveryState]
  )

  useEffect(() => {
    if (!currentUserId) {
      setConnectionStatus('disconnected')
      if (roundActiveRef.current) {
        interruptRound('登录状态已失效，本轮面试已中断。重新登录后请重新开始。')
      }
      return
    }

    let disposed = false

    const connect = () => {
      if (disposed) return
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }

      getAuthToken()
      setConnectionStatus('connecting')

      let socket: WebSocket
      try {
        socket = new WebSocket(buildInterviewWsUrl())
      } catch {
        setConnectionStatus('error')
        return
      }

      wsRef.current = socket

      socket.onopen = () => {
        if (disposed) return
        reconnectAttemptRef.current = 0
        setConnectionStatus('connected')
      }

      socket.onmessage = (messageEvent) => {
        let payload: unknown
        try {
          payload = JSON.parse(messageEvent.data)
        } catch {
          return
        }
        const event = parseInterviewServerEvent(payload)
        if (event) handleServerEvent(event)
      }

      socket.onerror = () => {
        if (!disposed) setConnectionStatus('error')
      }

      socket.onclose = () => {
        if (wsRef.current === socket) wsRef.current = null
        if (disposed) return

        if (roundActiveRef.current) {
          interruptRound(
            '实时连接已断开。当前后端没有面试恢复接口，本轮已中断；连接恢复后请重新开始。'
          )
        }
        setConnectionStatus('disconnected')
        const attempt = reconnectAttemptRef.current
        reconnectAttemptRef.current += 1
        const delay = Math.min(1_500 * 2 ** attempt, 10_000)
        reconnectTimerRef.current = window.setTimeout(connect, delay)
      }
    }

    connect()

    return () => {
      disposed = true
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }
      if (responseTimerRef.current !== null) {
        window.clearTimeout(responseTimerRef.current)
        responseTimerRef.current = null
      }
      pendingActionRef.current = null
      const socket = wsRef.current
      wsRef.current = null
      socket?.close()
    }
  }, [currentUserId, handleServerEvent, interruptRound])

  useEffect(() => {
    const container = messageListRef.current
    if (!container) return
    container.scrollTo({ top: container.scrollHeight, behavior: 'smooth' })
  }, [isAwaitingResponse, isStarting, messages])

  const sendEvent = useCallback((event: InterviewClientEvent) => {
    const socket = wsRef.current
    if (!socket || socket.readyState !== WebSocket.OPEN) return false
    try {
      socket.send(JSON.stringify(buildInterviewEnvelope(event)))
      return true
    } catch {
      return false
    }
  }, [])

  const handleSourceFile = async (target: 'jd' | 'resume', file: File) => {
    setSetupError('')
    try {
      const content = await readSourceFile(file)
      if (target === 'jd') {
        setJdText(content)
        setJdFileName(file.name)
      } else {
        setResumeText(content)
        setResumeFileName(file.name)
      }
    } catch (error) {
      setSetupError(error instanceof Error ? error.message : '读取文本文件失败')
    }
  }

  const openSetup = () => {
    setSetupError('')
    setSendError('')
    setShowSetup(true)
  }

  const handleStartInterview = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSetupError('')

    const jd = jdText.trim()
    const resume = resumeText.trim()
    if (!jd || !resume) {
      setSetupError('请先填写岗位描述和简历内容')
      return
    }
    if (connectionStatus !== 'connected') {
      setSetupError('实时服务尚未连接，请稍后再试')
      return
    }

    const sent = sendEvent({
      type: 'start_interview',
      jd,
      resume,
      candidate_name: candidateName.trim() || user?.name || '候选人',
    })
    if (!sent) {
      setSetupError('实时连接暂不可用，请稍后再试')
      return
    }

    const startMessage = makeMessage('answer', '岗位描述与简历已发送。', {
      deliveryLabel: '已发送 · 等待服务端启动',
      deliveryState: 'pending',
    })
    setMessages([
      startMessage,
      makeMessage('system', '正在等待服务端创建面试会话；收到面试事件后才视为启动成功。'),
    ])
    roundActiveRef.current = true
    activeSessionIdRef.current = ''
    setSessionId('')
    setCurrentStage('')
    setIsInterviewing(true)
    setIsStarting(true)
    setIsAwaitingResponse(false)
    setIsStopping(false)
    setHasFinished(false)
    setHasInterrupted(false)
    setShowSetup(false)
    setInput('')
    setSendError('')
    armResponseTimeout({
      action: 'start',
      sessionId: '',
      messageId: startMessage.id,
    })
  }

  const handleAnswer = () => {
    const answer = input.trim()
    setSendError('')
    if (!answer) return
    if (!sessionId) {
      setSendError('面试会话仍在准备中，请稍后再试')
      return
    }
    if (connectionStatus !== 'connected') {
      setSendError('实时连接尚未恢复，暂时无法提交回答')
      return
    }
    if (!sendEvent({ type: 'answer', session_id: sessionId, answer })) {
      setSendError('回答发送失败，请检查连接后重试')
      return
    }

    const answerMessage = makeMessage('answer', answer, {
      deliveryLabel: '已发送 · 等待服务端响应',
      deliveryState: 'pending',
    })
    setMessages((prev) => [...prev, answerMessage])
    setInput('')
    setIsAwaitingResponse(true)
    armResponseTimeout({
      action: 'answer',
      sessionId,
      messageId: answerMessage.id,
    })
  }

  const handleAnswerSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    handleAnswer()
  }

  const handleStopInterview = () => {
    if (!sessionId || isStarting || isAwaitingResponse || isStopping) return
    if (!window.confirm('确定提前结束本轮面试吗？系统会根据已完成的回答生成报告。')) return
    if (!sendEvent({ type: 'cancel_interview', session_id: sessionId })) {
      setSendError('终止指令发送失败，请检查连接后重试')
      return
    }
    appendMessage('system', '提前结束请求已发送，等待服务端响应。')
    setIsStopping(true)
    setIsAwaitingResponse(false)
    setSendError('')
    armResponseTimeout({ action: 'cancel', sessionId })
  }

  const handleQuestionUpload = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) return

    const extension = getFileExtension(file.name)
    if (!TEXT_FILE_EXTENSIONS.has(extension)) {
      setQuestionUpload({ status: 'error', message: '题库仅支持 TXT 或 Markdown 文本文件' })
      return
    }
    if (file.size > MAX_QUESTION_FILE_SIZE) {
      setQuestionUpload({ status: 'error', message: '题库文件不能超过 5MB' })
      return
    }

    setQuestionUpload({ status: 'uploading', message: `正在上传并解析 ${file.name}` })
    try {
      await uploadQuestionBank(file)
      setQuestionUpload({
        status: 'success',
        message: `${file.name} 已上传，后端已接收并处理`,
      })
    } catch (error) {
      setQuestionUpload({
        status: 'error',
        message: error instanceof Error ? error.message : '题库上传失败',
      })
    }
  }

  const connection = connectionMeta[connectionStatus]
  const inputDisabled =
    !isInterviewing ||
    isStarting ||
    isAwaitingResponse ||
    isStopping ||
    !sessionId ||
    connectionStatus !== 'connected'
  const canSubmitAnswer = Boolean(input.trim()) && !inputDisabled
  const activityLabel = isStarting
    ? pendingAction === 'start'
      ? '启动请求已发送，等待服务端响应…'
      : '服务端已响应，正在分析 JD、简历并规划问题…'
    : isStopping
      ? pendingAction === 'cancel'
        ? '提前结束请求已发送，等待服务端响应…'
        : '服务端已响应，正在生成提前结束后的评估报告…'
      : pendingAction === 'answer'
        ? '回答已发送，等待服务端评分或下一步响应…'
        : currentStage.startsWith('review_weak')
          ? '正在整理低分题目的参考答案…'
          : currentStage.startsWith('evaluation')
            ? '正在生成本轮面试评估报告…'
            : currentStage.startsWith('review_plan')
              ? '正在生成个性化复习计划…'
        : isAwaitingResponse
          ? '已收到服务端响应，正在准备下一步…'
          : ''

  const interviewStatusLabel = useMemo(() => {
    if (hasFinished) return '本轮已完成'
    if (hasInterrupted) return '本轮已中断'
    if (pendingAction === 'start') return '等待服务端启动'
    if (pendingAction === 'cancel') return '结束请求已发送'
    if (isStopping) return '正在收尾'
    if (isStarting) return '正在准备'
    if (isInterviewing) return '面试进行中'
    return '等待开始'
  }, [hasFinished, hasInterrupted, isInterviewing, isStarting, isStopping, pendingAction])

  return (
    <div className="mx-auto max-w-7xl space-y-5">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <Link
            to="/"
            className="mb-2 inline-flex items-center gap-2 text-sm text-gray-500 transition-colors hover:text-primary-600"
          >
            <ArrowLeft className="h-4 w-4" />
            返回首页
          </Link>
          <div className="flex items-center gap-3">
            <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-600 to-primary-500 text-white shadow-md shadow-primary-200/60">
              <GraduationCap className="h-6 w-6" />
            </span>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">面试 Agent</h1>
              <p className="text-sm text-gray-500">基于岗位与简历，完成一场有反馈的模拟面试</p>
            </div>
          </div>
        </div>

        <div className={`inline-flex items-center gap-2 self-start rounded-full px-3 py-1.5 text-xs font-medium ring-1 ${connection.tone}`}>
          <span className={`h-2 w-2 rounded-full ${connection.dot}`} />
          {connection.label}
        </div>
      </div>

      <div className="grid gap-5 lg:grid-cols-[270px_minmax(0,1fr)]">
        <aside className="space-y-4">
          <div className="card p-4">
            <div className="mb-4 flex items-center gap-3">
              <UserAvatar
                avatar={user?.avatar}
                name={user?.name || '候选人'}
                userId={user?.id}
                className="h-11 w-11 rounded-xl"
              />
              <div className="min-w-0">
                <p className="truncate text-sm font-semibold text-gray-900">{user?.name || '候选人'}</p>
                <p className="text-xs text-gray-500">{interviewStatusLabel}</p>
              </div>
            </div>

            <button
              type="button"
              onClick={openSetup}
              disabled={isInterviewing}
              className="btn-primary w-full"
            >
              {hasFinished || hasInterrupted ? (
                <RotateCcw className="h-4 w-4" />
              ) : (
                <Play className="h-4 w-4" />
              )}
              {isInterviewing
                ? pendingAction === 'start'
                  ? '等待服务端启动'
                  : '面试进行中'
                : hasFinished
                  ? '开启下一场'
                  : hasInterrupted
                    ? '重新开始'
                    : '开始模拟面试'}
            </button>

            <div className={`mt-3 rounded-xl px-3 py-2.5 text-xs ring-1 ${connection.tone}`}>
              <div className="flex items-center gap-2 font-semibold">
                {connectionStatus === 'connected' ? <Wifi className="h-4 w-4" /> : <WifiOff className="h-4 w-4" />}
                {connection.label}
              </div>
              <p className="mt-1 opacity-75">{connection.detail}</p>
            </div>
          </div>

          <div className="card p-4">
            <div className="mb-3 flex items-center gap-2">
              <BookOpenCheck className="h-5 w-5 text-primary-600" />
              <h2 className="text-sm font-semibold text-gray-900">自定义题库</h2>
            </div>
            <p className="mb-3 text-xs leading-relaxed text-gray-500">
              上传 UTF-8 文本题库，后端会接收并尝试纳入后续出题检索。
            </p>
            <label
              htmlFor="question-bank-file"
              className={`btn-secondary w-full cursor-pointer !px-3 !py-2 text-xs ${
                isInterviewing || questionUpload.status === 'uploading'
                  ? 'pointer-events-none opacity-50'
                  : ''
              }`}
            >
              {questionUpload.status === 'uploading' ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <FileUp className="h-4 w-4" />
              )}
              {questionUpload.status === 'uploading' ? '上传解析中' : '上传 TXT / MD'}
            </label>
            <input
              id="question-bank-file"
              type="file"
              accept=".txt,.md,.markdown,text/plain,text/markdown"
              className="hidden"
              disabled={isInterviewing || questionUpload.status === 'uploading'}
              onChange={handleQuestionUpload}
            />
            {questionUpload.message && (
              <div
                className={`mt-3 flex items-start gap-2 rounded-xl px-3 py-2 text-xs ${
                  questionUpload.status === 'success'
                    ? 'bg-emerald-50 text-emerald-700'
                    : questionUpload.status === 'error'
                      ? 'bg-red-50 text-red-700'
                      : 'bg-primary-50 text-primary-700'
                }`}
                role="status"
              >
                {questionUpload.status === 'success' ? (
                  <CheckCircle2 className="mt-0.5 h-3.5 w-3.5 flex-shrink-0" />
                ) : questionUpload.status === 'error' ? (
                  <AlertCircle className="mt-0.5 h-3.5 w-3.5 flex-shrink-0" />
                ) : (
                  <Loader2 className="mt-0.5 h-3.5 w-3.5 flex-shrink-0 animate-spin" />
                )}
                <span>{questionUpload.message}</span>
              </div>
            )}
          </div>

          <div className="card p-4">
            <div className="mb-3 flex items-center gap-2">
              <Target className="h-5 w-5 text-primary-600" />
              <h2 className="text-sm font-semibold text-gray-900">面试建议</h2>
            </div>
            <ol className="space-y-3 text-xs text-gray-600">
              {[
                '提供完整 JD，出题方向会更贴近目标岗位。',
                '简历保留项目成果和技术细节，便于生成针对性追问。',
                '回答尽量说明背景、行动与结果，评分会更准确。',
              ].map((tip, index) => (
                <li key={tip} className="flex items-start gap-2.5">
                  <span className="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary-50 text-[10px] font-bold text-primary-700 ring-1 ring-primary-100">
                    {index + 1}
                  </span>
                  <span className="leading-relaxed">{tip}</span>
                </li>
              ))}
            </ol>
          </div>
        </aside>

        <section className="min-w-0 overflow-hidden rounded-2xl bg-white/80 shadow-sm ring-1 ring-gray-200/60 backdrop-blur-md">
          <div className="flex items-center justify-between border-b border-gray-100 px-4 py-3.5 sm:px-6">
            <div className="flex items-center gap-3">
              <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary-50 text-primary-700 ring-1 ring-primary-100">
                <Bot className="h-5 w-5" />
              </span>
              <div>
                <h2 className="text-sm font-semibold text-gray-900">模拟面试室</h2>
                <p className="text-xs text-gray-500">
                  {isInterviewing
                    ? pendingAction === 'start'
                      ? '启动请求已发送，等待服务端响应'
                      : '请按问题逐题作答'
                    : hasFinished
                      ? '报告与复习计划已生成'
                      : hasInterrupted
                        ? '本轮已中断，不能从断点恢复'
                        : '准备好后开始一场新面试'}
                </p>
              </div>
            </div>
            {isInterviewing && sessionId && (
              <button
                type="button"
                onClick={handleStopInterview}
                disabled={
                  isStarting ||
                  isAwaitingResponse ||
                  isStopping ||
                  connectionStatus !== 'connected'
                }
                className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-xs font-semibold text-rose-600 transition-colors hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {isStopping ? <Loader2 className="h-4 w-4 animate-spin" /> : <StopCircle className="h-4 w-4" />}
                <span className="hidden sm:inline">提前结束</span>
              </button>
            )}
          </div>

          {(currentStage || (isInterviewing && pendingAction !== 'start')) && (
            <InterviewStageBar currentStage={currentStage} />
          )}

          {showSetup ? (
            <form onSubmit={handleStartInterview} className="p-4 sm:p-6">
              <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <BrainCircuit className="h-5 w-5 text-primary-600" />
                    <h2 className="text-lg font-semibold text-gray-900">准备面试材料</h2>
                  </div>
                  <p className="mt-1 text-sm text-gray-500">内容只用于本轮面试分析和个性化出题。</p>
                </div>
                <button
                  type="button"
                  onClick={() => setShowSetup(false)}
                  className="self-start rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                  aria-label="关闭面试准备"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>

              <div className="mb-4">
                <label htmlFor="candidate-name" className="mb-2 flex items-center gap-2 text-sm font-medium text-gray-700">
                  <User className="h-4 w-4 text-gray-400" />
                  候选人称呼
                </label>
                <input
                  id="candidate-name"
                  type="text"
                  value={candidateName}
                  onChange={(event) => setCandidateName(event.target.value)}
                  maxLength={40}
                  placeholder="用于面试报告中的称呼"
                  className="input max-w-md"
                />
              </div>

              <div className="grid gap-4 xl:grid-cols-2">
                <SourceField
                  id="jd-source"
                  title="岗位描述（JD）"
                  description="职责、技术栈、经验要求越完整越好"
                  placeholder="粘贴目标岗位的职责、任职要求和技术栈…"
                  value={jdText}
                  fileName={jdFileName}
                  disabled={isStarting}
                  onChange={(value) => {
                    setJdText(value)
                    setSetupError('')
                  }}
                  onFileSelect={(file) => void handleSourceFile('jd', file)}
                  onClearFile={() => {
                    setJdFileName('')
                    setJdText('')
                  }}
                />
                <SourceField
                  id="resume-source"
                  title="个人简历"
                  description="突出项目经历、技术难点和量化成果"
                  placeholder="粘贴个人经历、项目说明与技术能力…"
                  value={resumeText}
                  fileName={resumeFileName}
                  disabled={isStarting}
                  onChange={(value) => {
                    setResumeText(value)
                    setSetupError('')
                  }}
                  onFileSelect={(file) => void handleSourceFile('resume', file)}
                  onClearFile={() => {
                    setResumeFileName('')
                    setResumeText('')
                  }}
                />
              </div>

              {setupError && (
                <div className="mt-4 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700" role="alert">
                  <AlertCircle className="mt-0.5 h-4 w-4 flex-shrink-0" />
                  {setupError}
                </div>
              )}

              <div className="mt-5 flex flex-col-reverse gap-3 border-t border-gray-100 pt-5 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-xs text-gray-500">首次分析通常需要 1–3 分钟，请保持页面打开。</p>
                <div className="flex items-center gap-2">
                  <button type="button" onClick={() => setShowSetup(false)} className="btn-secondary">
                    取消
                  </button>
                  <button
                    type="submit"
                    disabled={!jdText.trim() || !resumeText.trim() || connectionStatus !== 'connected'}
                    className="btn-primary"
                  >
                    <Sparkles className="h-4 w-4" />
                    开始面试
                  </button>
                </div>
              </div>
            </form>
          ) : (
            <>
              <div
                ref={messageListRef}
                className="h-[54vh] min-h-[440px] overflow-y-auto px-4 py-5 sm:px-6"
                aria-live="polite"
              >
                {messages.length === 0 ? (
                  <div className="flex h-full flex-col items-center justify-center px-4 text-center">
                    <div className="relative mb-5">
                      <div className="absolute inset-0 rounded-full bg-primary-200/50 blur-2xl" />
                      <span className="relative flex h-20 w-20 items-center justify-center rounded-3xl bg-gradient-to-br from-primary-600 to-primary-500 text-white shadow-xl shadow-primary-200/70">
                        <GraduationCap className="h-10 w-10" />
                      </span>
                    </div>
                    <h2 className="text-xl font-bold text-gray-900">把每次练习，都变成可复盘的进步</h2>
                    <p className="mt-2 max-w-lg text-sm leading-relaxed text-gray-500">
                      面试 Agent 会分析岗位与简历，动态追问、逐题评分，并在结束后生成评估报告和复习计划。
                    </p>
                    <div className="mt-5 flex flex-wrap justify-center gap-2 text-xs text-gray-600">
                      {['岗位定向出题', '实时追问与评分', '报告与复习计划'].map((feature) => (
                        <span key={feature} className="rounded-full bg-gray-50 px-3 py-1.5 ring-1 ring-gray-200">
                          {feature}
                        </span>
                      ))}
                    </div>
                    <button type="button" onClick={openSetup} className="btn-primary mt-6">
                      <Play className="h-4 w-4" />
                      开始模拟面试
                    </button>
                  </div>
                ) : hasInterrupted ? (
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex items-start gap-2 text-sm text-rose-700">
                      <AlertCircle className="mt-0.5 h-5 w-5 flex-shrink-0" />
                      <span>本轮面试已中断，当前后端不支持断点恢复。请重新提交材料开始新面试。</span>
                    </div>
                    <button
                      type="button"
                      onClick={openSetup}
                      disabled={connectionStatus !== 'connected'}
                      className="btn-primary"
                    >
                      <RotateCcw className="h-4 w-4" />
                      重新开始
                    </button>
                  </div>
                ) : (
                  messages.map((message) => (
                    <InterviewMessageItem key={message.id} message={message} user={user} />
                  ))
                )}

                {activityLabel && (
                  <div className="my-4 flex items-center gap-3 text-sm text-gray-500">
                    <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary-50 text-primary-600 ring-1 ring-primary-100">
                      <Loader2 className="h-4 w-4 animate-spin" />
                    </span>
                    <span>{activityLabel}</span>
                  </div>
                )}
              </div>

              <div className="border-t border-gray-100 bg-white/70 p-4 sm:px-6">
                {isInterviewing ? (
                  <form onSubmit={handleAnswerSubmit}>
                    {sendError && (
                      <div className="mb-3 flex items-center gap-2 text-xs text-red-600" role="alert">
                        <AlertCircle className="h-3.5 w-3.5" />
                        {sendError}
                      </div>
                    )}
                    <div className="flex items-end gap-2.5">
                      <textarea
                        value={input}
                        onChange={(event) => setInput(event.target.value)}
                        onKeyDown={(event) => {
                          if (event.key === 'Enter' && !event.shiftKey && !isComposingRef.current) {
                            event.preventDefault()
                            handleAnswer()
                          }
                        }}
                        onCompositionStart={() => {
                          isComposingRef.current = true
                        }}
                        onCompositionEnd={() => {
                          isComposingRef.current = false
                        }}
                        placeholder={
                          isStarting
                            ? '正在准备面试问题…'
                            : isStopping
                              ? '正在整理评估报告…'
                              : isAwaitingResponse
                                ? '正在评估上一条回答…'
                                : !sessionId
                                  ? '正在建立面试会话…'
                                  : '输入你的回答，Enter 发送，Shift + Enter 换行'
                        }
                        rows={2}
                        disabled={inputDisabled}
                        className="textarea min-h-[52px] flex-1 resize-none bg-white"
                      />
                      <button
                        type="submit"
                        disabled={!canSubmitAnswer}
                        className="btn-primary h-[52px] !px-4"
                        aria-label="发送回答"
                      >
                        <Send className="h-4 w-4" />
                        <span className="hidden sm:inline">发送</span>
                      </button>
                    </div>
                  </form>
                ) : hasFinished ? (
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex items-center gap-2 text-sm text-emerald-700">
                      <CheckCircle2 className="h-5 w-5" />
                      本轮面试已完成，展开上方卡片查看结果。
                    </div>
                    <button type="button" onClick={openSetup} className="btn-primary">
                      <RotateCcw className="h-4 w-4" />
                      再练一场
                    </button>
                  </div>
                ) : (
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <p className="text-sm text-gray-500">提交 JD 与简历后，Agent 会自动开始分析和出题。</p>
                    <button type="button" onClick={openSetup} className="btn-primary">
                      <Play className="h-4 w-4" />
                      开始模拟面试
                    </button>
                  </div>
                )}
              </div>
            </>
          )}
        </section>
      </div>
    </div>
  )
}
