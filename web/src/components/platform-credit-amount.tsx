import type { ReactNode } from 'react'

import {
  formatPlatformAmount,
  formatQuotaWithCurrency,
  PLATFORM_CREDIT_SYMBOL,
} from '@/lib/currency'
import { cn } from '@/lib/utils'

type PlatformCreditAmountProps = {
  value: number
  className?: string
  rawQuota?: boolean
}

export function PlatformCredit(props: {
  children: ReactNode
  className?: string
}) {
  return (
    <span className={cn('inline-flex items-center gap-1', props.className)}>
      <PlatformCreditIcon />
      <span>{props.children}</span>
    </span>
  )
}

/** Shared platform-credit mark for amounts and quota summary cards. */
export function PlatformCreditIcon(props: { className?: string }) {
  return (
    <span
      className={cn(
        'text-foreground/80 inline-flex size-5 shrink-0 items-center justify-center text-lg leading-none',
        props.className
      )}
      aria-hidden='true'
    >
      ✦
    </span>
  )
}

/** Render an internal platform amount with the shared credits icon. */
export function PlatformCreditAmount(props: PlatformCreditAmountProps) {
  const formatted = (
    props.rawQuota
      ? formatQuotaWithCurrency(props.value, {
          digitsLarge: 4,
          digitsSmall: 6,
          abbreviate: false,
          showSymbol: false,
        })
      : formatPlatformAmount(props.value)
  ).replace(PLATFORM_CREDIT_SYMBOL, '')

  return (
    <PlatformCredit className={props.className}>{formatted}</PlatformCredit>
  )
}
