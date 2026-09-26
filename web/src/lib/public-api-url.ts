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

For commercial licensing, please contact support@quantumnous.com
*/
/**
 * Public API origin used in client-facing examples, integrations, and copied
 * connection metadata. The dashboard may be served from a different origin.
 */
export const PUBLIC_API_BASE_URL = 'https://api.modelpass.work'

/** Keep unrelated configured regions while retiring legacy site/IP API URLs. */
export function resolvePublicApiBaseUrl(value?: string): string {
  if (!value?.trim()) return PUBLIC_API_BASE_URL

  try {
    const url = new URL(value)
    const hostname = url.hostname.toLowerCase()
    if (
      hostname === 'www.modelpass.work' ||
      hostname === 'modelpass.work' ||
      hostname === 'api.modelpass.work' ||
      hostname === 'localhost' ||
      /^\d{1,3}(?:\.\d{1,3}){3}$/.test(hostname) ||
      hostname.startsWith('[')
    ) {
      return PUBLIC_API_BASE_URL
    }
  } catch {
    return PUBLIC_API_BASE_URL
  }

  return value
}
