import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { EmptyState } from '@/components/empty-state'
import { ErrorState } from '@/components/error-state'
import { Markdown } from '@/components/ui/markdown'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'

import { getDailyReviewReport } from './api'
import { highlightDailyReviewRisks } from './daily-review-risk'

type DailyReviewDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function DailyReviewDialog(props: DailyReviewDialogProps) {
  const { t } = useTranslation()
  const [selectedDate, setSelectedDate] = useState('')
  const report = useQuery({
    queryKey: ['moderation-daily-review', selectedDate],
    queryFn: () => getDailyReviewReport(selectedDate || undefined),
    enabled: props.open,
    staleTime: 0,
  })
  const result = report.data
  const displayedDate = result?.date
    ? `${result.date.slice(0, 4)}-${result.date.slice(4, 6)}-${result.date.slice(6)}`
    : ''

  return (
    <Dialog
      open={props.open}
      onOpenChange={(open) => {
        if (!open) setSelectedDate('')
        props.onOpenChange(open)
      }}
      title={t('Daily review results')}
      description={
        displayedDate ? `${t('Log date')}: ${displayedDate}` : undefined
      }
      contentClassName='sm:max-w-5xl'
      contentHeight='min(70vh, 760px)'
    >
      {result?.available_dates.length ? (
        <div className='mb-4 flex items-center gap-2'>
          <span className='text-sm'>{t('Log date')}</span>
          <Select
            items={result.available_dates.map((day) => ({
              value: day,
              label: day,
            }))}
            value={selectedDate || result.date}
            onValueChange={(value) => setSelectedDate(value ?? '')}
          >
            <SelectTrigger aria-label={t('Log date')}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                {result.available_dates.map((day) => (
                  <SelectItem key={day} value={day}>
                    {day}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
      ) : null}
      {report.isPending ? <Skeleton className='h-48 w-full' /> : null}
      {report.isError ? (
        <ErrorState
          description={t('Failed to load daily review')}
          onRetry={() => void report.refetch()}
        />
      ) : null}
      {result && !result.content ? (
        <EmptyState description={t('No daily review results yet')} />
      ) : null}
      {result?.content ? (
        <Markdown className='[&_mark]:rounded [&_mark]:bg-amber-300 [&_mark]:px-1 dark:[&_mark]:bg-amber-700 [&_tr:has(td:nth-child(3)_mark)]:bg-amber-100 dark:[&_tr:has(td:nth-child(3)_mark)]:bg-amber-950/70'>
          {highlightDailyReviewRisks(result.content)}
        </Markdown>
      ) : null}
    </Dialog>
  )
}
