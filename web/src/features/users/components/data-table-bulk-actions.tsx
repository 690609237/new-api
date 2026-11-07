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
import type { Table } from '@tanstack/react-table'
import { Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { DataTableBulkActions as BulkActionsToolbar } from '@/components/data-table'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { handleServerError } from '@/lib/handle-server-error'
import { AuthOperationError } from '@/lib/secure-verification'

import { batchDeleteUsers } from '../api'
import type { User } from '../types'
import { useUsers } from './users-provider'

interface DataTableBulkActionsProps {
  table: Table<User>
}

export function DataTableBulkActions({ table }: DataTableBulkActionsProps) {
  const { t } = useTranslation()
  const { requestVerification, triggerRefresh, verificationActive } = useUsers()
  const [deleteTargets, setDeleteTargets] = useState<User[] | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)
  const selectedRows = table.getFilteredSelectedRowModel().rows

  const handleDelete = async () => {
    if (!deleteTargets?.length) return

    const ids = [...new Set(deleteTargets.map((user) => user.id))].sort(
      (a, b) => a - b
    )
    setIsDeleting(true)
    try {
      const proof = await requestVerification({
        scope: 'admin.user.bulk_delete',
        context: { user_ids: ids },
        title: t('Verify to delete selected users'),
        description: t(
          'Confirm your identity before permanently deleting {{count}} user accounts.',
          { count: ids.length }
        ),
      })
      if (!proof) return

      const result = await batchDeleteUsers(ids, proof.proof_token)
      if (result.success) {
        const deleted = result.data?.deleted ?? ids.length
        toast.success(
          t('Successfully deleted {{count}} users', { count: deleted })
        )
        table.resetRowSelection()
        setDeleteTargets(null)
        triggerRefresh()
      } else {
        handleServerError(
          result,
          t('Failed to delete {{count}} users', { count: ids.length })
        )
      }
    } catch (error) {
      handleServerError(
        AuthOperationError.from(error),
        t('Failed to delete {{count}} users', { count: ids.length })
      )
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <BulkActionsToolbar table={table} entityName='user'>
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant='destructive'
              size='icon'
              className='size-8'
              aria-label={t('Delete selected users')}
              disabled={isDeleting}
              onClick={() =>
                setDeleteTargets(selectedRows.map((row) => row.original))
              }
            />
          }
        >
          <Trash2 aria-hidden='true' />
          <span className='sr-only'>{t('Delete selected users')}</span>
        </TooltipTrigger>
        <TooltipContent>
          <p>{t('Delete selected users')}</p>
        </TooltipContent>
      </Tooltip>

      <ConfirmDialog
        destructive
        open={deleteTargets !== null && !verificationActive}
        onOpenChange={(open) => {
          if (!open && !isDeleting) setDeleteTargets(null)
        }}
        title={t('Delete {{count}} users?', {
          count: deleteTargets?.length ?? 0,
        })}
        desc={t('This action cannot be undone.')}
        confirmText={isDeleting ? t('Deleting...') : t('Delete')}
        isLoading={isDeleting}
        disabled={!deleteTargets?.length}
        handleConfirm={handleDelete}
      />
    </BulkActionsToolbar>
  )
}
