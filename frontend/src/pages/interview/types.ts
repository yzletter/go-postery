export type InterviewConnectionStatus =
  | 'connecting'
  | 'connected'
  | 'disconnected'
  | 'error'

export type InterviewMessage = {
  id: string
  kind: 'answer' | 'error' | 'question' | 'report' | 'review_plan' | 'score' | 'stage' | 'system'
  content: string
  deliveryLabel?: string
  deliveryState?: 'confirmed' | 'failed' | 'pending'
  questionNum?: number
  score?: number
  keyPointsHit?: string[]
  keyPointsMissed?: string[]
}

type InterviewServerEventBase = {
  sessionId?: string
}

export type InterviewServerEvent =
  | (InterviewServerEventBase & {
      type: 'stage_change'
      stage: string
      message: string
    })
  | (InterviewServerEventBase & {
      type: 'question'
      questionNum: number
      content: string
    })
  | (InterviewServerEventBase & {
      type: 'score'
      score: number
      feedback: string
      keyPointsHit: string[]
      keyPointsMissed: string[]
    })
  | (InterviewServerEventBase & {
      type: 'report' | 'review_plan'
      content: string
    })

export type InterviewClientEvent =
  | {
      type: 'start_interview'
      jd: string
      resume: string
      candidate_name: string
    }
  | {
      type: 'answer'
      session_id: string
      answer: string
    }
  | {
      type: 'cancel_interview'
      session_id: string
    }
