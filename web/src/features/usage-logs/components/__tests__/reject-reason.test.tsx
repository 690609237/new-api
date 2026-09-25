/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, test } from 'vitest'

import type { UsageLog } from '../../data/schema'
import type { LogOtherData } from '../../types'
import { DetailsDialog } from '../dialogs/details-dialog'

const queryClients: QueryClient[] = []

function makeLog(other: LogOtherData): UsageLog {
  return {
    id: 1,
    user_id: 1,
    created_at: 1,
    type: 5,
    content: 'request rejected',
    username: 'user',
    token_name: 'token',
    model_name: 'gpt-test',
    quota: 0,
    prompt_tokens: 0,
    completion_tokens: 0,
    use_time: 0,
    is_stream: false,
    channel: 1,
    channel_name: 'channel',
    token_id: 1,
    group: 'default',
    ip: '',
    other: JSON.stringify(other),
    request_id: 'req-1',
    upstream_request_id: '',
  }
}

function renderDetails(
  isAdmin: boolean,
  other: LogOtherData = {
    admin_info: { reject_reason: 'blocked by channel policy' },
  }
): void {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const freshAt = Date.now() + 60_000
  queryClient.setQueryData(['status'], {}, { updatedAt: freshAt })
  queryClient.setQueryData(
    ['usage-log-moderation-detail', 'req-1', 5],
    { other: JSON.stringify(other) },
    { updatedAt: freshAt }
  )
  queryClients.push(queryClient)

  render(
    <QueryClientProvider client={queryClient}>
      <DetailsDialog
        log={makeLog(other)}
        isAdmin={isAdmin}
        isRoot={false}
        open
        onOpenChange={() => undefined}
      />
    </QueryClientProvider>
  )
}

afterEach(() => {
  for (const queryClient of queryClients) {
    queryClient.clear()
  }
  queryClients.length = 0
})

describe('usage log reject reason', () => {
  test('shows the nested admin reject reason to admins', () => {
    renderDetails(true)

    expect(screen.getByText('Reject Reason')).toBeInTheDocument()
    expect(screen.getByText('blocked by channel policy')).toBeInTheDocument()
  })

  test('hides the nested admin reject reason from non-admin users', () => {
    renderDetails(false)

    expect(screen.queryByText('Reject Reason')).toBeNull()
    expect(screen.queryByText('blocked by channel policy')).toBeNull()
  })
})

describe('moderation rejection details', () => {
  const moderationLog: LogOtherData = {
    admin_info: {
      moderation: {
        policy: 'moderation_api',
        flagged: true,
        prompt: 'sample prompt',
        rules: ['violence', 'harassment'],
        scores: { violence: 0.827431, harassment: 0.6 },
        threshold: 0.6,
      },
    },
  }

  test('shows matched categories, original scores and threshold to admins', () => {
    renderDetails(true, moderationLog)

    expect(screen.getByText('Matched categories and scores')).toBeVisible()
    expect(screen.getByText('violence: 0.827431')).toBeVisible()
    expect(screen.getByText('harassment: 0.6')).toBeVisible()
    expect(screen.getByText('Score threshold')).toBeVisible()
    expect(screen.getByText('sample prompt')).toBeVisible()
  })

  test('hides matched categories and scores from non-admin users', () => {
    renderDetails(false, moderationLog)

    expect(screen.queryByText('violence: 0.827431')).toBeNull()
    expect(screen.queryByText('sample prompt')).toBeNull()
  })
})
