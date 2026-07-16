import { Check, Loader2 } from 'lucide-react'

const stages = [
  { key: 'jd', label: 'JD 分析' },
  { key: 'resume', label: '简历匹配' },
  { key: 'plan', label: '出题规划' },
  { key: 'interview', label: '模拟面试' },
  { key: 'evaluation', label: '评估报告' },
  { key: 'review', label: '复习计划' },
] as const

const resolveStageIndex = (stage: string) => {
  if (!stage) return -1
  if (stage === 'completed') return stages.length
  if (stage.startsWith('review_plan')) return 5
  if (
    stage.startsWith('evaluation') ||
    stage.startsWith('review_weak') ||
    stage === 'terminated'
  ) {
    return 4
  }
  if (stage.startsWith('interview')) return 3
  if (stage.startsWith('question_plan') || stage === 'memory_loaded') return 2
  if (stage.startsWith('resume_match')) return 1
  if (stage.startsWith('jd_analysis')) return 0
  return -1
}

type InterviewStageBarProps = {
  currentStage: string
}

export default function InterviewStageBar({ currentStage }: InterviewStageBarProps) {
  const currentIndex = resolveStageIndex(currentStage)

  return (
    <div className="overflow-x-auto border-b border-gray-100 bg-white/70 px-4 py-3 sm:px-6">
      <ol className="flex min-w-[660px] items-center" aria-label="面试进度">
        {stages.map((stage, index) => {
          const done = currentIndex > index
          const active = currentIndex === index

          return (
            <li key={stage.key} className="flex flex-1 items-center last:flex-none">
              <div
                className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1.5 text-xs font-medium transition-colors ${
                  done
                    ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200'
                    : active
                      ? 'bg-primary-50 text-primary-700 ring-1 ring-primary-200'
                      : 'bg-gray-50 text-gray-400 ring-1 ring-gray-200'
                }`}
                aria-current={active ? 'step' : undefined}
              >
                <span
                  className={`flex h-4 w-4 items-center justify-center rounded-full ${
                    done
                      ? 'bg-emerald-500 text-white'
                      : active
                        ? 'bg-primary-100 text-primary-700'
                        : 'bg-gray-200 text-gray-400'
                  }`}
                >
                  {done ? (
                    <Check className="h-3 w-3" strokeWidth={3} />
                  ) : active ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    index + 1
                  )}
                </span>
                {stage.label}
              </div>
              {index < stages.length - 1 && (
                <div
                  className={`mx-2 h-px flex-1 ${done ? 'bg-emerald-300' : 'bg-gray-200'}`}
                  aria-hidden
                />
              )}
            </li>
          )
        })}
      </ol>
    </div>
  )
}
