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
import { describe, expect, test } from 'vitest'

import { getMessageErrorState } from '../message-error-utils'

describe('getMessageErrorState', () => {
  test('marks an insufficient user quota response as a quota error', () => {
    const state = getMessageErrorState(
      {
        key: 'quota-error',
        from: 'assistant',
        versions: [{ id: 'version', content: 'quota failed' }],
        status: 'error',
        errorCode: 'insufficient_user_quota',
      },
      false
    )

    expect(state).toEqual({
      content: 'quota failed',
      kind: 'quota',
      showSettingsLink: false,
      showChannelsLink: false,
    })
  })

  test('keeps unrelated errors generic', () => {
    const state = getMessageErrorState(
      {
        key: 'generic-error',
        from: 'assistant',
        versions: [{ id: 'version', content: 'request failed' }],
        status: 'error',
        errorCode: 'bad_request',
      },
      false
    )

    expect(state?.kind).toBe('generic')
  })

  test('marks do_request_failed as a channel configuration error', () => {
    expect(
      getMessageErrorState(
        {
          key: 'request-failed',
          from: 'assistant',
          versions: [{ id: 'version', content: 'upstream failed' }],
          status: 'error',
          errorCode: 'do_request_failed',
        },
        true
      )
    ).toEqual({
      content: 'upstream failed',
      kind: 'request-failed',
      showSettingsLink: false,
      showChannelsLink: true,
    })

    expect(
      getMessageErrorState(
        {
          key: 'request-failed-user',
          from: 'assistant',
          versions: [{ id: 'version', content: 'upstream failed' }],
          status: 'error',
          errorCode: 'do_request_failed',
        },
        false
      )?.showChannelsLink
    ).toBe(false)
  })

  test('marks get_channel_failed as a channel configuration error', () => {
    const state = getMessageErrorState(
      {
        key: 'channel-error',
        from: 'assistant',
        versions: [{ id: 'version', content: 'no channel' }],
        status: 'error',
        errorCode: 'get_channel_failed',
      },
      true
    )

    expect(state?.kind).toBe('request-failed')
    expect(state?.showChannelsLink).toBe(true)
  })
})
