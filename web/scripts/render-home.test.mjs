/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or (at your
option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { copyFileSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'

import { JSDOM } from 'jsdom'

const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const runner = path.join(rootDir, 'scripts/render-home.mjs')
const bunExecutable = process.env.BUN_EXECUTABLE || 'bun'

function renderFixture(t) {
  const outputDir = mkdtempSync(path.join(os.tmpdir(), 'home-prerender-test-'))
  t.after(() => rmSync(outputDir, { recursive: true, force: true }))
  copyFileSync(
    path.join(rootDir, 'dist/index.html'),
    path.join(outputDir, 'index.html')
  )
  const result = spawnSync(bunExecutable, [runner], {
    cwd: rootDir,
    encoding: 'utf8',
    timeout: 30_000,
    env: {
      ...process.env,
      HOME_PRERENDER_OUT_DIR: outputDir,
    },
  })
  assert.ifError(result.error)
  return { result, outputDir }
}

test('renders the source homepage with its brand, links and completed content', (t) => {
  const { result, outputDir } = renderFixture(t)
  assert.equal(result.status, 0, result.stderr || result.error?.message)
  const fragment = readFileSync(
    path.join(outputDir, 'home-fragment.html'),
    'utf8'
  )
  const bootstrap = JSON.parse(
    readFileSync(path.join(outputDir, 'home-bootstrap.json'), 'utf8')
  )
  const dom = new JSDOM(fragment)
  t.after(() => dom.window.close())
  const { document } = dom.window

  assert.match(document.querySelector('header').textContent, /ModelPass/)
  assert.match(document.querySelector('footer').textContent, /ModelPass/)
  assert.equal(document.querySelector('[data-slot="skeleton"]'), null)
  assert.equal(document.querySelector('[data-page="home"]') !== null, true)
  assert.equal(document.querySelector('a[href="/sign-up"]') !== null, true)
  assert.equal(document.querySelector('a[href="/pricing"]') !== null, true)
  assert.equal(document.querySelector('a[href="/docs"]') !== null, true)
  const backendContent = readFileSync(
    path.resolve(rootDir, '../common/home-page-content.html'),
    'utf8'
  )
  assert.equal(bootstrap.homePageContent, backendContent)
  assert.equal(bootstrap.systemConfig.systemName, 'ModelPass')
  const template = document.querySelector('template[shadowrootmode="open"]')
  assert.equal(template !== null, true)
  assert.match(
    template.content.querySelector('.html-content-root').textContent,
    /平台 1:1 充值/
  )
  assert.match(
    template.content
      .querySelector('link[rel="stylesheet"]')
      .getAttribute('href'),
    /\/static\/css\/index\.[a-z\d]+\.css$/
  )

  // Only entrance animations become visible; the closed mobile overlay keeps
  // its hidden state instead of being exposed over the pre-rendered page.
  assert.equal(
    document.querySelector('header + div').classList.contains('opacity-0'),
    true
  )
  assert.match(
    document.querySelector('[data-home-prerender-style]').textContent,
    /#seo-content:has\(#home-prerender\) ~ #root\{display:none\}/
  )
})
