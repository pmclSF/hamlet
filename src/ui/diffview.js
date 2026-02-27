/* global window */
import * as api from './api.js';
import { esc, spinner } from './components.js';
import { computeDiff, buildSideBySide } from './diff.js';

export async function render(container, state, actions) {
  const { diffSource, diffOutput } = state;
  if (!diffSource || !diffOutput) {
    container.innerHTML =
      '<div class="error-msg">No files specified for diff.</div>';
    return;
  }

  container.innerHTML = `<div class="loading">${spinner()} Loading files...</div>`;

  let oldContent, newContent;
  try {
    const [oldRes, newRes] = await Promise.all([
      api.readFile(diffSource),
      api.readFile(diffOutput),
    ]);
    oldContent = oldRes.content;
    newContent = newRes.content;
  } catch (err) {
    container.innerHTML = `<div class="error-msg">Failed to load files: ${esc(err.message)}</div>`;
    return;
  }

  const diffEntries = computeDiff(oldContent, newContent);
  const pairs = buildSideBySide(diffEntries);
  const stats = {
    added: diffEntries.filter((e) => e.type === 'insert').length,
    removed: diffEntries.filter((e) => e.type === 'delete').length,
    unchanged: diffEntries.filter((e) => e.type === 'equal').length,
  };

  container.innerHTML = `
    <div class="view-header">
      <h1>Diff View</h1>
      <div class="view-actions">
        <button class="btn btn-ghost" id="back-btn">\u2190 Back</button>
      </div>
    </div>
    <div class="diff-meta">
      <div class="diff-files">
        <span class="diff-label">Original:</span> <span class="mono">${esc(diffSource)}</span>
        <span class="diff-label">Converted:</span> <span class="mono">${esc(diffOutput)}</span>
      </div>
      <div class="diff-stats">
        <span class="text-success">+${stats.added}</span>
        <span class="text-danger">-${stats.removed}</span>
        <span class="text-secondary">${stats.unchanged} unchanged</span>
      </div>
    </div>
    <div class="diff-container">
      <div class="diff-pane diff-left">
        <div class="diff-pane-header">Original</div>
        <div class="diff-lines">
          ${pairs.map((p) => renderLeftLine(p)).join('')}
        </div>
      </div>
      <div class="diff-pane diff-right">
        <div class="diff-pane-header">Converted</div>
        <div class="diff-lines">
          ${pairs.map((p) => renderRightLine(p)).join('')}
        </div>
      </div>
    </div>`;

  attachEvents(container, state, actions);
  syncScroll(container);
}

function renderLeftLine(pair) {
  if (pair.type === 'insert') {
    return '<div class="diff-line diff-empty"><span class="diff-num"></span><span class="diff-code"></span></div>';
  }
  const cls = pair.type === 'delete' ? 'diff-removed' : '';
  return `<div class="diff-line ${cls}"><span class="diff-num">${pair.leftNum}</span><span class="diff-code">${highlightLine(esc(pair.left))}</span></div>`;
}

function renderRightLine(pair) {
  if (pair.type === 'delete') {
    return '<div class="diff-line diff-empty"><span class="diff-num"></span><span class="diff-code"></span></div>';
  }
  const cls = pair.type === 'insert' ? 'diff-added' : '';
  const content = pair.right || '';
  return `<div class="diff-line ${cls}"><span class="diff-num">${pair.rightNum}</span><span class="diff-code">${highlightTodo(highlightLine(esc(content)))}</span></div>`;
}

function highlightLine(html) {
  // Basic syntax highlights for common keywords
  return html
    .replace(
      /\b(import|export|from|const|let|var|function|class|return|async|await|describe|it|test|expect|beforeEach|afterEach)\b/g,
      '<span class="hl-keyword">$1</span>'
    )
    .replace(
      /(&#x27;[^&#]*&#x27;|&quot;[^&]*&quot;)/g,
      '<span class="hl-string">$1</span>'
    );
}

function highlightTodo(html) {
  return html.replace(
    /(HAMLET-TODO[^<]*)/g,
    '<mark class="todo-marker">$1</mark>'
  );
}

function syncScroll(container) {
  const left = container.querySelector('.diff-left .diff-lines');
  const right = container.querySelector('.diff-right .diff-lines');
  if (!left || !right) return;

  let syncing = false;
  function sync(source, target) {
    if (syncing) return;
    syncing = true;
    target.scrollTop = source.scrollTop;
    syncing = false;
  }
  left.addEventListener('scroll', () => sync(left, right));
  right.addEventListener('scroll', () => sync(right, left));
}

function attachEvents(container, _state, _actions) {
  const back = container.querySelector('#back-btn');
  if (back) back.addEventListener('click', () => window.history.back());
}
