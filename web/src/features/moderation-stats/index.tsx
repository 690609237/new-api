import { useQuery } from '@tanstack/react-query'
import { BarChart3, RefreshCw, ScanSearch } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { StaticDataTable } from '@/components/data-table'
import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { CompactDateTimeRangePicker } from '@/features/usage-logs/components/compact-date-time-range-picker'
import { dateToUnixTimestamp, getRollingDateRange } from '@/lib/time'

import { getModerationUsageStats, type ModerationUsageDimension } from './api'
import { DailyReviewDialog } from './daily-review-dialog'

type StatsQueryParams = {
  start_timestamp: number
  end_timestamp: number
  user_id?: number
  token_id?: number
}

const numberFormatter = new Intl.NumberFormat()

function StatCard(props: { label: string; value: string | number }) {
  return (
    <Card size='sm'>
      <CardHeader>
        <CardTitle className='text-muted-foreground text-xs font-medium'>
          {props.label}
        </CardTitle>
      </CardHeader>
      <CardContent className='font-mono text-2xl font-semibold tabular-nums'>
        {props.value}
      </CardContent>
    </Card>
  )
}

export function ModerationStats() {
  const { t } = useTranslation()
  const defaultRange = getRollingDateRange(1)
  const [start, setStart] = useState(defaultRange.start)
  const [end, setEnd] = useState(defaultRange.end)
  const [userID, setUserID] = useState('')
  const [tokenID, setTokenID] = useState('')
  const [reviewOpen, setReviewOpen] = useState(false)
  const [queryParams, setQueryParams] = useState<StatsQueryParams>({
    start_timestamp: dateToUnixTimestamp(start),
    end_timestamp: dateToUnixTimestamp(end),
  })
  const query = useQuery({
    queryKey: ['moderation-usage-stats', queryParams],
    queryFn: () => getModerationUsageStats(queryParams),
  })
  const data = query.data
  const dimensions = data?.dimensions ?? []
  const columns = useMemo(
    () => [
      {
        id: 'user',
        header: t('User ID'),
        cell: (row: ModerationUsageDimension) => row.user_id,
      },
      {
        id: 'token',
        header: t('Token ID'),
        cell: (row: ModerationUsageDimension) => row.token_id,
      },
      {
        id: 'requests',
        header: t('API requests'),
        cell: (row: ModerationUsageDimension) => row.api_requests,
      },
      {
        id: 'sensitive-word-hits',
        header: t('Sensitive word hits'),
        cell: (row: ModerationUsageDimension) => row.sensitive_word_hits,
      },
      {
        id: 'passed',
        header: t('Passed'),
        cell: (row: ModerationUsageDimension) => row.api_passed,
      },
      {
        id: 'succeeded',
        header: t('Successful'),
        cell: (row: ModerationUsageDimension) =>
          row.api_passed + row.api_violations,
      },
      {
        id: 'violations',
        header: t('Violations'),
        cell: (row: ModerationUsageDimension) => row.api_violations,
      },
      {
        id: 'failed',
        header: t('Failed'),
        cell: (row: ModerationUsageDimension) => row.api_failed,
      },
      {
        id: 'timeouts',
        header: t('Timeouts'),
        cell: (row: ModerationUsageDimension) => row.api_timeouts,
      },
      {
        id: 'circuit-skips',
        header: t('Circuit skips'),
        cell: (row: ModerationUsageDimension) => row.circuit_skips,
      },
      {
        id: 'cache-hits',
        header: t('Cache hits'),
        cell: (row: ModerationUsageDimension) => row.cache_hits,
      },
      {
        id: 'latency',
        header: t('Average latency'),
        cell: (row: ModerationUsageDimension) =>
          row.api_requests
            ? `${(row.api_latency_total_ms / row.api_requests).toFixed(0)} ms`
            : '—',
      },
    ],
    [t]
  )
  const apply = () =>
    setQueryParams({
      start_timestamp: Math.min(
        dateToUnixTimestamp(start),
        dateToUnixTimestamp(end)
      ),
      end_timestamp: Math.max(
        dateToUnixTimestamp(start),
        dateToUnixTimestamp(end)
      ),
      ...(Number.parseInt(userID) > 0
        ? { user_id: Number.parseInt(userID) }
        : {}),
      ...(Number.parseInt(tokenID) > 0
        ? { token_id: Number.parseInt(tokenID) }
        : {}),
    })
  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Moderation Statistics')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        <Button variant='outline' size='sm' onClick={() => setReviewOpen(true)}>
          <ScanSearch />
          {t('Daily review results')}
        </Button>
        <Button
          variant='outline'
          size='sm'
          onClick={() => void query.refetch()}
          disabled={query.isFetching}
        >
          <RefreshCw className={query.isFetching ? 'animate-spin' : ''} />
          {t('Refresh')}
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <DailyReviewDialog open={reviewOpen} onOpenChange={setReviewOpen} />
        <div className='flex flex-col gap-4'>
          <Card>
            <CardContent className='grid gap-3 p-4 sm:grid-cols-2 lg:grid-cols-4 lg:items-end'>
              <div className='flex flex-col gap-1.5'>
                <Label>{t('Time range')}</Label>
                <CompactDateTimeRangePicker
                  start={start}
                  end={end}
                  onChange={(range) => {
                    if (range.start) setStart(range.start)
                    if (range.end) setEnd(range.end)
                  }}
                />
              </div>
              <div className='flex flex-col gap-1.5'>
                <Label htmlFor='moderation-user'>{t('User ID')}</Label>
                <Input
                  id='moderation-user'
                  type='number'
                  placeholder={t('All users')}
                  value={userID}
                  onChange={(e) => setUserID(e.target.value)}
                />
              </div>
              <div className='flex flex-col gap-1.5'>
                <Label htmlFor='moderation-token'>{t('Token ID')}</Label>
                <Input
                  id='moderation-token'
                  type='number'
                  placeholder={t('All tokens')}
                  value={tokenID}
                  onChange={(e) => setTokenID(e.target.value)}
                />
              </div>
              <Button onClick={apply}>{t('Apply filters')}</Button>
            </CardContent>
          </Card>
          {query.isPending ? (
            <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7'>
              {Array.from({ length: 7 }, (_, index) => (
                <Skeleton key={index} className='h-24 rounded-xl' />
              ))}
            </div>
          ) : (
            data && (
              <>
                <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7'>
                  <StatCard
                    label={t('API requests')}
                    value={numberFormatter.format(data.api_requests)}
                  />
                  <StatCard
                    label={t('Passed')}
                    value={numberFormatter.format(data.api_passed)}
                  />
                  <StatCard
                    label={t('Successful')}
                    value={numberFormatter.format(data.api_succeeded)}
                  />
                  <StatCard
                    label={t('Violations')}
                    value={numberFormatter.format(data.api_violations)}
                  />
                  <StatCard
                    label={t('Failed')}
                    value={numberFormatter.format(data.api_failed)}
                  />
                  <StatCard
                    label={t('Average latency')}
                    value={`${data.api_latency_average_ms.toFixed(0)} ms`}
                  />
                  <StatCard
                    label={t('Sensitive word hits')}
                    value={numberFormatter.format(data.sensitive_word_hits)}
                  />
                </div>
                <div className='flex flex-wrap gap-2'>
                  <Badge variant='secondary'>
                    {t('Timeouts')}: {data.api_timeouts}
                  </Badge>
                  <Badge variant='secondary'>
                    {t('Circuit skips')}: {data.circuit_skips}
                  </Badge>
                  <Badge variant='secondary'>
                    {t('Cache hits')}: {data.cache_hits}
                  </Badge>
                </div>
                <Card>
                  <CardHeader>
                    <CardTitle className='flex items-center gap-2'>
                      <BarChart3 className='size-4' />
                      {t('Statistics by user and token')}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <StaticDataTable
                      columns={columns}
                      data={dimensions}
                      emptyContent={t('No moderation statistics found')}
                    />
                  </CardContent>
                </Card>
              </>
            )
          )}
          {query.isError && (
            <p className='text-destructive text-sm'>
              {query.error instanceof Error
                ? query.error.message
                : t('Failed to load moderation statistics')}
            </p>
          )}
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
