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
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { JSDOM } from 'jsdom'

const rootDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const outputDir = path.resolve(
  rootDir,
  process.env.HOME_PRERENDER_OUT_DIR || 'dist'
)
const builtIndex = await readFile(path.join(outputDir, 'index.html'), 'utf8')
const stylesheet = builtIndex.match(
  /<link[^>]+href=["']([^"']+\.css)["'][^>]*>/i
)
if (!stylesheet?.[1]) {
  throw new Error(
    'built index.html is missing its stylesheet; build the web app first'
  )
}
const stylesheetHref = stylesheet[1]

const dom = new JSDOM(
  `<!doctype html><html lang="zh-CN" class="light"><head><link rel="stylesheet" href="${stylesheetHref}"></head><body><script id="home-page-bootstrap" type="application/json"></script></body></html>`,
  { url: 'http://localhost/' }
)
const { window } = dom
const globals = {
  window,
  document: window.document,
  navigator: window.navigator,
  HTMLElement: window.HTMLElement,
  HTMLTemplateElement: window.HTMLTemplateElement,
  Node: window.Node,
  MutationObserver: window.MutationObserver,
  DOMParser: window.DOMParser,
  Image: window.Image,
  localStorage: window.localStorage,
  sessionStorage: window.sessionStorage,
  getComputedStyle: window.getComputedStyle.bind(window),
}
for (const [name, value] of Object.entries(globals)) {
  Object.defineProperty(globalThis, name, {
    configurable: true,
    value,
    writable: true,
  })
}

const matchMedia = (query) => ({
  matches: query.includes('prefers-color-scheme: light'),
  media: query,
  onchange: null,
  addEventListener: () => undefined,
  removeEventListener: () => undefined,
  addListener: () => undefined,
  removeListener: () => undefined,
  dispatchEvent: () => false,
})
Object.defineProperty(window, 'matchMedia', {
  configurable: true,
  value: matchMedia,
})
Object.defineProperty(globalThis, 'matchMedia', {
  configurable: true,
  value: matchMedia,
})

const homePageContent = await readFile(
  path.resolve(rootDir, '../common/home-page-content.html'),
  'utf8'
)
const config = { homePageContent }
const rendererModule = await import(
  path.join(rootDir, 'node_modules/.cache/home-renderer', 'home-renderer.cjs')
)
const renderer = rendererModule.renderHome || rendererModule.default?.renderHome
if (typeof renderer !== 'function') {
  throw new Error('home renderer does not export renderHome')
}

const { fragment, bootstrap } = await renderer(config)
await mkdir(outputDir, { recursive: true })
await writeFile(
  path.join(outputDir, 'home-fragment.html'),
  `<div id="home-prerender" data-home-prerender="true"><style data-home-prerender-style>#seo-content:has(#home-prerender) ~ #root{display:none}#home-prerender .landing-animate-fade-up,#home-prerender .landing-animate-fade-in,#home-prerender .landing-animate-scale-in,#home-prerender .landing-animate-fade-left,#home-prerender .landing-animate-fade-right{animation:none!important;opacity:1!important;transform:none!important}</style>${fragment}</div>\n`,
  'utf8'
)
await writeFile(
  path.join(outputDir, 'home-bootstrap.json'),
  `${JSON.stringify(bootstrap, null, 2)}\n`,
  'utf8'
)
dom.window.close()
