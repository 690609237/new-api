/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the License,
or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import { formatTimestampToDate } from '@/lib/format'
import { handleServerError } from '@/lib/handle-server-error'
import { createServerError } from '@/lib/server-error-message'

import {
  getCurrentAuditLogCleanupTask,
  getSystemTask,
  startAuditLogCleanupTask,
} from '../api'
import { SettingsControlGroup } from '../components/settings-form-layout'
import type { AuditLogCleanupTask } from '../types'

const DEFAULT_RETENTION_DAYS = 30
const MAX_RETENTION_DAYS = 3650
const SECONDS_PER_DAY = 24 * 60 * 60

function isActiveAuditLogCleanupTask(task: AuditLogCleanupTask | null) {
  return task?.status === 'pending' || task?.status === 'running'
}

export function AuditLogCleanupControl() {
  const { t } = useTranslation()
  const [retentionDays, setRetentionDays] = useState(DEFAULT_RETENTION_DAYS)
  const [targetTimestamp, setTargetTimestamp] = useState<number | null>(null)
  const [showConfirmDialog, setShowConfirmDialog] = useState(false)
  const [isStarting, setIsStarting] = useState(false)
  const [task, setTask] = useState<AuditLogCleanupTask | null>(null)

  useEffect(() => {
    let cancelled = false

    async function fetchCurrentTask() {
      try {
        const response = await getCurrentAuditLogCleanupTask()
        if (!cancelled && response.success && response.data) {
          setTask(response.data)
        }
      } catch {
        /* The control remains usable when no previous task can be loaded. */
      }
    }

    fetchCurrentTask()
    return () => {
      cancelled = true
    }
  }, [])

  const taskActive = isActiveAuditLogCleanupTask(task)
  const taskId = task?.task_id

  useEffect(() => {
    if (!taskId || !taskActive) return

    let cancelled = false
    const interval = window.setInterval(async () => {
      try {
        const response = await getSystemTask(taskId)
        if (cancelled || !response.success || !response.data) return

        setTask(response.data)
        if (!isActiveAuditLogCleanupTask(response.data)) {
          if (response.data.status === 'succeeded') {
            const count =
              response.data.result?.deleted_count ??
              response.data.state?.processed ??
              0
            toast.success(
              count > 0
                ? t('{{count}} audit log entries removed.', { count })
                : t('No audit log entries matched the retention period.')
            )
          } else if (response.data.status === 'failed') {
            handleServerError(response.data, t('Failed to clean audit logs'))
          }
        }
      } catch {
        /* Keep polling until the task reaches a terminal state. */
      }
    }, 1000)

    return () => {
      cancelled = true
      window.clearInterval(interval)
    }
  }, [taskActive, taskId, t])

  const requestCleanup = () => {
    if (
      !Number.isInteger(retentionDays) ||
      retentionDays < 1 ||
      retentionDays > MAX_RETENTION_DAYS
    ) {
      toast.error(
        t('Enter a retention period between 1 and {{max}} days.', {
          max: MAX_RETENTION_DAYS,
        })
      )
      return
    }

    setTargetTimestamp(
      Math.floor(Date.now() / 1000) - retentionDays * SECONDS_PER_DAY
    )
    setShowConfirmDialog(true)
  }

  const cleanAuditLogs = async () => {
    if (!targetTimestamp) return

    setIsStarting(true)
    try {
      const response = await startAuditLogCleanupTask(targetTimestamp)
      if (!response.success) {
        throw createServerError(response, t('Failed to clean audit logs'))
      }
      if (!response.data) {
        throw new Error(t('Failed to clean audit logs'))
      }
      setTask(response.data)
      setShowConfirmDialog(false)
      toast.success(t('Audit log cleanup task started.'))
    } catch (error) {
      handleServerError(error, t('Failed to clean audit logs'))
    } finally {
      setIsStarting(false)
    }
  }

  const progress = Math.min(100, Math.max(0, task?.state?.progress ?? 0))
  const processed = task?.state?.processed ?? 0
  const total = task?.state?.total ?? 0
  const formattedCutoff = targetTimestamp
    ? formatTimestampToDate(targetTimestamp * 1000, 'milliseconds')
    : ''

  return (
    <SettingsControlGroup className='space-y-3'>
      <div>
        <h4 className='text-sm font-medium'>{t('Clean audit logs')}</h4>
        <p className='text-muted-foreground text-sm'>
          {t(
            'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.'
          )}
        </p>
      </div>

      <div className='flex flex-wrap items-end gap-3'>
        <div className='grid gap-1.5'>
          <Label htmlFor='audit-log-retention-days' className='text-xs'>
            {t('Days to retain')}
          </Label>
          <Input
            id='audit-log-retention-days'
            type='number'
            min={1}
            max={MAX_RETENTION_DAYS}
            value={retentionDays}
            onChange={(event) =>
              setRetentionDays(Number(event.currentTarget.value))
            }
            className='w-[120px]'
          />
        </div>
        <Button
          type='button'
          variant='destructive'
          onClick={requestCleanup}
          disabled={isStarting || taskActive}
        >
          {isStarting || taskActive ? t('Cleaning...') : t('Clean audit logs')}
        </Button>
      </div>

      {task && (
        <div className='rounded-md border p-3'>
          <div className='mb-2 flex items-center justify-between gap-3 text-sm'>
            <span className='font-medium'>
              {t('Audit log cleanup progress')}
            </span>
            <span className='text-muted-foreground tabular-nums'>
              {progress}%
            </span>
          </div>
          <Progress value={progress} />
          <div className='text-muted-foreground mt-2 text-xs'>
            {t('{{processed}} of {{total}} audit log entries processed.', {
              processed,
              total,
            })}
          </div>
          {task.status === 'failed' && task.error && (
            <div className='text-destructive mt-2 text-xs'>{task.error}</div>
          )}
        </div>
      )}

      <ConfirmDialog
        open={showConfirmDialog}
        onOpenChange={setShowConfirmDialog}
        title={t('Confirm audit log cleanup')}
        desc={
          <span>
            {t(
              'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.',
              { date: formattedCutoff }
            )}{' '}
            {t('This action cannot be undone.')}
          </span>
        }
        destructive
        handleConfirm={cleanAuditLogs}
        isLoading={isStarting}
        confirmText={t('Delete audit logs')}
      />
    </SettingsControlGroup>
  )
}
