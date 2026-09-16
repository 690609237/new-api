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
import { describe, expect, it } from 'vitest'

import { createSensitiveSchema } from '../lib/sensitive-words'

const schema = createSensitiveSchema((key) => key)
const switches = {
  CheckSensitiveEnabled: true,
  CheckSensitiveOnPromptEnabled: true,
}

describe('sensitive word rule validation', () => {
  it('accepts single keywords and combined rules with up to five keywords', () => {
    expect(
      schema.safeParse({
        ...switches,
        SensitiveWords: 'single\none|two|three|four|five\nalpha｜beta',
      }).success
    ).toBe(true)
  })

  it('rejects combined rules with more than five keywords', () => {
    const result = schema.safeParse({
      ...switches,
      SensitiveWords: 'one|two|three|four|five|six',
    })

    expect(result.success).toBe(false)
    if (result.success) return
    expect(result.error.issues[0]?.message).toBe(
      'Each combined rule can contain at most 5 keywords'
    )
  })

  it('rejects empty keywords in a combined rule', () => {
    const result = schema.safeParse({
      ...switches,
      SensitiveWords: 'one||two',
    })

    expect(result.success).toBe(false)
    if (result.success) return
    expect(result.error.issues[0]?.message).toBe(
      'Combined rules cannot contain empty keywords'
    )
  })
})
