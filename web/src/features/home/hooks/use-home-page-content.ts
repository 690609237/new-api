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
import i18next from 'i18next'
import { useEffect, useState } from 'react'

import { handleServerError } from '@/lib/handle-server-error'

import { getHomePageContent } from '../api'
import type { HomePageContentResult } from '../types'

const STORAGE_KEY = 'home_page_information_content_v2'

function readStartupHomePageContent(): string | null {
  const bootstrap = document.querySelector<HTMLScriptElement>(
    '#home-page-bootstrap'
  )
  if (bootstrap) {
    try {
      const content: unknown = JSON.parse(bootstrap.textContent ?? '')
      if (typeof content === 'string') return content
    } catch {
      // Development HTML may still contain the unresolved template token.
    }
  }

  return null
}

function readLegacyCachedHomePageContent(): string | null {
  return localStorage.getItem(STORAGE_KEY)
}

/**
 * Loads the configurable Markdown/HTML content for the home-page information
 * area. The surrounding home-page layout is always rendered by the app.
 */
export function useHomePageContent(): HomePageContentResult {
  const [content, setContent] = useState<string>(
    () => readStartupHomePageContent() ?? readLegacyCachedHomePageContent() ?? ''
  )
  const [isLoaded, setIsLoaded] = useState(false)

  useEffect(() => {
    let mounted = true

    if (readStartupHomePageContent() !== null) {
      setIsLoaded(true)
      return () => {
        mounted = false
      }
    }

    const loadContent = async () => {
      try {
        const response = await getHomePageContent()
        const { success, data } = response

        if (!mounted) return

        if (success && data) {
          setContent(data)
          localStorage.setItem(STORAGE_KEY, data)
        } else {
          setContent('')
          localStorage.removeItem(STORAGE_KEY)
        }
      } catch (error) {
        if (!mounted) return
        handleServerError(error, i18next.t('Failed to load home page content'))
      } finally {
        if (mounted) {
          setIsLoaded(true)
        }
      }
    }

    loadContent()

    return () => {
      mounted = false
    }
  }, [])

  return { content, isLoaded }
}
