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
import { after, describe, test } from 'node:test'

import { Window } from 'happy-dom'

const domWindow = new Window({ url: 'http://localhost/pricing' })
const domGlobals = [
  'window',
  'document',
  'navigator',
  'HTMLElement',
  'HTMLButtonElement',
  'SVGElement',
  'Node',
  'Element',
  'Event',
  'MouseEvent',
  'CustomEvent',
  'MutationObserver',
  'ResizeObserver',
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
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} = await import('@tanstack/react-router')
const { createInstance } = await import('i18next')
const { I18nextProvider, initReactI18next } = await import('react-i18next')
const { PublicHeader } = await import('../public-header')
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
        'Model Square': '模型广场',
        Rankings: '排行榜',
        WeChat: '微信',
      },
    },
  },
})

async function renderHeader(path: '/' | '/pricing') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  queryClient.setQueryData(['status'], {
    header_nav_modules: JSON.stringify({
      home: false,
      console: false,
      pricing: { enabled: true, requireAuth: false },
      rankings: { enabled: true, requireAuth: false },
      docs: false,
      about: false,
    }),
    announcements_enabled: false,
  })
  queryClient.setQueryData(['notice'], { success: true, data: '' })
  useSystemConfigStore.setState((state) => ({
    ...state,
    loading: false,
    loadedLogoUrl: state.config.logo,
  }))

  const HeaderFixture = () => (
    <PublicHeader
      siteName='Test'
      showAuthButtons={false}
      showLanguageSwitcher={false}
      showNotifications={false}
      showThemeSwitch={false}
    />
  )
  const rootRoute = createRootRoute({ component: Outlet })
  const homeRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: HeaderFixture,
  })
  const pricingRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/pricing',
    component: HeaderFixture,
  })
  const router = createRouter({
    routeTree: rootRoute.addChildren([homeRoute, pricingRoute]),
    history: createMemoryHistory({ initialEntries: [path] }),
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

  return {
    container,
    cleanup: async () => {
      await act(async () => root.unmount())
      container.remove()
      queryClient.clear()
    },
  }
}

describe('public header navigation layout', () => {
  test('keeps desktop labels horizontal and uses the mobile menu below the large breakpoint', async () => {
    const rendered = await renderHeader('/pricing')

    const pricingLink = [...rendered.container.querySelectorAll('a')].find(
      (link) => link.textContent === '模型广场'
    )
    assert.ok(pricingLink)
    assert.equal(pricingLink.classList.contains('whitespace-nowrap'), true)
    assert.equal(pricingLink.classList.contains('shrink-0'), true)

    const desktopNavigation = pricingLink.parentElement
    assert.ok(desktopNavigation)
    assert.equal(desktopNavigation.classList.contains('lg:flex'), true)
    assert.equal(desktopNavigation.classList.contains('sm:flex'), false)

    const menuButton = rendered.container.querySelector(
      'button[aria-label="Toggle navigation menu"]'
    )
    assert.ok(menuButton)
    assert.equal(
      menuButton.parentElement?.classList.contains('lg:hidden'),
      true
    )

    await rendered.cleanup()
  })

  test('shows all contact methods in a wrapping mobile brand line on public pages', async () => {
    const rendered = await renderHeader('/pricing')
    const mobileContact = rendered.container.querySelector(
      '[data-slot="brand-contact-line"][data-variant="compact"]'
    )

    assert.ok(mobileContact)
    assert.match(mobileContact.textContent || '', /QQ 1549277597/)
    assert.match(mobileContact.textContent || '', /微信 ModelPass/)
    assert.match(mobileContact.textContent || '', /1549277597@qq\.com/)
    assert.equal(mobileContact.classList.contains('flex-wrap'), true)
    assert.equal(mobileContact.classList.contains('lg:hidden'), true)

    await rendered.cleanup()
  })
})

after(() => {
  domWindow.close()
})
