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
import { render, screen, within } from '@testing-library/react'
import type React from 'react'
import { describe, expect, test, vi } from 'vitest'

vi.mock('@/components/layout', () => ({
  PublicLayout: (props: { children: React.ReactNode }) => props.children,
}))

const { createInstance } = await import('i18next')
const { I18nextProvider, initReactI18next } = await import('react-i18next')
const { Help } = await import('../index')

const i18n = createInstance()
await i18n.use(initReactI18next).init({
  lng: 'en',
  resources: { en: { translation: {} } },
})

describe('CC Switch help page', () => {
  test('explains the configuration principle before presenting CC Switch as an example', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <Help />
      </I18nextProvider>
    )

    expect(
      screen.getByRole('heading', {
        level: 1,
        name: 'Configure Codex to connect to this service',
      })
    ).toBeInTheDocument()
    expect(screen.getByRole('note')).toHaveTextContent(
      'CC Switch is one example'
    )
    expect(
      screen.getByRole('heading', {
        level: 2,
        name: 'Example: Connect Codex with CC Switch',
      })
    ).toBeInTheDocument()
  })

  test('shows all six ordered setup steps and only four screenshot panels', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <Help />
      </I18nextProvider>
    )

    const steps = within(screen.getByRole('list')).getAllByRole('listitem')
    expect(steps).toHaveLength(6)
    expect(within(steps[0]).getByRole('heading')).toHaveTextContent(
      'Download CC Switch'
    )
    expect(within(steps[5]).getByRole('heading')).toHaveTextContent(
      'Restart ChatGPT'
    )
    expect(screen.getAllByRole('img')).toHaveLength(4)

    const firstStepLayout = steps[0].querySelector(
      '[data-slot="card-content"] > div'
    )
    const illustratedStepLayout = steps[1].querySelector(
      '[data-slot="card-content"] > div'
    )
    const lastStepLayout = steps[5].querySelector(
      '[data-slot="card-content"] > div'
    )
    expect(firstStepLayout).not.toHaveClass('grid')
    expect(illustratedStepLayout).toHaveClass('grid')
    expect(lastStepLayout).not.toHaveClass('grid')
  })

  test('opens the official CC Switch download page in a separate tab', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <Help />
      </I18nextProvider>
    )

    const downloadButton = screen.getByRole('button', {
      name: /Download CC Switch/,
    })
    expect(downloadButton).toHaveAttribute('href', 'https://ccswitch.io/zh/')
    expect(downloadButton).toHaveAttribute('target', '_blank')
    expect(downloadButton).toHaveAttribute('rel', 'noopener noreferrer')
  })
})
