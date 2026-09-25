/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { ModerationStats } from '..'
import * as moderationStatsAPI from '../api'

describe('moderation statistics dashboard', () => {
  it('shows sensitive word hits without check-count or hit-rate metrics', async () => {
    vi.spyOn(moderationStatsAPI, 'getModerationUsageStats').mockResolvedValue({
      start_timestamp: 1_700_000_000,
      end_timestamp: 1_700_000_300,
      api_requests: 3,
      api_passed: 2,
      api_violations: 1,
      api_succeeded: 3,
      api_failed: 0,
      cache_hits: 0,
      api_latency_total_ms: 30,
      api_latency_average_ms: 10,
      api_timeouts: 0,
      circuit_skips: 0,
      sensitive_word_hits: 5,
      buckets: [],
      dimensions: [
        {
          user_id: 7,
          token_id: 9,
          api_requests: 3,
          api_passed: 2,
          api_violations: 1,
          api_failed: 0,
          cache_hits: 0,
          api_latency_total_ms: 30,
          api_timeouts: 0,
          circuit_skips: 0,
          sensitive_word_hits: 5,
        },
      ],
    })
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <ModerationStats />
      </QueryClientProvider>
    )

    expect(await screen.findAllByText('Sensitive word hits')).not.toHaveLength(
      0
    )
    expect(screen.getAllByText('5')).not.toHaveLength(0)
    expect(screen.queryByText('Sensitive word checks')).not.toBeInTheDocument()
    expect(
      screen.queryByText('Sensitive word hit rate')
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Daily review results' })
    ).toBeInTheDocument()
  })

  it('opens the latest available report with its log date and highlighted high-risk row', async () => {
    vi.spyOn(moderationStatsAPI, 'getModerationUsageStats').mockResolvedValue({
      start_timestamp: 0,
      end_timestamp: 0,
      api_requests: 0,
      api_passed: 0,
      api_violations: 0,
      api_succeeded: 0,
      api_failed: 0,
      cache_hits: 0,
      api_latency_total_ms: 0,
      api_latency_average_ms: 0,
      api_timeouts: 0,
      circuit_skips: 0,
      sensitive_word_hits: 0,
      buckets: [],
      dimensions: [],
    })
    const getReport = vi
      .spyOn(moderationStatsAPI, 'getDailyReviewReport')
      .mockResolvedValue({
        date: '20260924',
        available_dates: ['20260924', '20260922'],
        content:
          '| username | 行为 | 风险分档 | 简要描述 | 示例requestid（最多3个） |\n| --- | --- | --- | --- | --- |\n| alice | 测试 | 高 | 描述 | req1 |',
      })
    render(
      <QueryClientProvider
        client={
          new QueryClient({ defaultOptions: { queries: { retry: false } } })
        }
      >
        <ModerationStats />
      </QueryClientProvider>
    )

    fireEvent.click(
      screen.getByRole('button', { name: 'Daily review results' })
    )
    await waitFor(() => expect(getReport).toHaveBeenCalledWith(undefined))
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(await screen.findByText('Log date: 2026-09-24')).toBeInTheDocument()
    expect(screen.getByText('高').closest('mark')).not.toBeNull()
  })
})
