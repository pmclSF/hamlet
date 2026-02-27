export function esc(s) {
  if (s == null) return '';
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

export function card(label, value, sub) {
  return `<div class="card">
    <div class="card-value">${esc(value)}</div>
    <div class="card-label">${esc(label)}</div>
    ${sub ? `<div class="card-sub">${esc(sub)}</div>` : ''}
  </div>`;
}

export function badge(text, variant) {
  return `<span class="badge badge-${variant || 'default'}">${esc(text)}</span>`;
}

export function bar(pct, variant) {
  const p = Math.max(0, Math.min(100, pct));
  return `<div class="bar"><div class="bar-fill bar-${variant || 'accent'}" style="width:${p}%"></div></div>`;
}

export function frameworkBadge(fw) {
  if (!fw) return badge('unknown', 'muted');
  const colors = {
    jest: 'success',
    vitest: 'success',
    mocha: 'warning',
    jasmine: 'warning',
    cypress: 'accent',
    playwright: 'accent',
    selenium: 'muted',
    pytest: 'info',
    unittest: 'info',
    nose2: 'info',
    junit4: 'danger',
    junit5: 'danger',
    testng: 'danger',
    webdriverio: 'muted',
    puppeteer: 'muted',
    testcafe: 'muted',
  };
  return badge(fw, colors[fw] || 'default');
}

export function confidenceBadge(val) {
  const pct = Math.round(val);
  let variant = 'danger';
  if (pct >= 80) variant = 'success';
  else if (pct >= 50) variant = 'warning';
  return `<span class="confidence">
    ${badge(pct + '%', variant)}
    ${bar(pct, variant)}
  </span>`;
}

export function statusBadge(status) {
  const map = {
    queued: 'muted',
    running: 'accent',
    completed: 'success',
    failed: 'danger',
    converted: 'success',
    skipped: 'warning',
  };
  return badge(status, map[status] || 'default');
}

export function btn(text, cls, attrs) {
  const extra = attrs || '';
  return `<button class="btn ${cls || ''}" ${extra}>${esc(text)}</button>`;
}

export function btnIcon(text, icon, cls, attrs) {
  const extra = attrs || '';
  return `<button class="btn ${cls || ''}" ${extra}>${icon} ${esc(text)}</button>`;
}

export function copyBtn(text) {
  return `<button class="btn btn-sm btn-ghost copy-btn" data-copy="${esc(text)}" title="Copy to clipboard">
    <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"/><path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"/></svg>
  </button>`;
}

export function spinner() {
  return '<span class="spinner"></span>';
}

export function emptyState(msg) {
  return `<div class="empty-state">${esc(msg)}</div>`;
}
