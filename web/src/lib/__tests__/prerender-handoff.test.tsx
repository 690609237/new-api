/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or (at your
option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router'
import {
  act,
  cleanup,
  render,
  renderHook,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { useSystemConfig } from '@/hooks/use-system-config'
import { api } from '@/lib/api'
import { STATUS_QUERY_KEY, STATUS_STORAGE_KEY } from '@/lib/status-query'
import { useSystemConfigStore } from '@/stores/system-config-store'

import { installPrerenderHandoff } from '../prerender-handoff'

type MockableApi = {
  get: (url: string, config?: unknown) => Promise<{ data: unknown }>
}

const apiClient = api as unknown as MockableApi
const originalGet = apiClient.get

beforeEach(() => {
  window.localStorage.clear()
  useSystemConfigStore.setState(useSystemConfigStore.getInitialState(), true)
})

afterEach(() => {
  cleanup()
  document.body.replaceChildren()
  apiClient.get = originalGet
  window.localStorage.clear()
  useSystemConfigStore.setState(useSystemConfigStore.getInitialState(), true)
})

it('keeps prerendered branding while the live status request is pending', async () => {
  document.body.innerHTML = `<main id="seo-content"><div id="home-prerender"><script id="home-static-bootstrap" type="application/json">${JSON.stringify(
    {
      version: 1,
      status: { HeaderNavModules: { pricing: false } },
      systemConfig: {
        systemName: 'Build Snapshot',
        logo: '/snapshot.svg',
      },
    }
  )}</script></div></main><div id="root"></div>`
  const router = {
    state: { location: { pathname: '/' }, matches: [] },
    subscribe: () => () => {},
  }
  installPrerenderHandoff(router)

  let resolveStatus!: (value: { data: unknown }) => void
  const pendingStatus = new Promise<{ data: unknown }>((resolve) => {
    resolveStatus = resolve
  })
  const requests: string[] = []
  apiClient.get = async (url) => {
    requests.push(url)
    return pendingStatus
  }

  const queryClient = new QueryClient()
  const hook = renderHook(() => useSystemConfig({ autoLoad: true }), {
    wrapper: ({ children }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
  })

  await waitFor(() => expect(requests).toEqual(['/api/status']))
  expect(hook.result.current.systemName).toBe('Build Snapshot')
  expect(hook.result.current.logo).toBe('/snapshot.svg')
  expect(hook.result.current.loading).toBe(false)
  expect(queryClient.getQueryData(STATUS_QUERY_KEY)).toBeUndefined()
  expect(window.localStorage.getItem(STATUS_STORAGE_KEY)).toBeNull()

  resolveStatus({ data: { success: true, data: { system_name: 'Live' } } })
  await waitFor(() => expect(hook.result.current.systemName).toBe('Live'))
  queryClient.clear()
})

it('keeps the rendered home visible while the route is pending and removes it once the page is ready', async () => {
  document.body.innerHTML =
    '<main id="seo-content"><div id="home-prerender"><h1>Static home</h1></div></main><div id="root"></div>'
  vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
  let resolveReady!: () => void
  const ready = new Promise<void>((resolve) => {
    resolveReady = resolve
  })
  const rootRoute = createRootRoute({ component: Outlet })
  const homeRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    beforeLoad: () => ready,
    component: () => <h1 data-page='home'>Interactive home</h1>,
  })
  const router = createRouter({
    routeTree: rootRoute.addChildren([homeRoute]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  const dispose = installPrerenderHandoff(router)
  const client = new QueryClient()
  const rootElement = document.querySelector('#root')
  if (!rootElement) throw new Error('root element missing')
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
    {
      container: rootElement,
    }
  )

  expect(
    screen.getByRole('heading', { name: 'Static home' })
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('heading', { name: 'Interactive home' })
  ).not.toBeInTheDocument()

  await act(async () => resolveReady())
  expect(
    await screen.findByRole('heading', { name: 'Interactive home' })
  ).toBeInTheDocument()
  await waitFor(() =>
    expect(
      screen.queryByRole('heading', { name: 'Static home' })
    ).not.toBeInTheDocument()
  )
  dispose()
  client.clear()
})

it('preserves the static home when a lazy route fails and hands off after navigation to another page', async () => {
  document.body.innerHTML =
    '<main id="seo-content"><div id="home-prerender"><h1>Static home</h1></div></main><div id="root"></div>'
  vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
  const rootRoute = createRootRoute({ component: Outlet })
  const homeRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    errorComponent: () => <h1>Route failed</h1>,
  }).lazy(async () => {
    throw new Error('Failed to download home chunk')
  })
  const docsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/docs',
    component: () => <h1>Documentation</h1>,
  })
  const router = createRouter({
    routeTree: rootRoute.addChildren([homeRoute, docsRoute]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  const dispose = installPrerenderHandoff(router)
  const rootElement = document.querySelector('#root')
  if (!rootElement) throw new Error('root element missing')
  render(<RouterProvider router={router} />, {
    container: rootElement,
  })
  await screen.findByRole('heading', { name: 'Route failed' })
  expect(
    screen.getByRole('heading', { name: 'Static home' })
  ).toBeInTheDocument()

  await act(async () => {
    await router.navigate({ to: '/docs' })
  })
  await screen.findByRole('heading', { name: 'Documentation' })
  await waitFor(() => expect(document.querySelector('#seo-content')).toBeNull())
  dispose()
})
