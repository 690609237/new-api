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
import { render, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ModerationSection } from '../moderation-section'

const defaultValues = {
  ModerationEnabled: true,
  ModerationBeforeChannel: false,
  ModerationBaseURL: 'https://api.openai.com/v1',
  ModerationAPIKey: '',
  ModerationModel: 'omni-moderation-latest',
  ModerationAlertEmail: '',
  ModerationAlertThreshold: 3,
  ModerationCacheTTLSeconds: 300,
  ModerationExemptUserIDs: '',
  ModerationExemptGroups: '',
  ModerationSampleRate: 100,
  ModerationForceTokenIDs: '',
  ModerationTimeoutSeconds: 5,
  ModerationTimeoutWindowSeconds: 300,
  ModerationTimeoutThreshold: 3,
  ModerationTimeoutPauseSeconds: 60,
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

  it('groups forced tokens after both exemption fields at the bottom', () => {
    const view = renderModerationSection()
    const audienceRegion = view.container.querySelector(
      '[data-moderation-layout="audience-rules"]'
    )

    expect(audienceRegion).not.toBeNull()
    const region = within(audienceRegion as HTMLElement)
    const exemptUsers = region.getByLabelText('Exempt user IDs')
    const exemptGroups = region.getByLabelText('Exempt user groups')
    const forcedTokens = region.getByLabelText('Required moderation token IDs')

    expect(
      exemptUsers.compareDocumentPosition(forcedTokens) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(
      exemptGroups.compareDocumentPosition(forcedTokens) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(audienceRegion?.nextElementSibling).toBeNull()
  })
})
