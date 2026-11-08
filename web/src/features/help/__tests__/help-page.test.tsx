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
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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
  test('opens the first document by default', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <Help />
      </I18nextProvider>
    )

    expect(
      screen.getByRole('tab', { name: 'Codex setup guide' })
    ).toHaveAttribute('aria-selected', 'true')
  })

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

  test('switches to the command-line API document', async () => {
    const user = userEvent.setup()

    render(
      <I18nextProvider i18n={i18n}>
        <Help />
      </I18nextProvider>
    )

    const cliTab = screen.getByRole('tab', {
      name: '命令行调用中转站 API',
    })
    expect(cliTab).toHaveAttribute('aria-selected', 'false')

    await user.click(cliTab)

    expect(cliTab).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByTitle('命令行调用中转站 API 教程')).toHaveAttribute(
      'src',
      '/help/api-cli-tutorial.html#embedded'
    )
  })

  test('stacks the command-line document hero and removes shortcut buttons', () => {
    const source = readFileSync(
      resolve(
        import.meta.dirname,
        '../../../../public/help/api-cli-tutorial.html'
      ),
      'utf8'
    )
    const document = new DOMParser().parseFromString(source, 'text/html')
    const hero = document.querySelector('.hero')
    const heroCard = hero?.querySelector('.hero-card')

    expect(hero).not.toBeNull()
    expect(hero?.querySelector('.hero-actions')).toBeNull()
    expect(hero?.querySelector('a[href="#quick-start"]')).toBeNull()
    expect(hero?.querySelector('a[href="#platforms"]')).toBeNull()
    expect(heroCard?.previousElementSibling?.querySelector('h1')).not.toBeNull()
    expect(
      heroCard?.previousElementSibling?.querySelector('.hero-lead')
    ).not.toBeNull()
    expect(source).toMatch(
      /\.hero \{[^}]*display: flex;[^}]*flex-direction: column;/
    )
  })

  test('keeps the document collection below the fixed public header', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <Help />
      </I18nextProvider>
    )

    expect(screen.getByRole('main')).toHaveClass('pt-20')
    expect(
      screen
        .getByRole('tablist', { name: '文档类型' })
        .closest('[data-slot="tabs"]')
    ).toHaveClass('max-w-7xl')
  })
})
