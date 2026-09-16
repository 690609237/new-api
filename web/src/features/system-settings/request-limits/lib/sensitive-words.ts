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
import { z } from 'zod'

export const createSensitiveSchema = (t: (key: string) => string) =>
  z.object({
    CheckSensitiveEnabled: z.boolean(),
    CheckSensitiveOnPromptEnabled: z.boolean(),
    SensitiveWords: z
      .string()
      .optional()
      .superRefine((value, context) => {
        for (const line of (value ?? '').split(/\r?\n/)) {
          const rule = line.trim()
          if (!rule) continue

          const keywords = rule
            .replaceAll('｜', '|')
            .split('|')
            .map((keyword) => keyword.trim())
          if (keywords.some((keyword) => !keyword)) {
            context.addIssue({
              code: 'custom',
              message: t('Combined rules cannot contain empty keywords'),
            })
            return
          }
          if (keywords.length > 5) {
            context.addIssue({
              code: 'custom',
              message: t('Each combined rule can contain at most 5 keywords'),
            })
            return
          }
        }
      }),
  })

export type SensitiveFormValues = z.infer<
  ReturnType<typeof createSensitiveSchema>
>
