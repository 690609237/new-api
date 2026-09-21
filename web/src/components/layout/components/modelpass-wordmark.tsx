/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import { cn } from '@/lib/utils'

interface ModelPassWordmarkProps {
  name: string
  className?: string
}

/** Applies the ModelPass palette to the exact brand name. */
export function ModelPassWordmark(props: ModelPassWordmarkProps) {
  const isModelPass = props.name.trim().toLowerCase() === 'modelpass'

  if (!isModelPass) {
    return <span>{props.name}</span>
  }

  return (
    <span
      className={cn(
        'whitespace-nowrap font-[820] tracking-[-0.055em]',
        props.className
      )}
      aria-label='ModelPass'
    >
      <span className='text-[#008b8c]'>Model</span>
      <span className='text-[#e95d21]'>Pass</span>
    </span>
  )
}
