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
import { useSystemConfigStore } from '@/stores/system-config-store'

interface RenderEventRouter {
  subscribe: (eventType: 'onRendered', listener: () => void) => () => void
  state: {
    location: { pathname: string }
    matches: ReadonlyArray<{ status: string; globalNotFound?: boolean }>
  }
}

/**
 * Keep the server-rendered page until TanStack Router has committed a route.
 * This covers setup/auth guards and lazy route chunks without showing a blank
 * document on slow connections.
 */
export function installPrerenderHandoff(router: RenderEventRouter): () => void {
  const bootstrap = document.querySelector('#home-static-bootstrap')
  if (document.querySelector('#home-prerender') && bootstrap) {
    try {
      const snapshot: unknown = JSON.parse(bootstrap.textContent || '{}')
      if (
        snapshot &&
        typeof snapshot === 'object' &&
        'version' in snapshot &&
        snapshot.version === 1 &&
        'systemConfig' in snapshot
      ) {
        const config = snapshot.systemConfig
        if (
          config &&
          typeof config === 'object' &&
          'systemName' in config &&
          'logo' in config &&
          typeof config.systemName === 'string' &&
          typeof config.logo === 'string'
        ) {
          // Only seed presentation. Live status and authentication guards must
          // still fetch their current flags from the server.
          const store = useSystemConfigStore.getState()
          store.setConfig({
            systemName: config.systemName,
            logo: config.logo,
            footerHtml:
              'footerHtml' in config && typeof config.footerHtml === 'string'
                ? config.footerHtml
                : '',
          })
          store.setLoadedLogoUrl(config.logo)
          store.setLoading(false)
        }
      }
    } catch {
      // An unresolved template token in development has no build snapshot.
    }
  }

  return router.subscribe('onRendered', () => {
    if (typeof document === 'undefined') return
    const home = document.querySelector('#home-prerender')
    const path = router.state.location.pathname
    if (home && (path === '/' || path === '/index.html')) {
      // onRendered also fires for error boundaries. A failed guard/chunk must
      // not replace an already readable home with a client error page.
      if (
        router.state.matches.some(
          (match) => match.status !== 'success' || match.globalNotFound
        ) ||
        !document.querySelector('#root [data-page="home"]')
      ) {
        return
      }
      document.querySelector('#root')?.setAttribute('data-home-ready', '')
    }
    document.querySelector<HTMLElement>('#seo-content')?.remove()
  })
}
