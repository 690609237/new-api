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
import { cleanup, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, test } from 'vitest'

import { useStatusContent } from '@/hooks/use-status-content'
import { api } from '@/lib/api'
import { useSystemConfigStore } from '@/stores/system-config-store'

type ApiMethod = (url: string, config?: unknown) => Promise<{ data: unknown }>
type MockableApi = { get: ApiMethod }

const apiClient = api as unknown as MockableApi
const originalGet = apiClient.get
const queryClients: QueryClient[] = []

function createQueryClient(): QueryClient {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  queryClients.push(queryClient)
  return queryClient
}

function wrapper(queryClient: QueryClient) {
  return function QueryWrapper(props: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {props.children}
      </QueryClientProvider>
    )
  }
}

beforeEach(() => {
  window.localStorage.clear()
  useSystemConfigStore.setState(useSystemConfigStore.getInitialState(), true)
})

afterEach(() => {
  cleanup()
  queryClients.splice(0).forEach((client) => client.clear())
  apiClient.get = originalGet
  window.localStorage.clear()
  useSystemConfigStore.setState(useSystemConfigStore.getInitialState(), true)
})

describe('dedicated status content queries', () => {
  test('loads enabled content from its dedicated endpoint', async () => {
    const requests: string[] = []
    apiClient.get = async (url) => {
      requests.push(url)
      if (url === '/api/status') {
        return {
          data: {
            success: true,
            data: {
              announcements_enabled: true,
              announcements: [{ content: 'legacy bundled content' }],
            },
          },
        }
      }
      if (url === '/api/status/announcements') {
        return {
          data: {
            success: true,
            data: [{ content: 'dedicated content' }],
          },
        }
      }
      throw new Error(`Unexpected GET ${url}`)
    }

    const hook = renderHook(
      () =>
        useStatusContent<{ content: string }>(
          'announcements_enabled',
          'announcements'
        ),
      { wrapper: wrapper(createQueryClient()) }
    )

    await waitFor(() =>
      expect(hook.result.current.items).toEqual([
        { content: 'dedicated content' },
      ])
    )
    expect(requests).toEqual(['/api/status', '/api/status/announcements'])
  })

  test('does not request content when its status flag is disabled', async () => {
    const requests: string[] = []
    apiClient.get = async (url) => {
      requests.push(url)
      return {
        data: {
          success: true,
          data: { faq_enabled: false },
        },
      }
    }

    const hook = renderHook(() => useStatusContent('faq_enabled', 'faq'), {
      wrapper: wrapper(createQueryClient()),
    })

    await waitFor(() => expect(hook.result.current.loading).toBe(false))
    expect(hook.result.current.items).toEqual([])
    expect(requests).toEqual(['/api/status'])
  })
})
