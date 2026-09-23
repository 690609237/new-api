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
import type { TFunction } from 'i18next'

function escapeHtml(value: string): string {
  return value.replaceAll(
    /[&<>'"]/g,
    (character) =>
      ({
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        "'": '&#39;',
        '"': '&quot;',
      })[character] ?? character
  )
}

/**
 * Default HTML shown in the configurable information area on the home page.
 *
 * Administrators can copy this template from System Information and edit it
 * directly. Keep the styles scoped under `.home-feature-stack` so configured
 * content cannot accidentally change the surrounding home page.
 */
export function getDefaultHomePageContent(t: TFunction): string {
  const summaryLabel = escapeHtml(
    t('Platform pricing and channel source summary')
  )
  const rechargeTitle = escapeHtml(t('1:1 platform recharge'))
  const pricingDescription = escapeHtml(
    t(
      'Model prices use the $ symbol. Overseas models follow international pricing, while domestic models follow domestic pricing.'
    )
  )
  const tokenGroupsTitle = escapeHtml(t('First-party token groups'))

  return `<style>
.home-feature-stack {
  display: grid;
  width: 100%;
  gap: 12px;
  color: var(--foreground);
  font-family: inherit;
}

.home-feature-stack *,
.home-feature-stack *::before,
.home-feature-stack *::after {
  box-sizing: border-box;
}

.home-feature-card {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 20px;
  border: 1px solid color-mix(in oklch, var(--feature-color) 16%, transparent);
  border-radius: 24px;
  background: color-mix(in oklch, var(--feature-color) 5%, transparent);
  box-shadow: 0 8px 28px -24px color-mix(in oklch, var(--feature-color) 70%, transparent);
}

.home-feature-card--pricing {
  --feature-color: var(--info);
}

.home-feature-card--groups {
  --feature-color: var(--chart-3);
  align-items: center;
}

.home-feature-icon {
  display: grid;
  width: 48px;
  height: 48px;
  flex: 0 0 48px;
  place-items: center;
  border-radius: 16px;
  background: color-mix(in oklch, var(--feature-color) 13%, transparent);
  color: var(--feature-color);
}

.home-feature-icon svg {
  width: 24px;
  height: 24px;
}

.home-feature-body {
  min-width: 0;
  flex: 1;
}

.home-feature-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.home-feature-title {
  margin: 0;
  color: var(--foreground);
  font-size: 16px;
  font-weight: 650;
  line-height: 24px;
}

.home-feature-description {
  margin: 8px 0 0;
  color: var(--muted-foreground);
  font-size: 14px;
  line-height: 24px;
}

.home-feature-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.home-feature-tag {
  display: inline-flex;
  min-height: 24px;
  align-items: center;
  border: 1px solid color-mix(in oklch, var(--feature-color) 28%, transparent);
  border-radius: 999px;
  padding: 2px 10px;
  background: color-mix(in oklch, var(--background) 65%, transparent);
  color: color-mix(in oklch, var(--feature-color) 78%, var(--foreground));
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 18px;
}

.home-feature-groups-row {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

@media (max-width: 480px) {
  .home-feature-card {
    gap: 12px;
    padding: 16px;
    border-radius: 20px;
  }

  .home-feature-icon {
    width: 42px;
    height: 42px;
    flex-basis: 42px;
    border-radius: 14px;
  }

  .home-feature-groups-row {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>

<div class="home-feature-stack" role="list" aria-label="${summaryLabel}">
  <article class="home-feature-card home-feature-card--pricing" role="listitem">
    <span class="home-feature-icon" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"></circle>
        <path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8"></path>
        <path d="M12 18V6"></path>
      </svg>
    </span>
    <div class="home-feature-body">
      <div class="home-feature-heading">
        <h2 class="home-feature-title">${rechargeTitle}</h2>
        <span class="home-feature-tag">$ USD</span>
      </div>
      <p class="home-feature-description">${pricingDescription}</p>
    </div>
  </article>

  <article class="home-feature-card home-feature-card--groups" role="listitem">
    <span class="home-feature-icon" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="m12 2 9 5-9 5-9-5 9-5Z"></path>
        <path d="m3 12 9 5 9-5"></path>
        <path d="m3 17 9 5 9-5"></path>
      </svg>
    </span>
    <div class="home-feature-groups-row">
      <h2 class="home-feature-title">${tokenGroupsTitle}</h2>
      <div class="home-feature-tags" aria-label="${tokenGroupsTitle}">
        <span class="home-feature-tag">default</span>
        <span class="home-feature-tag">standard</span>
      </div>
    </div>
  </article>
</div>`
}
