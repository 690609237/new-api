import { api } from '@/lib/api'

export type ModerationUsageBucket = {
  bucket_start: number
  user_id: number
  token_id: number
  api_requests: number
  api_passed: number
  api_violations: number
  api_failed: number
  cache_hits: number
  api_latency_total_ms: number
  api_timeouts: number
  circuit_skips: number
  sensitive_word_hits: number
}

export type ModerationUsageDimension = Omit<
  ModerationUsageBucket,
  'bucket_start'
>

export type ModerationUsageResponse = {
  success: boolean
  message: string
  data?: {
    start_timestamp: number
    end_timestamp: number
    api_requests: number
    api_passed: number
    api_violations: number
    api_succeeded: number
    api_failed: number
    cache_hits: number
    api_latency_total_ms: number
    api_latency_average_ms: number
    api_timeouts: number
    circuit_skips: number
    sensitive_word_hits: number
    buckets: ModerationUsageBucket[]
    dimensions: ModerationUsageDimension[]
  }
}

export async function getModerationUsageStats(params: {
  start_timestamp: number
  end_timestamp: number
  user_id?: number
  token_id?: number
}) {
  const response = await api.get<ModerationUsageResponse>(
    '/api/moderation/stats',
    { params }
  )
  if (!response.data.success || !response.data.data) {
    throw new Error(
      response.data.message || 'Failed to load moderation statistics'
    )
  }
  return response.data.data
}
