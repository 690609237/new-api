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
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { AuditLogCleanupControl } from '../audit-log-cleanup-control'

const apiMocks = vi.hoisted(() => ({
  getCurrentAuditLogCleanupTask: vi.fn(),
  getSystemTask: vi.fn(),
  startAuditLogCleanupTask: vi.fn(),
}))

vi.mock('../../api', () => apiMocks)
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

describe('audit log cleanup control', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    apiMocks.getCurrentAuditLogCleanupTask.mockResolvedValue({
      success: true,
      data: null,
    })
    apiMocks.startAuditLogCleanupTask.mockResolvedValue({
      success: true,
      data: {
        id: 1,
        task_id: 'audit-cleanup-1',
        type: 'audit_log_cleanup',
        status: 'succeeded',
        state: { total: 2, processed: 2, progress: 100, remaining: 0 },
        result: { deleted_count: 2 },
        created_at: 1,
        updated_at: 1,
      },
    })
  })

  it('defaults to 30 days and starts an audit-only physical cleanup after confirmation', async () => {
    const user = userEvent.setup()
    render(<AuditLogCleanupControl />)

    const retentionInput = screen.getByLabelText('Days to retain')
    expect(retentionInput).toHaveValue(30)

    const earliestCutoff = Math.floor(Date.now() / 1000) - 30 * 24 * 60 * 60
    await user.click(screen.getByRole('button', { name: 'Clean audit logs' }))

    const dialog = await screen.findByRole('alertdialog')
    expect(dialog).toHaveTextContent('Only the audit_logs table is affected')
    expect(dialog).toHaveTextContent(
      'usage logs and server log files will remain unchanged'
    )

    await user.click(screen.getByRole('button', { name: 'Delete audit logs' }))

    await waitFor(() => {
      expect(apiMocks.startAuditLogCleanupTask).toHaveBeenCalledOnce()
    })
    const latestCutoff = Math.floor(Date.now() / 1000) - 30 * 24 * 60 * 60
    const targetTimestamp = apiMocks.startAuditLogCleanupTask.mock.calls[0]?.[0]
    expect(targetTimestamp).toBeGreaterThanOrEqual(earliestCutoff)
    expect(targetTimestamp).toBeLessThanOrEqual(latestCutoff)
  })
})
