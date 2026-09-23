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
import { render, screen } from '@testing-library/react'
import type { TFunction } from 'i18next'
import React from 'react'
import { describe, expect, test, vi } from 'vitest'

import { getDefaultHomePageContent } from '@/features/home/default-home-page-content'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))

vi.mock('@lobehub/icons', () => ({
  CherryStudio: { Color: () => null },
}))

vi.mock('@tanstack/react-router', () => ({
  Link: (props: { children?: React.ReactNode; to: string }) => (
    <a href={props.to}>{props.children}</a>
  ),
}))

vi.mock('@/components/ui/button', () => ({
  Button: (props: {
    children?: React.ReactNode
    render?: React.ReactElement<{ children?: React.ReactNode }>
  }) => {
    if (props.render) {
      return React.cloneElement(props.render, undefined, props.children)
    }

    return <button type='button'>{props.children}</button>
  },
}))

vi.mock('@/features/home/components/hero-terminal-demo', () => ({
  HeroTerminalDemo: () => null,
}))

const { Hero } = await import('../hero')

describe('Hero documentation link', () => {
  test('localizes and escapes the default information content', () => {
    const translations: Record<string, string> = {
      'Platform pricing and channel source summary': '摘要 <区域>',
      '1:1 platform recharge': '充值 & 计费',
      'Model prices use the $ symbol. Overseas models follow international pricing, while domestic models follow domestic pricing.':
        '价格使用 "美元"',
      'First-party token groups': "一手 token '分组'",
    }
    const t = ((key: string) => translations[key] ?? key) as TFunction

    const content = getDefaultHomePageContent(t)

    expect(content).toContain('aria-label="摘要 &lt;区域&gt;"')
    expect(content).toContain('充值 &amp; 计费')
    expect(content).toContain('价格使用 &quot;美元&quot;')
    expect(content).toContain('一手 token &#39;分组&#39;')
    expect(content).not.toContain('摘要 <区域>')
  })

  test('opens the local help page', () => {
    render(<Hero />)

    expect(screen.getByRole('link', { name: 'Docs' })).toHaveAttribute(
      'href',
      '/docs'
    )
  })

  test('shows the responsible-use notice above the hero content', () => {
    const { container } = render(<Hero />)

    const notice = screen.getByRole('note')
    const informationArea = screen.getByRole('region', {
      name: 'Platform pricing and channel source summary',
    })

    expect(notice).toHaveTextContent(
      'Use responsibly; breaking limits is strictly prohibited!'
    )
    expect(
      notice.compareDocumentPosition(informationArea) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
    expect(notice.parentElement).toContainElement(
      screen.getByRole('link', { name: 'Docs' })
    )
    expect(container.querySelector('section')).toHaveClass(
      'pt-16',
      'pb-10',
      'md:pt-20',
      'md:pb-14',
      'lg:pt-24',
      'lg:pb-16'
    )
  })

  test('explains platform pricing and channel sources', () => {
    render(<Hero />)

    const summary = screen.getByRole('region', {
      name: 'Platform pricing and channel source summary',
    })
    const contentHost = summary.querySelector<HTMLElement>(
      '.home-feature-content'
    )
    const contentRoot = contentHost?.shadowRoot

    expect(contentRoot).not.toBeNull()
    expect(contentRoot?.querySelectorAll('[role="listitem"]')).toHaveLength(2)
    expect(contentRoot?.textContent).toContain('1:1 platform recharge')
    expect(contentRoot?.textContent).toContain('$ USD')
    expect(contentRoot?.textContent).toContain('First-party token groups')
    expect(contentRoot?.textContent).toContain('default')
    expect(contentRoot?.textContent).toContain('standard')
  })

  test('sanitizes custom HTML inside the information area only', () => {
    render(
      <Hero homePageContent='<div data-testid="custom-card">自定义卡片</div><script>window.bad = true</script>' />
    )

    const summary = screen.getByRole('region', {
      name: 'Platform pricing and channel source summary',
    })
    const contentRoot = summary.querySelector<HTMLElement>(
      '.home-feature-content'
    )?.shadowRoot

    expect(contentRoot?.textContent).toContain('自定义卡片')
    expect(contentRoot?.querySelector('script')).toBeNull()
    expect(screen.getByRole('note')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Docs' })).toBeInTheDocument()
  })
})
