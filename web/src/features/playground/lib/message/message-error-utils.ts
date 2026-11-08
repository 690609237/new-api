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
import { MESSAGE_STATUS } from '../../constants'
import type { Message } from '../../types'
import { getMessageContent } from './message-utils'

export const MODEL_PRICING_SETTINGS_PATH =
  '/system-settings/billing/model-pricing'

const MODEL_PRICE_ERROR_CODE = 'model_price_error'
const INSUFFICIENT_QUOTA_ERROR_CODE = 'insufficient_user_quota'
const DO_REQUEST_FAILED_ERROR_CODE = 'do_request_failed'
const GET_CHANNEL_FAILED_ERROR_CODE = 'get_channel_failed'
export const FALLBACK_ERROR_CONTENT = 'An unknown error occurred'

type MessageErrorState = {
  content: string
  kind: 'generic' | 'model-price' | 'quota' | 'request-failed'
  showSettingsLink: boolean
  showChannelsLink: boolean
}

export function isAdminRole(role?: number | null): boolean {
  return role != null && role >= 10
}

export function isErrorMessage(message: Message): boolean {
  return message.status === MESSAGE_STATUS.ERROR
}

export function getMessageErrorState(
  message: Message,
  isAdmin: boolean
): MessageErrorState | null {
  if (!isErrorMessage(message)) {
    return null
  }

  const content = getMessageContent(message) || FALLBACK_ERROR_CONTENT
  const isModelPriceError = message.errorCode === MODEL_PRICE_ERROR_CODE
  const isQuotaError = message.errorCode === INSUFFICIENT_QUOTA_ERROR_CODE
  const isRequestFailedError = [
    DO_REQUEST_FAILED_ERROR_CODE,
    GET_CHANNEL_FAILED_ERROR_CODE,
  ].includes(message.errorCode ?? '')
  let kind: MessageErrorState['kind'] = 'generic'

  if (isModelPriceError) {
    kind = 'model-price'
  } else if (isQuotaError) {
    kind = 'quota'
  } else if (isRequestFailedError) {
    kind = 'request-failed'
  }

  return {
    content,
    kind,
    showSettingsLink: isModelPriceError && isAdmin,
    showChannelsLink: isRequestFailedError && isAdmin,
  }
}
