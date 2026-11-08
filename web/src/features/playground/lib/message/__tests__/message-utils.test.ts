/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or (at your
option) any later version.
*/
import { describe, expect, test } from 'vitest'

import {
  buildMessageContent,
  createUserMessage,
  formatMessageForAPI,
} from '../message-utils'

describe('playground message attachments', () => {
  const attachments = [
    {
      id: 'image-1',
      filename: 'screenshot.png',
      mediaType: 'image/png',
      url: 'data:image/png;base64,abc',
    },
    {
      id: 'file-1',
      filename: 'notes.txt',
      mediaType: 'text/plain',
      url: 'data:text/plain;base64,YWJj',
    },
  ]

  test('builds image and file content parts while preserving text', () => {
    expect(buildMessageContent('Describe these', attachments)).toEqual([
      { type: 'text', text: 'Describe these' },
      { type: 'image_url', image_url: { url: attachments[0].url } },
      {
        type: 'file',
        file: {
          filename: attachments[1].filename,
          file_data: attachments[1].url,
        },
      },
    ])
  })

  test('includes attachments in the API message payload', () => {
    const message = createUserMessage('Read this', 0, attachments)
    const payload = formatMessageForAPI(message)
    expect(payload.role).toBe('user')
    expect(payload.content).toEqual([
      { type: 'text', text: 'Read this' },
      { type: 'image_url', image_url: { url: attachments[0].url } },
      {
        type: 'file',
        file: {
          filename: attachments[1].filename,
          file_data: attachments[1].url,
        },
      },
    ])
  })
})
