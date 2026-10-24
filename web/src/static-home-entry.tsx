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
*/

/**
 * Build-only entry point for the public landing page.
 *
 * The build runner installs a jsdom window before importing this module. The
 * module then renders the real Home component with an in-memory router and a
 * fixed anonymous configuration; no root route guards or network requests
 * are involved.
 */

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRoute,
  createRootRoute,
  createRouter,
  RouterContextProvider,
} from '@tanstack/react-router'
import { renderToStaticMarkup } from 'react-dom/server'
import { z } from 'zod'

import { HtmlContentPrerenderContext } from '@/components/html-content-prerender-context'
import { DirectionProvider } from '@/context/direction-provider'
import { FontProvider } from '@/context/font-provider'
import { ThemeCustomizationProvider } from '@/context/theme-customization-provider'
import { ThemeProvider } from '@/context/theme-provider'
import { Home } from '@/features/home'
import i18n from '@/i18n/config'
import { INTERFACE_LANGUAGE_OPTIONS } from '@/i18n/languages'
import {
  mapStatusDataToConfig,
  STATUS_QUERY_KEY,
  STATUS_STORAGE_KEY,
} from '@/lib/status-query'
import { useAuthStore } from '@/stores/auth-store'
import { useSystemConfigStore } from '@/stores/system-config-store'

const staticHomeConfigSchema = z.strictObject({
  locale: z
    .string()
    .refine(
      (value) =>
        INTERFACE_LANGUAGE_OPTIONS.some((language) => language.code === value),
      'unsupported locale'
    )
    .optional(),
  homePageContent: z.string().optional(),
  status: z.record(z.string(), z.unknown()).optional(),
  systemConfig: z
    .strictObject({
      systemName: z.string().optional(),
      logo: z.string().optional(),
      footerHtml: z.string().optional(),
      demoSiteEnabled: z.boolean().optional(),
      displayTokenStatEnabled: z.boolean().optional(),
    })
    .optional(),
})

export type StaticHomeConfig = z.infer<typeof staticHomeConfigSchema>

const DEFAULT_STATUS: Record<string, unknown> = {
  system_name: 'ModelPass',
  logo: '/logo.svg?v=3',
  HeaderNavModules: JSON.stringify({
    home: true,
    console: true,
    pricing: { enabled: true, requireAuth: false },
    rankings: { enabled: true, requireAuth: false },
    docs: true,
    about: true,
  }),
  announcements_enabled: false,
  user_agreement_enabled: false,
  privacy_policy_enabled: false,
}

function createStaticRouter() {
  const rootRoute = createRootRoute({ component: () => null })
  const emptyRoute = (path: string) =>
    createRoute({
      getParentRoute: () => rootRoute,
      path,
      component: () => null,
    })
  const homeRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: Home,
  })

  // Link only needs target paths to exist in the route tree. These routes are
  // never rendered by the build entry.
  const routeTree = rootRoute.addChildren([
    homeRoute,
    emptyRoute('/about'),
    emptyRoute('/dashboard'),
    emptyRoute('/docs'),
    emptyRoute('/pricing'),
    emptyRoute('/rankings'),
    emptyRoute('/sign-in'),
    emptyRoute('/sign-up'),
    emptyRoute('/privacy-policy'),
    emptyRoute('/user-agreement'),
  ])

  return createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
}

export async function renderHome(input: unknown = {}): Promise<{
  fragment: string
  bootstrap: Record<string, unknown>
}> {
  const config = staticHomeConfigSchema.parse(input)
  const homeBootstrap = document.querySelector('#home-page-bootstrap')
  if (homeBootstrap) {
    homeBootstrap.textContent = JSON.stringify(config.homePageContent ?? '')
  }
  const status = { ...DEFAULT_STATUS, ...config.status }
  const systemConfig = {
    ...mapStatusDataToConfig(status),
    ...config.systemConfig,
  }
  const bootstrap = {
    version: 1,
    locale: config.locale ?? 'zhCN',
    homePageContent: config.homePageContent ?? '',
    status,
    systemConfig,
  }

  await i18n.changeLanguage(config.locale ?? 'zhCN')
  useAuthStore.getState().auth.reset('complete')
  await useSystemConfigStore.persist.rehydrate()
  useSystemConfigStore.setState((state) => ({
    ...state,
    config: { ...state.config, ...systemConfig },
    loading: false,
    loadedLogoUrl: systemConfig.logo ?? '/logo.svg?v=3',
  }))
  // Zustand's server snapshot reader retains the original state object. Pin
  // that build-only snapshot to the public configuration as well as the live
  // state so the static header and footer do not render their loading UI.
  Object.assign(
    useSystemConfigStore.getInitialState(),
    useSystemConfigStore.getState()
  )

  window.localStorage.setItem(STATUS_STORAGE_KEY, JSON.stringify(status))
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  queryClient.setQueryData(STATUS_QUERY_KEY, status)
  queryClient.setQueryData(['notice'], { success: true, data: '' })
  queryClient.setQueryData(['status-content', 'announcements'], [])

  const router = createStaticRouter()
  router.update({ context: { queryClient } })
  await router.load()

  const app = (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider defaultTheme='light'>
        <ThemeCustomizationProvider>
          <FontProvider>
            <DirectionProvider>
              <HtmlContentPrerenderContext value>
                <RouterContextProvider router={router}>
                  <Home />
                </RouterContextProvider>
              </HtmlContentPrerenderContext>
            </DirectionProvider>
          </FontProvider>
        </ThemeCustomizationProvider>
      </ThemeProvider>
    </QueryClientProvider>
  )

  const fragment = renderToStaticMarkup(app)
  queryClient.clear()
  return { fragment, bootstrap }
}
