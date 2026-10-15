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
import { renderToStaticMarkup } from 'react-dom/server'
import { afterEach, expect, it } from 'vitest'

import { HtmlContent } from '@/components/html-content'
import { HtmlContentPrerenderContext } from '@/components/html-content-prerender-context'

afterEach(() => {
  document.body.replaceChildren()
  document.querySelector('[data-test-stylesheet]')?.remove()
})

it('includes sanitized isolated content and its stylesheet in the initial HTML without running effects', () => {
  const style = document.createElement('link')
  style.rel = 'stylesheet'
  style.href = '/static/css/app.hash.css'
  style.dataset.testStylesheet = ''
  document.head.append(style)

  const markup = renderToStaticMarkup(
    <HtmlContentPrerenderContext value>
      <HtmlContent
        variant='isolated'
        content='<style>.offer{color:red}</style><h2 class="offer">Build-time offer</h2><p><span style="color: #ef4444; font-weight: 700">5折</span></p><script>bad()</script><a href="/docs" target="_blank" onclick="bad()">Docs</a><iframe src="https://example.com" srcdoc="<script>bad()</script>"></iframe>'
      />
    </HtmlContentPrerenderContext>
  )
  document.body.innerHTML = markup
  const template = document.querySelector<HTMLTemplateElement>(
    'template[shadowrootmode="open"]'
  )

  expect(template).not.toBeNull()
  if (!template) {
    throw new Error('prerendered shadow root template is missing')
  }
  const content = template.content
  expect(content.querySelector('h2')?.textContent).toBe('Build-time offer')
  expect(content.querySelector('span')?.getAttribute('style')).toContain(
    'color: #ef4444'
  )
  expect(content.querySelector('link')?.getAttribute('href')).toBe(
    '/static/css/app.hash.css'
  )
  const isolatedStyles = content.querySelector('style')?.textContent ?? ''
  expect(isolatedStyles).toContain(':host')
  expect(isolatedStyles).not.toContain('color: inherit !important')
  expect(content.querySelector('script')).toBeNull()
  expect(content.querySelector('[onclick]')).toBeNull()
  expect(content.querySelector('a')?.getAttribute('rel')).toBe(
    'noopener noreferrer'
  )
  expect(content.querySelector('iframe')?.getAttribute('srcdoc')).toBeNull()
  expect(
    content.querySelector('iframe')?.getAttribute('sandbox')
  ).not.toContain('allow-scripts')
})
