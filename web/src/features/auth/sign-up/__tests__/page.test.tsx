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
import { render, screen } from '@testing-library/react'
import type React from 'react'
import { describe, expect, test, vi } from 'vitest'

vi.mock('@tanstack/react-router', () => ({
  Link: (props: { children: React.ReactNode; to: string }) => (
    <a href={props.to}>{props.children}</a>
  ),
}))
vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({ status: null }),
}))
vi.mock('../../auth-layout', () => ({
  AuthLayout: (props: { children: React.ReactNode }) => props.children,
}))
vi.mock('../components/sign-up-form', () => ({
  SignUpForm: () => <form aria-label='sign-up form' />,
}))
vi.mock('../components/terms-footer', () => ({
  TermsFooter: () => null,
}))

const { createInstance } = await import('i18next')
const { I18nextProvider, initReactI18next } = await import('react-i18next')
const { SignUp } = await import('../index')

const i18n = createInstance()
await i18n.use(initReactI18next).init({
  lng: 'zh',
  resources: {
    zh: {
      translation: {
        'Create an account': '创建一个账户',
        'Join group 450997742 to claim free credits': '加群450997742领免费额度',
        'Already have an account?': '已有账户？',
        'Sign in': '登录',
      },
    },
  },
})

describe('sign-up page', () => {
  test('shows the free-credit group reminder below the account heading', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <SignUp />
      </I18nextProvider>
    )

    const heading = screen.getByRole('heading', { name: '创建一个账户' })
    const reminder = screen.getByRole('note')

    expect(reminder).toHaveTextContent('加群450997742领免费额度')
    expect(heading.compareDocumentPosition(reminder)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING
    )
  })
})
