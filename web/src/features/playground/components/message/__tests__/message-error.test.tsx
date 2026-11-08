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
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, test } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import type { Message } from '../../../types'
import { MessageError } from '../message-error'

const transportErrorMessage: Message = {
  key: 'transport-error',
  from: 'assistant',
  versions: [
    {
      id: 'version-1',
      content:
        'The upstream host could not be resolved. Check the API Base URL and DNS settings.',
    },
  ],
  status: 'error',
  errorCode: 'do_request_failed',
}

afterEach(() => {
  useAuthStore.getState().auth.setUser(null)
})

describe('MessageError upstream configuration guidance', () => {
  test('shows the transport reason and administrator guidance to regular users', () => {
    useAuthStore.getState().auth.setUser({ id: 2, username: 'user', role: 1 })

    render(<MessageError message={transportErrorMessage} />)

    expect(
      screen.getByText(
        'The upstream host could not be resolved. Check the API Base URL and DNS settings.'
      )
    ).toBeInTheDocument()
    expect(
      screen.getByText(
        'The model request failed. Please contact an administrator to verify the channel configuration.'
      )
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('link', { name: 'Channels' })
    ).not.toBeInTheDocument()
  })
})
