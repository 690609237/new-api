/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the License,
or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, waitFor, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ModerationSection } from '../moderation-section'

const defaultValues = {
  ModerationEnabled: true,
  ModerationBeforeChannel: false,
  ModerationBaseURL: 'https://api.openai.com/v1',
  ModerationAPIKey: '',
  ModerationModel: 'omni-moderation-latest',
  ModerationScoreThreshold: 0.6,
  ModerationAlertEmail: '',
  ModerationAlertThreshold: 3,
  ModerationCacheTTLSeconds: 300,
  ModerationExemptUserIDs: '',
  ModerationExemptGroups: '',
  ModerationSampleRate: 100,
  ModerationForceUserIDs: '',
  ModerationForceTokenIDs: '',
  ModerationTimeoutSeconds: 5,
  ModerationTimeoutWindowSeconds: 300,
  ModerationTimeoutThreshold: 3,
  ModerationTimeoutPauseSeconds: 60,
  DailyReviewEnabled: false,
  DailyReviewHour: 2,
  DailyReviewPrompt: 'Review user messages',
  DailyReviewBaseURL: 'https://api.openai.com/v1',
  DailyReviewModel: 'gpt-6-luna',
  DailyReviewAPIKey: '',
}

function renderModerationSection() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <ModerationSection defaultValues={defaultValues} />
    </QueryClientProvider>
  )
}

describe('content moderation layout', () => {
  it('rejects an invalid moderation score threshold in the settings form', async () => {
    const view = renderModerationSection()
    const threshold = within(view.container).getByLabelText(
      'Moderation score threshold'
    )
    expect(threshold).toHaveValue(0.6)

    fireEvent.change(threshold, { target: { value: '0' } })
    await waitFor(() =>
      expect(threshold).toHaveAttribute('aria-invalid', 'true')
    )
  })

  it('shows the daily review configuration and rejects an out-of-range hour', async () => {
    const view = renderModerationSection()
    const hour = within(view.container).getByLabelText('Daily review hour')
    expect(hour).toHaveValue(2)
    expect(
      within(view.container).getByLabelText('Daily review prompt')
    ).toHaveValue('Review user messages')
    expect(
      within(view.container).getByRole('button', { name: 'Review today now' })
    ).toBeEnabled()

    fireEvent.change(hour, { target: { value: '24' } })
    await waitFor(() => expect(hour).toHaveAttribute('aria-invalid', 'true'))
  })

  it('keeps the connection test with the API key input', () => {
    const view = renderModerationSection()
    const connectionRegion = view.container.querySelector(
      '[data-moderation-layout="connection"]'
    )

    expect(connectionRegion).not.toBeNull()
    expect(
      within(connectionRegion as HTMLElement).getByLabelText(
        'Moderation API key'
      )
    ).toBeTruthy()
    expect(
      within(connectionRegion as HTMLElement).getByRole('button', {
        name: 'Test moderation connection',
      })
    ).toBeTruthy()
  })

  it('groups forced user and token IDs after both exemption fields', () => {
    const view = renderModerationSection()
    const audienceRegion = view.container.querySelector(
      '[data-moderation-layout="audience-rules"]'
    )

    expect(audienceRegion).not.toBeNull()
    const region = within(audienceRegion as HTMLElement)
    const exemptUsers = region.getByLabelText('Exempt user IDs')
    const exemptGroups = region.getByLabelText('Exempt user groups')
    const forcedUsers = region.getByLabelText('Required moderation user IDs')
    const forcedTokens = region.getByLabelText('Required moderation token IDs')

    expect(
      exemptUsers.compareDocumentPosition(forcedUsers) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(
      exemptGroups.compareDocumentPosition(forcedUsers) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(
      exemptUsers.compareDocumentPosition(forcedTokens) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(
      forcedUsers.compareDocumentPosition(forcedTokens) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(
      exemptGroups.compareDocumentPosition(forcedTokens) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(audienceRegion?.nextElementSibling).toHaveTextContent(
      'Daily content review'
    )
  })
})
