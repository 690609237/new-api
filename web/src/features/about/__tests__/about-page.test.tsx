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
const { About } = await import('../index')

const i18n = createInstance()
await i18n.use(initReactI18next).init({
  lng: 'en',
  resources: { en: { translation: {} } },
})

function renderAboutPage() {
  render(
    <I18nextProvider i18n={i18n}>
      <About />
    </I18nextProvider>
  )
}

describe('About page', () => {
  test('shows the author contact heading and all four contact methods', () => {
    renderAboutPage()

    expect(
      screen.getByRole('heading', { level: 1, name: 'Get in touch' })
    ).toBeInTheDocument()

    const contactList = screen.getByRole('list')
    expect(within(contactList).getAllByRole('listitem')).toHaveLength(4)

    for (const value of [
      '1549277597',
      'ModelPass',
      '1549277597@qq.com',
      '450997742',
    ]) {
      expect(within(contactList).getByText(value)).toBeInTheDocument()
    }
  })

  test('provides a direct email link', () => {
    renderAboutPage()

    expect(
      screen.getByRole('link', { name: '1549277597@qq.com' })
    ).toHaveAttribute('href', 'mailto:1549277597@qq.com')
  })

  test('marks the QQ group as recommended and explains its activities', () => {
    renderAboutPage()

    expect(screen.getByText('Recommended')).toBeInTheDocument()
    expect(
      screen.getByText(
        'Occasional AI application exchange activities are held here.'
      )
    ).toBeInTheDocument()
  })

  test('does not show project information on the contact page', () => {
    renderAboutPage()

    expect(
      screen.queryByRole('heading', { name: 'Project information' })
    ).not.toBeInTheDocument()
  })
})
