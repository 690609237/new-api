import { createFileRoute } from '@tanstack/react-router'

import { ModerationStats } from '@/features/moderation-stats'

export const Route = createFileRoute('/_authenticated/moderation-stats')({
  component: ModerationStats,
})
