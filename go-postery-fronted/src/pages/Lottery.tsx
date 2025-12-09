import { useEffect, useMemo, useRef, useState } from 'react'
import { CheckCircle2, Clock3, RotateCcw, Sparkles, Ticket, Wallet, XCircle } from 'lucide-react'

type Prize = {
  label: string
  desc: string
  color: string
  type?: 'lose' | 'retry'
  payAmount?: number
}

type HistoryItem = {
  id: number
  prize: string
  status: string
  time: string
}

const DECISION_SECONDS = 15

const prizes: Prize[] = [
  { label: '会员月卡', desc: '30 天尊享', color: '#0ea5e9', payAmount: 12 },
  { label: '双倍积分', desc: '签到翻倍', color: '#14b8a6', payAmount: 3 },
  { label: '定制周边', desc: '帆布袋+贴纸', color: '#22c55e', payAmount: 39 },
  { label: '谢谢参与', desc: '下次再来', color: '#e5e7eb', type: 'lose' },
  { label: '再来一次', desc: '免费重抽', color: '#a855f7', type: 'retry' },
  { label: '加速券', desc: '中奖翻倍', color: '#6366f1', payAmount: 5 },
  { label: '线下沙龙票', desc: '城市共创会', color: '#f59e0b', payAmount: 29 },
  { label: '周卡升级券', desc: '特权直升', color: '#f97316', payAmount: 9 },
]

