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
import { CherryStudio } from '@lobehub/icons'
import { Link } from '@tanstack/react-router'
import { ArrowRight, BookOpen } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { RichContent } from '@/components/rich-content'
import { Button } from '@/components/ui/button'
import { isLikelyHtml } from '@/lib/content-format'

import { HeroTerminalDemo } from '../hero-terminal-demo'

interface HeroProps {
  className?: string
  homePageContent?: string
  isAuthenticated?: boolean
}

// Stylized three-dots indicator representing "More"
const MoreIcon = () => (
  <svg
    className='text-muted-foreground/60 group-hover:text-foreground size-6 shrink-0 transition-colors'
    viewBox='0 0 24 24'
    fill='none'
    xmlns='http://www.w3.org/2000/svg'
  >
    <circle cx='6' cy='12' r='2' fill='currentColor' />
    <circle cx='12' cy='12' r='2' fill='currentColor' />
    <circle cx='18' cy='12' r='2' fill='currentColor' />
  </svg>
)

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const homePageContent = props.homePageContent?.trim() ?? ''
  const displayHomePageContent = homePageContent
    .replaceAll(
      'Supports one-click configuration and perfectly adapts to NewAPI multi-protocol configuration.',
      t('Connect popular AI applications with flexible API compatibility.')
    )
    .replaceAll(
      /支持一键配置并完美适配\s*NewAPI\s*多协议配置[。.]?/g,
      t('Connect popular AI applications with flexible API compatibility.')
    )
  const contentIsHtml = isLikelyHtml(displayHomePageContent)

  const renderDocsButton = () => {
    return (
      <Button
        variant='outline'
        className='group border-border/50 hover:border-border hover:bg-muted/50 inline-flex h-11 items-center gap-1.5 rounded-lg px-5 text-sm font-medium'
        render={<Link to='/docs' />}
      >
        <BookOpen className='text-muted-foreground/80 group-hover:text-foreground size-4 transition-colors duration-200' />
        <span>{t('Docs')}</span>
      </Button>
    )
  }

  return (
    <section className='relative z-10 overflow-hidden px-6 pt-16 pb-10 md:pt-20 md:pb-14 lg:pt-24 lg:pb-16'>
      {/* Radial gradient background */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 opacity-25 dark:opacity-[0.12]'
        style={{
          background: [
            'radial-gradient(ellipse 60% 50% at 20% 20%, oklch(0.72 0.18 250 / 80%) 0%, transparent 70%)',
            'radial-gradient(ellipse 50% 40% at 80% 15%, oklch(0.65 0.15 200 / 60%) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 35% at 40% 80%, oklch(0.70 0.12 280 / 40%) 0%, transparent 70%)',
          ].join(', '),
        }}
      />
      {/* Grid pattern */}
      <div
        aria-hidden
        className='absolute inset-0 -z-10 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_30%,black_20%,transparent_100%)] bg-[size:4rem_4rem] opacity-[0.08]'
      />

      <div className='mx-auto max-w-6xl'>
        <div className='landing-animate-fade-up text-center opacity-0'>
          <h1 className='from-foreground via-primary to-foreground/55 mx-auto max-w-5xl bg-gradient-to-r bg-clip-text text-3xl leading-tight font-bold tracking-tight text-balance text-transparent sm:text-4xl lg:text-5xl'>
            {t('ModelPass unified AI API gateway')}
          </h1>
          <div
            role='region'
            aria-label={t('ModelPass platform overview')}
            className='mx-auto mt-7 grid max-w-5xl gap-5 text-left md:grid-cols-3 md:gap-0'
          >
            <article className='md:not-first:border-border/40 relative px-1 py-1 md:px-6 md:not-first:border-l md:first:border-l-0'>
              <h2 className='text-foreground/80 flex items-center text-sm font-semibold tracking-wide'>
                <span className='mr-2 size-1.5 rounded-full bg-sky-400' />
                {t('First-party ChatGPT models')}
              </h2>
              <p className='text-muted-foreground mt-2 text-sm leading-6'>
                {t(
                  'Self-operated ChatGPT models with a stable, unified access point.'
                )}
              </p>
            </article>
            <article className='md:not-first:border-border/40 relative px-1 py-1 md:px-6 md:not-first:border-l md:first:border-l-0'>
              <h2 className='text-foreground/80 flex items-center text-sm font-semibold tracking-wide'>
                <span className='mr-2 size-1.5 rounded-full bg-violet-400' />
                {t('Channel and domestic models')}
              </h2>
              <p className='text-muted-foreground mt-2 text-sm leading-6'>
                {t(
                  'ChatGPT and Claude through partners, plus Kimi, Qwen, and DeepSeek.'
                )}
              </p>
            </article>
            <article className='md:not-first:border-border/40 relative px-1 py-1 md:px-6 md:not-first:border-l md:first:border-l-0'>
              <h2 className='text-foreground/80 flex items-center text-sm font-semibold tracking-wide'>
                <span className='mr-2 size-1.5 rounded-full bg-amber-400' />
                {t('Safety and compliance')}
              </h2>
              <p className='text-muted-foreground mt-2 text-sm leading-6'>
                {t(
                  'Real-time and routine risk controls. Illegal use and bypassing platform safeguards are prohibited.'
                )}
              </p>
            </article>
          </div>
        </div>

        <div className='mt-12 grid grid-cols-1 items-start gap-12 lg:grid-cols-12 lg:gap-8'>
          {/* Left Column: pricing details, action buttons and application support */}
          <div className='flex flex-col items-start text-left lg:col-span-6'>
            <div
              role='region'
              aria-label={t('Platform pricing and model groups')}
              className='landing-animate-fade-up mt-4 flex w-full max-w-xl flex-col gap-3 opacity-0'
              style={{ animationDelay: '60ms' }}
            >
              <RichContent
                mode={contentIsHtml ? 'html' : 'markdown'}
                htmlVariant='isolated'
                content={displayHomePageContent}
                className={
                  contentIsHtml
                    ? 'home-feature-content'
                    : 'bg-card/60 border-border/60 rounded-2xl border p-5 shadow-sm'
                }
              />
            </div>

            <div
              className='landing-animate-fade-up mt-6 flex flex-wrap items-center gap-3 opacity-0'
              style={{ animationDelay: '120ms' }}
            >
              {props.isAuthenticated ? (
                <>
                  <Button
                    className='group h-11 rounded-lg px-5 text-sm font-medium'
                    render={<Link to='/dashboard' />}
                  >
                    {t('Go to Dashboard')}
                    <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                  </Button>
                  {renderDocsButton()}
                </>
              ) : (
                <>
                  <Button
                    className='group h-11 rounded-lg px-5 text-sm font-medium'
                    render={<Link to='/sign-up' />}
                  >
                    {t('Get Started')}
                    <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                  </Button>
                  <Button
                    variant='outline'
                    className='border-border/50 hover:border-border hover:bg-muted/50 h-11 rounded-lg px-5 text-sm font-medium'
                    render={<Link to='/pricing' />}
                  >
                    {t('View Pricing')}
                  </Button>
                  {renderDocsButton()}
                </>
              )}
            </div>

            {/* Supported Apps (参考图二样式，进行卡片化和信息扩充设计，增加视觉高度) */}
            <div
              className='landing-animate-fade-up mt-10 w-full max-w-xl opacity-0'
              style={{ animationDelay: '240ms' }}
            >
              <div className='mb-4 flex flex-col gap-1'>
                <span className='text-muted-foreground/50 text-[10px] font-bold tracking-[0.15em] uppercase'>
                  {t('Supported Applications')}
                </span>
                <p className='text-muted-foreground/60 text-xs leading-relaxed'>
                  {t(
                    'Connect popular AI applications with flexible API compatibility.'
                  )}
                </p>
              </div>
              <div className='flex flex-wrap items-center gap-3'>
                {/* Cherry Studio */}
                <a
                  href='https://cherry-ai.com'
                  target='_blank'
                  rel='noopener noreferrer'
                  className='group border-border/40 bg-muted/15 text-foreground/80 hover:border-border hover:bg-muted/30 hover:text-foreground flex items-center gap-3 rounded-full border px-5 py-2.5 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]'
                >
                  <CherryStudio.Color size={24} className='shrink-0' />
                  <span>Cherry Studio</span>
                </a>

                {/* CC Switch */}
                <a
                  href='https://ccswitch.io'
                  target='_blank'
                  rel='noopener noreferrer'
                  className='group border-border/40 bg-muted/15 text-foreground/80 hover:border-border hover:bg-muted/30 hover:text-foreground flex items-center gap-3 rounded-full border px-5 py-2.5 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]'
                >
                  <img
                    src='https://ccswitch.io/favicon.png'
                    alt='CC Switch'
                    className='size-6 shrink-0 rounded-md object-contain'
                    onError={(e) => {
                      // Fallback to a styled text avatar if the remote favicon fails to load in sandbox or local environments
                      e.currentTarget.style.display = 'none'
                      const fallback = e.currentTarget
                        .nextSibling as HTMLElement
                      if (fallback) fallback.style.display = 'flex'
                    }}
                  />
                  <span
                    style={{ display: 'none' }}
                    className='size-6 shrink-0 items-center justify-center rounded-md bg-blue-500/10 text-[10px] font-bold text-blue-600 dark:bg-blue-400/10 dark:text-blue-400'
                  >
                    CC
                  </span>
                  <span>CC Switch</span>
                </a>

                {/* "更多" */}
                <div className='group border-border/40 bg-muted/15 text-foreground/55 hover:border-border hover:bg-muted/30 hover:text-foreground flex cursor-default items-center gap-2.5 rounded-full border px-5 py-2.5 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]'>
                  <MoreIcon />
                  <span>{t('More Apps')}</span>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column: Hero Terminal API Demo */}
          <div
            className='landing-animate-fade-up flex w-full justify-center opacity-0 lg:col-span-6'
            style={{ animationDelay: '320ms' }}
          >
            <HeroTerminalDemo className='mt-8 lg:mt-0' />
          </div>
        </div>
      </div>
    </section>
  )
}
