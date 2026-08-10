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
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

type BrandContactLineProps = {
  className?: string
  variant?: 'inline' | 'compact'
}

export function BrandContactLine(props: BrandContactLineProps) {
  const { t } = useTranslation()
  const variant = props.variant ?? 'inline'
  const compact = variant === 'compact'

  return (
    <span
      data-slot='brand-contact-line'
      data-variant={variant}
      className={cn(
        'text-muted-foreground text-[10px] font-normal tracking-normal',
        compact
          ? 'flex max-w-full flex-wrap items-center gap-x-1 gap-y-0.5 leading-[1.15]'
          : 'block leading-none whitespace-nowrap',
        props.className
      )}
    >
      <span className='whitespace-nowrap'>
        {t('Contact the author')}：QQ 1549277597
      </span>
      <span className='whitespace-nowrap'>
        <span className={cn('text-border', !compact && 'ms-1')}>·</span>{' '}
        {t('WeChat')} ModelPass
      </span>
      <span className='whitespace-nowrap'>
        <span className={cn('text-border', !compact && 'ms-1')}>·</span>{' '}
        1549277597@qq.com
      </span>
    </span>
  )
}