export default function Lottery() {
  const angle = 360 / prizes.length
  const [rotation, setRotation] = useState(0)
  const [isSpinning, setIsSpinning] = useState(false)
  const [result, setResult] = useState<Prize | null>(null)
  const [decisionState, setDecisionState] = useState<'idle' | 'pending' | 'paid' | 'abandoned' | 'missed'>('idle')
  const [countdown, setCountdown] = useState(0)
  const [history, setHistory] = useState<HistoryItem[]>([])
  const activeRecordId = useRef<number | null>(null)
  const decisionDeadlineRef = useRef<number | null>(null)
  const spinTimerRef = useRef<number | null>(null)
  const countdownTimerRef = useRef<number | null>(null)

  const wheelGradient = useMemo(() => {
    return prizes
      .map((item, index) => {
        const start = index * angle
        const end = start + angle
        return `${item.color} ${start}deg ${end}deg`
      })
      .join(', ')
  }, [angle])

  const resetDecisionTimer = () => {
    if (countdownTimerRef.current) {
      clearInterval(countdownTimerRef.current)
    }
    countdownTimerRef.current = null
    decisionDeadlineRef.current = null
    setCountdown(0)
  }

  const updateHistoryStatus = (status: string) => {
    if (!activeRecordId.current) return
    setHistory(prev =>
      prev.map(item => (item.id === activeRecordId.current ? { ...item, status } : item))
    )
  }

  const startDecisionWindow = (prize: Prize) => {
    if (prize.type === 'lose') {
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
        resetDecisionTimer()
        setDecisionState('abandoned')
        updateHistoryStatus('超时自动放弃')
      }
    }, 250)
  }

  const spin = () => {
    if (isSpinning) return

    setIsSpinning(true)
    setDecisionState('idle')
    setResult(null)
    resetDecisionTimer()

    const targetIndex = Math.floor(Math.random() * prizes.length)
    const extraTurns = 6 + Math.floor(Math.random() * 3)
    const targetRotation =
      extraTurns * 360 + (360 - targetIndex * angle - angle / 2)

    setRotation(prev => prev + targetRotation)

    spinTimerRef.current = window.setTimeout(() => {
      const prize = prizes[targetIndex]
      setResult(prize)
      setIsSpinning(false)
      const recordId = Date.now()
      activeRecordId.current = recordId
      setHistory(prev => [
        {
          id: recordId,
          prize: prize.label,
          status: prize.type === 'lose' ? '未中奖' : '等待确认',
          time: new Date().toLocaleTimeString(),
        },
        ...prev,
      ].slice(0, 8))

      startDecisionWindow(prize)
    }, 4400)
  }

  const handlePay = () => {
    if (decisionState !== 'pending') return
    setDecisionState('paid')
    resetDecisionTimer()
    updateHistoryStatus('已支付领取')
  }

  const handleAbandon = () => {
    if (decisionState !== 'pending') return
    setDecisionState('abandoned')
    resetDecisionTimer()
    updateHistoryStatus('已主动放弃')
  }

  const handleReset = () => {
    resetDecisionTimer()
    setDecisionState('idle')
    setResult(null)
    activeRecordId.current = null
  }

  useEffect(() => {
    return () => {
      if (spinTimerRef.current) clearTimeout(spinTimerRef.current)
      resetDecisionTimer()
    }
  }, [])

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
              旋转轮盘，抽取会员月卡、沙龙门票、定制周边等权益。中奖后需在 {DECISION_SECONDS} 秒内决定支付或放弃，超时将自动作废。支付仅为模拟，不会真实扣款，用于确认你已锁定奖品。
            </p>
            <div className="flex flex-wrap gap-3 text-sm text-gray-600">
              <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white border border-gray-200">
                <Clock3 className="h-4 w-4 text-primary-600" />
                {DECISION_SECONDS} 秒决策期
              </span>
              <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white border border-gray-200">
                <Wallet className="h-4 w-4 text-primary-600" />
                中后付款可立即锁定
              </span>
              <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-white border border-gray-200">
                <Ticket className="h-4 w-4 text-primary-600" />
                次数不限，随时畅玩
              </span>
            </div>
          </div>
          <div className="flex gap-2 items-center">
            <button
              type="button"
              onClick={spin}
              disabled={isSpinning}
              className="btn-primary flex items-center gap-2 shadow-md disabled:cursor-not-allowed"
            >
              <Sparkles className="h-5 w-5" />
              启动轮盘
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
        </div>
      </div>

      <div className="grid lg:grid-cols-[1.1fr_0.9fr] gap-6">
        <div className="card relative overflow-hidden">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(14,165,233,0.08),transparent_45%),radial-gradient(circle_at_80%_10%,rgba(99,102,241,0.08),transparent_40%)]" />
          <div className="relative flex flex-col items-center gap-6">
            <div className="text-center space-y-1">
              <p className="text-gray-500 text-sm">幸运值越高越容易拿奖</p>
              <p className="text-lg font-semibold text-gray-900">
                {isSpinning ? '正在加速...' : result ? result.label : '点击上方按钮开始抽奖'}
              </p>
            </div>

            <div className="relative w-full max-w-[480px] aspect-square">
              <div
                className={`absolute inset-0 rounded-full shadow-lg border-8 border-white transition-transform duration-[4200ms] ease-out`}
                style={{
                  background: `conic-gradient(${wheelGradient})`,
                  transform: `rotate(${rotation}deg)`,
                }}
              >
                {prizes.map((prize, index) => {
                  const prizeRotate = angle * index + angle / 2
                  return (
                    <div
                      key={prize.label}
                      className="absolute inset-0 flex items-start"
                      style={{ transform: `rotate(${prizeRotate}deg)` }}
                    >
                      <div className="origin-center translate-y-2 flex flex-col items-center text-center w-1/2 ml-auto mr-auto"
                        style={{ transform: `rotate(${-prizeRotate}deg)` }}
                      >
                        <span className="text-xs font-semibold text-gray-900 drop-shadow-sm">
                          {prize.label}
                        </span>
                        <span className="text-[11px] text-gray-700 opacity-80">{prize.desc}</span>
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
                <div className="text-xs text-gray-500">手气值</div>
                <div className="text-3xl font-bold text-gray-900 flex items-baseline gap-1">
                  {(100 - countdown).toLocaleString()}
                  <span className="text-sm text-gray-500">pt</span>
                </div>
                <p className="text-[13px] text-gray-500 mt-1">保持专注，幸运会更高</p>
              </div>
            </div>

            <div className="grid sm:grid-cols gap-3 w-full">
              <div className="rounded-xl bg-primary-50 border border-primary-100 px-4 py-3 text-sm text-primary-800">
                中奖需在 {DECISION_SECONDS} 秒内选择支付或放弃，超时自动视为放弃
              </div>
          
            </div>
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
                  剩余 {countdown}s
                </div>
              )}
            </div>

            {!result && (
              <div className="text-sm text-gray-600">
                先启动轮盘，抽中后会在这里展示奖品并进入支付/放弃倒计时。
              </div>
            )}

            {result && (
              <div className="space-y-3">
                <div className="flex items-start gap-3 p-3 rounded-lg bg-gray-50 border border-gray-200">
                  <div className="w-10 h-10 rounded-full flex items-center justify-center text-white font-semibold" style={{ backgroundColor: result.color }}>
                    {result.label.slice(0, 2)}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <p className="font-semibold text-gray-900">{result.label}</p>
                      {result.type === 'lose' && (
                        <span className="text-xs text-gray-500 bg-white border border-gray-200 rounded-full px-2 py-0.5">未中奖</span>
                      )}
                      {result.type === 'retry' && (
                        <span className="text-xs text-primary-700 bg-primary-50 border border-primary-100 rounded-full px-2 py-0.5">再抽一次</span>
                      )}
                    </div>
                    <p className="text-sm text-gray-600">{result.desc}</p>
                    {result.type !== 'lose' && (
                      <p className="text-sm text-gray-700 mt-1">
                        需支付 <span className="font-semibold text-primary-700">¥{result.payAmount ?? 0}</span> 锁定资格
                      </p>
                    )}
                  </div>
                </div>

                {result.type === 'lose' && (
                  <div className="flex items-center gap-2 text-sm text-gray-600">
                    <XCircle className="h-4 w-4 text-gray-400" />
                    没关系，休息片刻再来试试。
                  </div>
                )}

                {result.type !== 'lose' && (
                  <div className="flex flex-col gap-3">
                    <div className="flex items-center gap-2 text-sm">
                      <Clock3 className="h-4 w-4 text-primary-600" />
                      <span className="text-gray-700">
                        请在倒计时结束前确认，超时视为放弃。
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <button
                        type="button"
                        onClick={handlePay}
                        disabled={decisionState !== 'pending'}
                        className="btn-primary flex-1 flex items-center justify-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
                      >
                        <Wallet className="h-4 w-4" />
                        支付并领取
                      </button>
                      <button
                        type="button"
                        onClick={handleAbandon}
                        disabled={decisionState !== 'pending'}
                        className="btn-secondary flex-1 flex items-center justify-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
                      >
                        <XCircle className="h-4 w-4" />
                        放弃
                      </button>
                    </div>
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
                    {decisionState === 'missed' && (
                      <div className="flex items-center gap-2 text-sm text-gray-600 bg-gray-50 border border-gray-200 px-3 py-2 rounded-lg">
                        <XCircle className="h-4 w-4" />
                        未中奖，本次不进入支付流程。
                      </div>
                    )}
                  </div>
                )}
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
              <p className="text-sm text-gray-600">暂无记录，开始旋转轮盘吧。</p>
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
              <li>本活动不限次数，随时可抽，抽中「再来一次」也不会限制次数。</li>
              <li>抽中后需在 {DECISION_SECONDS} 秒内选择「支付并领取」或「放弃」。</li>
              <li>超时未选择视为放弃，奖品会自动释放。</li>
              <li>所有支付均为模拟，用于展示支付/放弃交互流程。</li>
              <li>如遇异常可刷新页面重试，记录仅保留本地最近 8 条。</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
