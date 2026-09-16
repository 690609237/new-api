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
import { useQuery } from '@tanstack/react-query'

import { useStatus } from '@/hooks/use-status'
import { getStatusContent, type StatusContent } from '@/lib/api'
import { requireServerSuccess } from '@/lib/server-error-message'

export function useStatusContent<T>(
  enabledKey: string,
  content: StatusContent
): { items: T[]; loading: boolean } {
  const { status, loading: statusLoading } = useStatus()
  const enabled = Boolean(status) && status?.[enabledKey] !== false
  const query = useQuery({
    queryKey: ['status-content', content],
    queryFn: async () => {
      const response = requireServerSuccess(await getStatusContent<T>(content))
      return response.data ?? []
    },
    enabled,
    staleTime: 5 * 60 * 1000,
  })

  return {
    items: enabled ? (query.data ?? []) : [],
    loading: statusLoading || (enabled && query.isLoading),
  }
}
