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
import assert from 'node:assert/strict'
import { after, test } from 'node:test'

import { Window } from 'happy-dom'

const domWindow = new Window({ url: 'http://localhost/dashboard' })
const domGlobals = [
  'window',
  'document',
  'navigator',
  'HTMLElement',
  'SVGElement',
  'Node',
  'Element',
  'Event',
  'CustomEvent',
  'MutationObserver',
  'requestAnimationFrame',
  'cancelAnimationFrame',
  'getComputedStyle',
] as const

for (const key of domGlobals) {
  Object.defineProperty(globalThis, key, {
    configurable: true,
    value: domWindow[key],
  })
}

Object.defineProperty(globalThis, 'Image', {
  configurable: true,
  value: domWindow.Image,
})
Object.defineProperty(globalThis, 'scrollTo', {
  configurable: true,
  value: () => undefined,
})

const { act } = await import('react')
const { createRoot } = await import('react-dom/client')
const { QueryClient, QueryClientProvider } =
  await import('@tanstack/react-query')
const {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} = await import('@tanstack/react-router')
const { createInstance } = await import('i18next')
const { I18nextProvider, initReactI18next } = await import('react-i18next')
const { SystemBrand } = await import('../system-brand')
const { useSystemConfigStore } = await import('@/stores/system-config-store')

const reactTestGlobals = globalThis as typeof globalThis & {
  IS_REACT_ACT_ENVIRONMENT?: boolean
}
reactTestGlobals.IS_REACT_ACT_ENVIRONMENT = true

const i18n = createInstance()
await i18n.use(initReactI18next).init({
  lng: 'zh',
  resources: {
    zh: {
      translation: {
        'Contact the author': '联系作者交流',
        WeChat: '微信',
      },
    },
  },
})

test('inline system brand shows all contact methods in the mobile app header', async () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  queryClient.setQueryData(['status'], { system_name: 'Test' })
  useSystemConfigStore.setState((state) => ({
    ...state,
    loading: false,
    loadedLogoUrl: state.config.logo,
  }))

  const rootRoute = createRootRoute({
    component: () => <SystemBrand variant='inline' />,
  })
  const router = createRouter({
    routeTree: rootRoute,
    history: createMemoryHistory({ initialEntries: ['/dashboard'] }),
  })
  const container = document.createElement('div')
  document.body.append(container)
  const root = createRoot(container)

  await act(async () => {
    root.render(
      <I18nextProvider i18n={i18n}>
        <QueryClientProvider client={queryClient}>
          <RouterProvider router={router} />
        </QueryClientProvider>
      </I18nextProvider>
    )
    await router.load()
  })

  const mobileContact = container.querySelector(
    '[data-slot="brand-contact-line"][data-variant="compact"]'
  )
  assert.ok(mobileContact)
  assert.match(mobileContact.textContent || '', /QQ 1549277597/)
  assert.match(mobileContact.textContent || '', /微信 ModelPass/)
  assert.match(mobileContact.textContent || '', /1549277597@qq\.com/)
  assert.equal(mobileContact.classList.contains('absolute'), true)
  assert.equal(mobileContact.classList.contains('lg:hidden'), true)

  await act(async () => root.unmount())
  container.remove()
  queryClient.clear()
})

after(() => {
  domWindow.close()
})
