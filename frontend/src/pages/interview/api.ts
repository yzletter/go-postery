import { apiGet, API_BASE_URL } from '../../utils/api'
import type { InterviewClientEvent, InterviewServerEvent } from './types'

type OSSUploadPolicy = {
  policy?: string
  signature?: string
  x_oss_signature_version?: string
  x_oss_credential?: string
  x_oss_date?: string
  security_token?: string
  host?: string
  dir?: string
  callback?: string
}

type QuestionUploadSignResponse = {
  response?: string
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null

const asString = (value: unknown) => {
  if (typeof value === 'string') return value.trim()
  if (typeof value === 'number' && Number.isFinite(value)) return String(value)
  return ''
}

const asStringList = (value: unknown) => {
  if (!Array.isArray(value)) return []
  return value
    .map((item) => asString(item))
    .filter(Boolean)
}

export const buildInterviewWsUrl = () => {
  const base = new URL(API_BASE_URL, window.location.origin)
  base.protocol = base.protocol === 'https:' ? 'wss:' : 'ws:'
  base.pathname = `${base.pathname.replace(/\/+$/, '')}/ws/interview`
  base.search = ''
  base.hash = ''
  return base.toString()
}

export const buildInterviewEnvelope = (event: InterviewClientEvent) => ({
  biz_type: 'interview',
  biz_data: event,
})

export const parseInterviewServerEvent = (raw: unknown): InterviewServerEvent | null => {
  if (!isRecord(raw)) return null

  const bizType = asString(raw.biz_type)
  if (bizType && !bizType.startsWith('interview_')) return null

  const payload = isRecord(raw.biz_data) ? raw.biz_data : raw
  const type = asString(payload.type)
  const sessionId = asString(payload.session_id)

  switch (type) {
    case 'stage_change':
      return {
        type,
        sessionId: sessionId || undefined,
        stage: asString(payload.stage),
        message: asString(payload.message),
      }
    case 'question':
      return {
        type,
        sessionId: sessionId || undefined,
        questionNum: Number.isFinite(Number(payload.question_num)) ? Number(payload.question_num) : 0,
        content: asString(payload.content),
      }
    case 'score':
      return {
        type,
        sessionId: sessionId || undefined,
        score: Number.isFinite(Number(payload.score)) ? Number(payload.score) : 0,
        feedback: asString(payload.feedback),
        keyPointsHit: asStringList(payload.key_points_hit),
        keyPointsMissed: asStringList(payload.key_points_missed),
      }
    case 'report':
    case 'review_plan':
      return {
        type,
        sessionId: sessionId || undefined,
        content: asString(payload.content),
      }
    default:
      return null
  }
}

const parseUploadPolicy = (raw: unknown): Required<OSSUploadPolicy> => {
  if (typeof raw !== 'string' || !raw.trim()) {
    throw new Error('题库上传签名无效')
  }

  let parsed: OSSUploadPolicy
  try {
    parsed = JSON.parse(raw) as OSSUploadPolicy
  } catch {
    throw new Error('题库上传签名格式错误')
  }

  const policy = parsed.policy?.trim()
  const signature = parsed.signature?.trim()
  const signatureVersion = parsed.x_oss_signature_version?.trim() || 'OSS4-HMAC-SHA256'
  const credential = parsed.x_oss_credential?.trim()
  const date = parsed.x_oss_date?.trim()
  const securityToken = parsed.security_token?.trim()
  const host = parsed.host?.trim()
  const dir = parsed.dir?.trim()
  const callback = parsed.callback?.trim()

  if (!policy || !signature || !credential || !date || !securityToken || !host || !dir || !callback) {
    throw new Error('题库上传签名字段不完整')
  }

  return {
    policy,
    signature,
    x_oss_signature_version: signatureVersion,
    x_oss_credential: credential,
    x_oss_date: date,
    security_token: securityToken,
    host,
    dir,
    callback,
  }
}

const sanitizeObjectFileName = (fileName: string) => {
  const normalized = Array.from(fileName.normalize('NFKC'))
    .map((character) => {
      const codePoint = character.charCodeAt(0)
      return character === '/' || character === '\\' || codePoint < 32 || codePoint === 127
        ? '-'
        : character
    })
    .join('')
    .trim()
  if (!normalized || normalized === '.' || normalized === '..') {
    return `questions-${Date.now()}.md`
  }
  return normalized.slice(0, 180)
}

export const uploadQuestionBank = async (file: File) => {
  const { data } = await apiGet<QuestionUploadSignResponse>('/interviews/questions/upload')
  const policy = parseUploadPolicy(data?.response)
  const objectKey = `${policy.dir}${sanitizeObjectFileName(file.name)}`

  const formData = new FormData()
  formData.append('success_action_status', '200')
  formData.append('policy', policy.policy)
  formData.append('x-oss-signature', policy.signature)
  formData.append('x-oss-signature-version', policy.x_oss_signature_version)
  formData.append('x-oss-credential', policy.x_oss_credential)
  formData.append('x-oss-date', policy.x_oss_date)
  formData.append('key', objectKey)
  formData.append('x-oss-security-token', policy.security_token)
  formData.append('callback', policy.callback)
  formData.append('file', file)

  const response = await fetch(policy.host, {
    method: 'POST',
    body: formData,
  })

  const responseText = await response.text()
  if (response.status !== 200) {
    throw new Error(
      response.status === 203
        ? '题库已上传，但后端回调或解析失败'
        : '题库上传失败，请稍后重试'
    )
  }

  if (responseText.trim()) {
    let callbackResult: unknown
    try {
      callbackResult = JSON.parse(responseText) as unknown
    } catch {
      throw new Error('题库回调响应格式错误')
    }

    if (isRecord(callbackResult)) {
      const code = Number(callbackResult.code)
      if (Number.isFinite(code) && code !== 0) {
        throw new Error(asString(callbackResult.msg) || '题库解析失败')
      }
    }
  }

  return objectKey
}
