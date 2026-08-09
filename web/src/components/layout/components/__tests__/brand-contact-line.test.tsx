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
import { describe, test } from 'node:test'

const { renderToStaticMarkup } = await import('react-dom/server')
const { createInstance } = await import('i18next')
const { I18nextProvider, initReactI18next } = await import('react-i18next')
const { BrandContactLine } = await import('../brand-contact-line')

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

describe('brand contact line', () => {
  test('shows all contact details while keeping the email as plain text', () => {
    const markup = renderToStaticMarkup(
      <I18nextProvider i18n={i18n}>
        <BrandContactLine />
      </I18nextProvider>
    )

    assert.equal(markup.includes('联系作者交流'), true)
    assert.equal(markup.includes('QQ 1549277597'), true)
    assert.equal(markup.includes('微信 ModelPass'), true)
    assert.equal(markup.includes('1549277597@qq.com'), true)
    assert.equal(markup.includes('<a'), false)
  })
})
