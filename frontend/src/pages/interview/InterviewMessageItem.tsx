import { useState, type AnchorHTMLAttributes, type HTMLAttributes } from 'react'
import {
  AlertCircle,
  Bot,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  ClipboardList,
  Sparkles,
  Target,
} from 'lucide-react'
import ReactMarkdown, { type Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import UserAvatar from '../../components/UserAvatar'
import type { InterviewMessage } from './types'

const markdownComponents: Components = {
  h1: (props: HTMLAttributes<HTMLHeadingElement>) => (
    <h1 {...props} className="mb-2 mt-4 text-lg font-bold first:mt-0" />
  ),
  h2: (props: HTMLAttributes<HTMLHeadingElement>) => (
    <h2 {...props} className="mb-2 mt-4 text-base font-semibold first:mt-0" />
  ),
  h3: (props: HTMLAttributes<HTMLHeadingElement>) => (
    <h3 {...props} className="mb-2 mt-3 text-sm font-semibold first:mt-0" />
  ),
  p: (props: HTMLAttributes<HTMLParagraphElement>) => (
    <p {...props} className="mb-2 leading-relaxed last:mb-0" />
  ),
  ul: (props: HTMLAttributes<HTMLUListElement>) => (
    <ul {...props} className="mb-2 list-disc space-y-1 pl-5" />
  ),
  ol: (props: HTMLAttributes<HTMLOListElement>) => (
    <ol {...props} className="mb-2 list-decimal space-y-1 pl-5" />
  ),
  li: (props: HTMLAttributes<HTMLLIElement>) => <li {...props} className="leading-relaxed" />,
  blockquote: (props: HTMLAttributes<HTMLQuoteElement>) => (
    <blockquote
      {...props}
      className="my-2 rounded-r-lg border-l-2 border-primary-300 bg-primary-50/70 px-3 py-2 text-gray-700"
    />
  ),
  a: (props: AnchorHTMLAttributes<HTMLAnchorElement>) => (
    <a
      {...props}
      className="font-medium text-primary-600 underline underline-offset-2 hover:text-primary-700"
      target="_blank"
      rel="noreferrer"
    />
  ),
  pre: (props: HTMLAttributes<HTMLPreElement>) => (
    <pre
      {...props}
      className="my-2 overflow-x-auto rounded-xl bg-gray-900 p-3 text-xs text-gray-100"
    />
  ),
  code: ({ className, ...props }) => (
    <code
      {...props}
      className={`${className ?? ''} rounded bg-gray-200/80 px-1 py-0.5 font-mono text-xs`}
    />
  ),
}

const Markdown = ({ content }: { content: string }) => (
  <div className="break-words text-sm text-inherit">
    <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
      {content}
    </ReactMarkdown>
  </div>
)

type InterviewMessageItemProps = {
  message: InterviewMessage
  user?: {
    id?: string
    name?: string
    avatar?: string
  } | null
}

const ExpandableResult = ({
  message,
  tone,
}: {
  message: InterviewMessage
  tone: 'blue' | 'purple'
}) => {
  const [expanded, setExpanded] = useState(false)
  const isReport = message.kind === 'report'
  const toneClasses =
    tone === 'blue'
      ? 'border-primary-200 bg-primary-50/70 text-primary-800 hover:bg-primary-100/70'
      : 'border-violet-200 bg-violet-50/70 text-violet-800 hover:bg-violet-100/70'

  return (
    <div className={`my-4 overflow-hidden rounded-2xl border ${toneClasses}`}>
      <button
        type="button"
        onClick={() => setExpanded((prev) => !prev)}
        className="flex w-full items-center gap-3 px-4 py-3.5 text-left transition-colors"
        aria-expanded={expanded}
      >
        <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-white/80 shadow-sm ring-1 ring-black/5">
          {isReport ? <ClipboardList className="h-5 w-5" /> : <Sparkles className="h-5 w-5" />}
        </span>
        <span className="flex-1">
          <span className="block text-sm font-semibold">
            {isReport ? '面试评估报告' : '个性化复习计划'}
          </span>
          <span className="mt-0.5 block text-xs font-normal opacity-70">
            {expanded ? '点击收起详细内容' : '点击展开查看完整内容'}
          </span>
        </span>
        {expanded ? <ChevronUp className="h-5 w-5" /> : <ChevronDown className="h-5 w-5" />}
      </button>
      {expanded && (
        <div className="border-t border-current/10 bg-white/80 px-5 py-4 text-gray-800">
          <Markdown content={message.content} />
        </div>
      )}
    </div>
  )
}

const ScoreResult = ({ message }: { message: InterviewMessage }) => {
  const score = Number.isFinite(message.score) ? Number(message.score) : 0
  const scoreTone =
    score >= 70
      ? 'border-emerald-200 bg-emerald-50/80 text-emerald-700'
      : score >= 50
        ? 'border-amber-200 bg-amber-50/80 text-amber-700'
        : 'border-rose-200 bg-rose-50/80 text-rose-700'

  return (
    <div className={`my-4 rounded-2xl border p-4 ${scoreTone}`}>
      <div className="flex items-start gap-4">
        <div className="flex h-14 w-14 flex-shrink-0 flex-col items-center justify-center rounded-2xl bg-white/80 shadow-sm ring-1 ring-black/5">
          <span className="text-2xl font-bold leading-none">{Math.round(score)}</span>
          <span className="mt-1 text-[10px] font-medium uppercase tracking-wide opacity-60">得分</span>
        </div>
        <div className="min-w-0 flex-1 text-gray-700">
          <div className="mb-1 flex items-center gap-2 text-sm font-semibold text-gray-900">
            <Target className="h-4 w-4" />
            本题反馈
          </div>
          <p className="text-sm leading-relaxed">{message.content || '本题评分已完成。'}</p>
        </div>
      </div>

      {message.keyPointsHit && message.keyPointsHit.length > 0 && (
        <div className="mt-4 border-t border-emerald-200/70 pt-3">
          <p className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-emerald-700">
            <CheckCircle2 className="h-3.5 w-3.5" />
            已命中要点
          </p>
          <div className="flex flex-wrap gap-1.5">
            {message.keyPointsHit.map((point, index) => (
              <span
                key={`${point}-${index}`}
                className="rounded-full bg-white/80 px-2.5 py-1 text-xs text-emerald-700 ring-1 ring-emerald-200"
              >
                {point}
              </span>
            ))}
          </div>
        </div>
      )}

      {message.keyPointsMissed && message.keyPointsMissed.length > 0 && (
        <div className="mt-3 border-t border-rose-200/70 pt-3">
          <p className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-rose-700">
            <AlertCircle className="h-3.5 w-3.5" />
            建议补充
          </p>
          <div className="flex flex-wrap gap-1.5">
            {message.keyPointsMissed.map((point, index) => (
              <span
                key={`${point}-${index}`}
                className="rounded-full bg-white/80 px-2.5 py-1 text-xs text-rose-700 ring-1 ring-rose-200"
              >
                {point}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default function InterviewMessageItem({ message, user }: InterviewMessageItemProps) {
  if (message.kind === 'stage' || message.kind === 'system') {
    return (
      <div className="my-3 flex justify-center px-4">
        <span className="inline-flex max-w-full items-center gap-2 rounded-full bg-gray-100 px-3 py-1.5 text-center text-xs text-gray-500 ring-1 ring-gray-200/70">
          {message.kind === 'stage' && <Sparkles className="h-3.5 w-3.5 text-primary-500" />}
          {message.content}
        </span>
      </div>
    )
  }

  if (message.kind === 'error') {
    return (
      <div className="my-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
        <div className="flex items-start gap-2">
          <AlertCircle className="mt-0.5 h-4 w-4 flex-shrink-0" />
          <span>{message.content}</span>
        </div>
      </div>
    )
  }

  if (message.kind === 'score') return <ScoreResult message={message} />
  if (message.kind === 'report') return <ExpandableResult message={message} tone="blue" />
  if (message.kind === 'review_plan') return <ExpandableResult message={message} tone="purple" />

  const isAnswer = message.kind === 'answer'

  return (
    <div className={`my-4 flex items-end gap-2.5 ${isAnswer ? 'justify-end' : 'justify-start'}`}>
      {!isAnswer && (
        <span className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-primary-600 to-primary-500 text-white shadow-sm">
          <Bot className="h-5 w-5" />
        </span>
      )}
      <div
        className={`max-w-[82%] rounded-2xl px-4 py-3 shadow-sm sm:max-w-[74%] ${
          isAnswer
            ? 'rounded-br-md bg-gradient-to-br from-primary-600 to-primary-500 text-white'
            : 'rounded-bl-md bg-gray-100 text-gray-900 ring-1 ring-gray-200/70'
        }`}
      >
        {message.kind === 'question' && (
          <div className="mb-2 flex items-center gap-2 text-xs font-semibold text-primary-700">
            <span className="rounded-full bg-white px-2 py-0.5 ring-1 ring-primary-100">
              {message.questionNum && message.questionNum > 0 ? `第 ${message.questionNum} 题` : '巩固题'}
            </span>
            <span className="font-normal text-gray-400">请结合实际经历作答</span>
          </div>
        )}
        <Markdown content={message.content} />
        {isAnswer && message.deliveryLabel && (
          <p
            className={`mt-2 text-right text-[11px] ${
              message.deliveryState === 'failed'
                ? 'text-rose-100'
                : message.deliveryState === 'confirmed'
                  ? 'text-emerald-100'
                  : 'text-primary-100'
            }`}
          >
            {message.deliveryLabel}
          </p>
        )}
      </div>
      {isAnswer && (
        <UserAvatar
          avatar={user?.avatar}
          name={user?.name || '我'}
          userId={user?.id}
          fallbackSeed={user?.name || 'candidate'}
          className="h-9 w-9 flex-shrink-0 rounded-xl"
        />
      )}
    </div>
  )
}
