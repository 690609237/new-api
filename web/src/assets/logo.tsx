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
import type { SVGProps } from 'react'

import { cn } from '@/lib/utils'

export function Logo({ className, ...props }: SVGProps<SVGSVGElement>) {
  return (
    <svg
      id='modelpass-logo'
      viewBox='0 0 128 128'
      xmlns='http://www.w3.org/2000/svg'
      height='128'
      width='128'
      className={cn('size-6', className)}
      {...props}
    >
      <title>ModelPass</title>
      <defs>
        <linearGradient
          id='modelpass-logo-gradient'
          x1='0'
          y1='0'
          x2='1'
          y2='0'
        >
          <stop offset='0' stopColor='#008b8c' />
          <stop offset='0.5' stopColor='#008b8c' />
          <stop offset='0.5' stopColor='#e95d21' />
          <stop offset='1' stopColor='#e95d21' />
        </linearGradient>
      </defs>
      <rect
        x='1'
        y='1'
        width='126'
        height='126'
        rx='29'
        fill='url(#modelpass-logo-gradient)'
        stroke='#fff'
        strokeOpacity='.76'
        strokeWidth='2'
      />
      <text
        x='18'
        y='83'
        fill='#fff'
        fontFamily='Inter, PingFang SC, Microsoft YaHei, system-ui, sans-serif'
        fontSize='50'
        fontWeight='900'
        letterSpacing='-3'
      >
        M
      </text>
      <text
        x='66'
        y='83'
        fill='#fff'
        fontFamily='Inter, PingFang SC, Microsoft YaHei, system-ui, sans-serif'
        fontSize='50'
        fontWeight='900'
        letterSpacing='-3'
      >
        P
      </text>
    </svg>
  )
}
