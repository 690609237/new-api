/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import assert from 'node:assert/strict'

import { Window } from 'happy-dom'

import type { PlanRecord } from '../../../types'

const bunTestModule = 'bun:test'
const { afterEach, test } = (await import(bunTestModule)) as {
  afterEach: typeof import('node:test').afterEach
  test: typeof import('node:test').test
}

const domWindow = new Window()
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
  'KeyboardEvent',
  'PointerEvent',
  'MouseEvent',
  'FocusEvent',
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

const React = await import('react')
const { act } = React
const { createRoot } = await import('react-dom/client')
const { createInstance } = await import('i18next')
const { I18nextProvider, initReactI18next } = await import('react-i18next')
const { QueryClient, QueryClientProvider } =
  await import('@tanstack/react-query')
const { SubscriptionsProvider, useSubscriptions } =
  await import('../../subscriptions-provider')
const { ToggleStatusDialog } = await import('../toggle-status-dialog')

const warning =
  'Disabling this plan immediately terminates active subscriptions and moves related API keys to an available fallback group. Re-enabling it will not restore terminated subscriptions. Continue?'

const i18n = createInstance()
await i18n.use(initReactI18next).init({
  lng: 'en',
  resources: {
    en: {
      translation: {
        Cancel: 'Cancel',
        'Confirm disable': 'Confirm disable',
        Disable: 'Disable',
        [warning]: warning,
      },
    },
  },
})

const reactTestGlobals = globalThis as typeof globalThis & {
  IS_REACT_ACT_ENVIRONMENT?: boolean
}
reactTestGlobals.IS_REACT_ACT_ENVIRONMENT = true

const enabledPlan: PlanRecord = {
  plan: {
    id: 1,
    title: 'Monthly plan',
    price_amount: 10,
    currency: 'USD',
    duration_unit: 'month',
    duration_value: 1,
    quota_reset_period: 'never',
    enabled: true,
    sort_order: 0,
    allow_balance_pay: true,
    allow_wallet_overflow: true,
    max_purchase_per_user: 0,
    total_amount: 100,
  },
}

let root: ReturnType<typeof createRoot> | null = null

function OpenDisabledPlanDialog() {
  const subscriptions = useSubscriptions()

  React.useEffect(() => {
    subscriptions.setCurrentRow(enabledPlan)
    subscriptions.setOpen('toggle-status')
  }, [subscriptions])

  return <ToggleStatusDialog />
}

afterEach(async () => {
  if (root) {
    await act(async () => root?.unmount())
    root = null
  }
  document.body.replaceChildren()
})

test('disabling an enabled plan warns that subscriptions are terminated irreversibly', async () => {
  const host = document.createElement('div')
  document.body.append(host)
  root = createRoot(host)
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  queryClient.setQueryData(
    ['system-options'],
    { data: [] },
    { updatedAt: Date.now() + 60_000 }
  )

  await act(async () => {
    root?.render(
      <I18nextProvider i18n={i18n}>
        <QueryClientProvider client={queryClient}>
          <SubscriptionsProvider>
            <OpenDisabledPlanDialog />
          </SubscriptionsProvider>
        </QueryClientProvider>
      </I18nextProvider>
    )
  })

  assert.match(document.body.textContent ?? '', new RegExp(warning))
})
