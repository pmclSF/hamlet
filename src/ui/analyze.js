/* global navigator, alert */
import * as api from './api.js';
import {
  esc,
  card,
  badge,
  frameworkBadge,
  confidenceBadge,
  copyBtn,
  spinner,
  emptyState,
} from './components.js';
import { computeDiff, buildSideBySide } from './diff.js';

let report = null;
let selectedFile = null;
let sortCol = 'path';
let sortAsc = true;
let searchQuery = '';
let typeFilter = '';

export async function render(container, state, actions) {
  if (!report) {
    container.innerHTML = `<div class="loading">${spinner()} Analyzing project...</div>`;
    try {
      report = await api.analyze(state.root);
    } catch (err) {
      container.innerHTML = `<div class="error-msg">Analysis failed: ${esc(err.message)}</div>`;
      return;
    }
  }
  renderContent(container, state, actions);
}

function renderContent(container, state, actions) {
  const s = report.summary;
  const files = getFilteredFiles();

  container.innerHTML = `
    <div class="view-header">
      <h1>Project Analysis</h1>
      <div class="view-actions">
        <button class="btn btn-ghost" id="dl-report">Download Report</button>
        <button class="btn btn-ghost" id="refresh-btn">Refresh</button>
      </div>
    </div>
    <div class="cards-row">
      ${card('Files Scanned', s.fileCount)}
      ${card('Test Files', s.testFileCount)}
      ${card('Frameworks', s.frameworksDetected.length, s.frameworksDetected.join(', ') || 'none')}
      ${card('Avg Confidence', s.confidenceAvg > 0 ? s.confidenceAvg + '%' : '--')}
      ${card('Directions', s.directionsSupported.length)}
    </div>
    <div class="split-layout">
      <div class="table-panel">
        <div class="table-toolbar">
          <input type="text" class="search-input" id="search" placeholder="Search files..." value="${esc(searchQuery)}" />
          <select class="select-input" id="type-filter">
            <option value="">All types</option>
            <option value="test" ${typeFilter === 'test' ? 'selected' : ''}>Tests</option>
            <option value="config" ${typeFilter === 'config' ? 'selected' : ''}>Config</option>
            <option value="helper" ${typeFilter === 'helper' ? 'selected' : ''}>Helper</option>
            <option value="fixture" ${typeFilter === 'fixture' ? 'selected' : ''}>Fixture</option>
          </select>
          ${files.length > 0 ? `<button class="btn btn-primary" id="convert-all-btn">Convert All Tests</button>` : ''}
        </div>
        ${renderTable(files)}
      </div>
      <div class="detail-panel ${selectedFile ? 'open' : ''}" id="detail-panel">
        ${selectedFile ? renderDetail(selectedFile, report, actions) : '<div class="detail-empty">Select a file to view details</div>'}
      </div>
    </div>
    <div id="preview-full-width"></div>`;

  attachEvents(container, state, actions);
}

function getFilteredFiles() {
  let files = report.files;
  if (searchQuery) {
    const q = searchQuery.toLowerCase();
    files = files.filter((f) => f.path.toLowerCase().includes(q));
  }
  if (typeFilter) {
    files = files.filter((f) => f.type === typeFilter);
  }
  files = [...files].sort((a, b) => {
    let va = a[sortCol];
    let vb = b[sortCol];
    if (va == null) va = '';
    if (vb == null) vb = '';
    if (typeof va === 'number' && typeof vb === 'number') {
      return sortAsc ? va - vb : vb - va;
    }
    va = String(va).toLowerCase();
    vb = String(vb).toLowerCase();
    if (va < vb) return sortAsc ? -1 : 1;
    if (va > vb) return sortAsc ? 1 : -1;
    return 0;
  });
  return files;
}

function sortIcon(col) {
  if (sortCol !== col) return '';
  return sortAsc ? ' \u25B2' : ' \u25BC';
}

function renderTable(files) {
  if (files.length === 0) return emptyState('No files match current filters');
  return `<div class="table-wrap"><table class="data-table">
    <thead><tr>
      <th class="sortable" data-col="path">Path${sortIcon('path')}</th>
      <th class="sortable" data-col="type">Type${sortIcon('type')}</th>
      <th class="sortable" data-col="framework">Framework${sortIcon('framework')}</th>
      <th class="sortable" data-col="confidence">Confidence${sortIcon('confidence')}</th>
      <th>Candidates</th>
      <th>Warnings</th>
    </tr></thead>
    <tbody>${files
      .map(
        (
          f
        ) => `<tr class="file-row ${selectedFile && selectedFile.path === f.path ? 'selected' : ''}" data-path="${esc(f.path)}">
      <td class="mono">${esc(f.path)}</td>
      <td>${badge(f.type, f.type === 'test' ? 'success' : 'muted')}</td>
      <td>${frameworkBadge(f.framework)}</td>
      <td>${confidenceBadge(f.confidence)}</td>
      <td>${
        f.candidates.length > 0
          ? f.candidates
              .slice(0, 2)
              .map((c) => badge(c.framework, 'muted'))
              .join(' ')
          : '--'
      }</td>
      <td>${f.warnings.length > 0 ? badge(f.warnings.length, 'warning') : '--'}</td>
    </tr>`
      )
      .join('')}</tbody>
  </table></div>`;
}

function renderDetail(file, rpt, actions) {
  const dirs = rpt.summary.directionsSupported.filter(
    (d) => d.from === file.framework
  );
  const inboundDirs = rpt.summary.directionsSupported.filter(
    (d) => d.to === file.framework
  );
  const cmd =
    dirs.length > 0
      ? `hamlet convert ${file.path} --from ${file.framework} --to ${dirs[0].to} -o hamlet-out/`
      : '';
  const maxScore =
    file.candidates.length > 0
      ? Math.max(...file.candidates.map((c) => c.score))
      : 1;

  return `<div class="detail-content">
    <div class="detail-header">
      <h3>${esc(file.path)}</h3>
      <button class="btn btn-sm btn-ghost" id="close-detail">\u2715</button>
    </div>
    <div class="detail-section">
      <div class="detail-row"><span class="detail-label">Type</span>${badge(file.type, file.type === 'test' ? 'success' : 'muted')}</div>
      <div class="detail-row"><span class="detail-label">Framework</span>${frameworkBadge(file.framework)}</div>
      <div class="detail-row"><span class="detail-label">Confidence</span>${confidenceBadge(file.confidence)}</div>
    </div>
    ${
      file.candidates.length > 0
        ? `<div class="detail-section">
      <h4>Detection Candidates</h4>
      ${file.candidates
        .map(
          (c) =>
            `<div class="candidate-row">
          <div class="candidate-info">${frameworkBadge(c.framework)} <span class="text-secondary">score: ${c.score}</span></div>
          <div class="candidate-bar"><div class="candidate-bar-fill" style="width:${Math.round((c.score / maxScore) * 100)}%"></div></div>
        </div>`
        )
        .join('')}
    </div>`
        : ''
    }
    ${
      file.warnings.length > 0
        ? `<div class="detail-section">
      <h4>Warnings</h4>
      ${file.warnings.map((w) => `<div class="warning-item">${esc(w)}</div>`).join('')}
    </div>`
        : ''
    }
    <div class="detail-section">
      <h4>Convert To</h4>
      ${
        dirs.length > 0
          ? dirs
              .map(
                (d) =>
                  `<div class="detail-row">${frameworkBadge(d.to)} ${d.pipelineBacked ? badge('pipeline', 'success') : badge('legacy', 'muted')}</div>`
              )
              .join('')
          : `<div class="text-secondary" style="font-size:12px">${esc(file.framework || 'unknown')} is a target-only framework</div>`
      }
    </div>
    ${
      dirs.length === 0 && inboundDirs.length > 0
        ? `<div class="detail-section">
      <h4>Convert From</h4>
      <div class="text-secondary" style="font-size:12px;margin-bottom:6px">Other frameworks can convert to ${esc(file.framework)}</div>
      ${inboundDirs
        .map(
          (d) =>
            `<div class="detail-row">${frameworkBadge(d.from)} \u2192 ${frameworkBadge(d.to)} ${d.pipelineBacked ? badge('pipeline', 'success') : badge('legacy', 'muted')}</div>`
        )
        .join('')}
    </div>`
        : ''
    }
    ${
      cmd
        ? `<div class="detail-section">
      <h4>Recommended Command</h4>
      <div class="command-block"><code>${esc(cmd)}</code>${copyBtn(cmd)}</div>
    </div>`
        : ''
    }
    ${
      file.type === 'test' && dirs.length > 0
        ? `<div class="detail-section">
      <h4>Preview Conversion</h4>
      <div class="form-row">
        <select class="select-input" id="preview-target">
          <option value="">Select target\u2026</option>
          ${dirs.map((d) => `<option value="${esc(d.to)}">${esc(d.to)}${d.pipelineBacked ? ' (pipeline)' : ''}</option>`).join('')}
        </select>
        <button class="btn btn-primary btn-sm" id="preview-btn" disabled>Preview</button>
      </div>
      <div id="preview-result"></div>
    </div>`
        : ''
    }
    <div class="detail-section">
      <h4>Source Preview</h4>
      <div class="source-preview" id="source-preview">
        <div class="source-loading">${spinner()} Loading...</div>
      </div>
    </div>
    ${file.type === 'test' && dirs.length > 0 ? renderConvertForm(file, dirs, rpt, actions) : ''}
  </div>`;
}

function renderConvertForm(file, dirs, _rpt, _actions) {
  return `<div class="detail-section convert-section">
    <h4>Convert File</h4>
    <div class="form-group">
      <label>Target Framework</label>
      <select class="select-input" id="target-fw">
        ${dirs.map((d) => `<option value="${esc(d.to)}">${esc(d.to)}${d.pipelineBacked ? ' (pipeline)' : ''}</option>`).join('')}
      </select>
    </div>
    <div class="form-group">
      <label>Output Mode</label>
      <select class="select-input" id="output-mode">
        <option value="out-dir" selected>Output directory</option>
        <option value="in-place">In-place</option>
      </select>
    </div>
    <div class="form-group" id="outdir-group">
      <label>Output Directory</label>
      <input type="text" class="text-input" id="output-dir" value="./hamlet-out" />
    </div>
    <button class="btn btn-primary" id="run-convert-btn">Convert</button>
  </div>`;
}

function attachEvents(container, state, actions) {
  // Sort
  container.querySelectorAll('.sortable').forEach((th) => {
    th.addEventListener('click', () => {
      const col = th.dataset.col;
      if (sortCol === col) sortAsc = !sortAsc;
      else {
        sortCol = col;
        sortAsc = true;
      }
      renderContent(container, state, actions);
    });
  });

  // Search
  const search = container.querySelector('#search');
  if (search) {
    search.addEventListener('input', () => {
      searchQuery = search.value;
      renderContent(container, state, actions);
    });
    search.focus();
  }

  // Type filter
  const tf = container.querySelector('#type-filter');
  if (tf)
    tf.addEventListener('change', () => {
      typeFilter = tf.value;
      renderContent(container, state, actions);
    });

  // File row click
  container.querySelectorAll('.file-row').forEach((row) => {
    row.addEventListener('click', () => {
      const p = row.dataset.path;
      selectedFile = report.files.find((f) => f.path === p) || null;
      renderContent(container, state, actions);
    });
  });

  // Load source preview
  const sourcePreview = container.querySelector('#source-preview');
  if (sourcePreview && selectedFile) {
    loadSourcePreview(sourcePreview, selectedFile.path, state.root);
  }

  // Preview conversion
  const previewTarget = container.querySelector('#preview-target');
  const previewBtn = container.querySelector('#preview-btn');
  const previewResult = container.querySelector('#preview-result');
  const previewFullWidth = container.querySelector('#preview-full-width');
  if (previewTarget && previewBtn) {
    previewTarget.addEventListener('change', () => {
      previewBtn.disabled = !previewTarget.value;
    });
    previewBtn.addEventListener('click', async () => {
      if (!previewTarget.value || !selectedFile) return;
      previewBtn.disabled = true;
      previewBtn.textContent = 'Converting\u2026';
      previewResult.innerHTML = `<div class="source-loading">${spinner()} Converting...</div>`;
      if (previewFullWidth) previewFullWidth.innerHTML = '';
      try {
        const absPath = selectedFile.path.startsWith('/')
          ? selectedFile.path
          : state.root + '/' + selectedFile.path;
        const data = await api.preview(
          absPath,
          selectedFile.framework,
          previewTarget.value
        );
        // Render diff into full-width panel below the table
        const target = previewFullWidth || previewResult;
        renderPreviewDiff(
          target,
          data.source,
          data.converted,
          previewTarget.value,
          selectedFile.path
        );
        previewResult.innerHTML = `<div class="text-secondary" style="font-size:12px">\u2193 Preview shown below</div>`;
        if (previewFullWidth)
          previewFullWidth.scrollIntoView({
            behavior: 'smooth',
            block: 'start',
          });
      } catch (err) {
        previewResult.innerHTML = `<div class="error-msg">Preview failed: ${esc(err.message)}</div>`;
      }
      previewBtn.disabled = false;
      previewBtn.textContent = 'Preview';
    });
  }

  // Close detail
  const close = container.querySelector('#close-detail');
  if (close)
    close.addEventListener('click', () => {
      selectedFile = null;
      renderContent(container, state, actions);
    });

  // Copy buttons
  container.querySelectorAll('.copy-btn').forEach((b) => {
    b.addEventListener('click', (e) => {
      e.stopPropagation();
      navigator.clipboard.writeText(b.dataset.copy);
      b.textContent = 'Copied!';
      setTimeout(() => (b.innerHTML = b.title ? b.innerHTML : 'Copy'), 1500);
    });
  });

  // Download report
  const dl = container.querySelector('#dl-report');
  if (dl)
    dl.addEventListener('click', () =>
      api.downloadJson(report, 'hamlet-analysis.json')
    );

  // Refresh
  const ref = container.querySelector('#refresh-btn');
  if (ref)
    ref.addEventListener('click', async () => {
      report = null;
      selectedFile = null;
      render(container, state, actions);
    });

  // Output mode toggle
  const om = container.querySelector('#output-mode');
  const og = container.querySelector('#outdir-group');
  if (om && og) {
    om.addEventListener('change', () => {
      og.style.display = om.value === 'in-place' ? 'none' : '';
    });
  }

  // Run convert (single file)
  const runBtn = container.querySelector('#run-convert-btn');
  if (runBtn) {
    runBtn.addEventListener('click', async () => {
      const targetFw = container.querySelector('#target-fw').value;
      const outputMode = container.querySelector('#output-mode').value;
      const outputDir =
        container.querySelector('#output-dir')?.value || './hamlet-out';
      try {
        const { jobId } = await api.startConvert({
          root: state.root,
          direction: { from: selectedFile.framework, to: targetFw },
          outputMode,
          outputDir: outputMode === 'out-dir' ? outputDir : undefined,
          includeFiles: [selectedFile.path],
        });
        actions.navigate(`/runs/${jobId}`);
      } catch (err) {
        alert('Convert failed: ' + err.message);
      }
    });
  }

  // Convert all tests
  const allBtn = container.querySelector('#convert-all-btn');
  if (allBtn) {
    allBtn.addEventListener('click', () => {
      actions.setState({ convertAll: true, analysisReport: report });
      actions.navigate('/convert-all');
    });
  }
}

function renderPreviewDiff(el, source, converted, targetFw, filePath) {
  const diffEntries = computeDiff(source, converted);
  const pairs = buildSideBySide(diffEntries);
  const added = diffEntries.filter((e) => e.type === 'insert').length;
  const removed = diffEntries.filter((e) => e.type === 'delete').length;
  const unchanged = diffEntries.filter((e) => e.type === 'equal').length;

  el.innerHTML = `
    <div class="preview-full-header">
      <div>
        <h3>Preview: ${esc(filePath || 'source')} \u2192 ${esc(targetFw)}</h3>
        <div class="preview-stats">
          <span class="text-success">+${added}</span>
          <span class="text-danger">-${removed}</span>
          <span class="text-secondary">${unchanged} unchanged</span>
        </div>
      </div>
      <button class="btn btn-sm btn-ghost" id="close-preview">\u2715 Close</button>
    </div>
    <div class="diff-container">
      <div class="diff-pane">
        <div class="diff-pane-header">Original</div>
        <div class="diff-lines preview-scroll">${pairs.map((p) => renderPreviewLeft(p)).join('')}</div>
      </div>
      <div class="diff-pane">
        <div class="diff-pane-header">Converted</div>
        <div class="diff-lines preview-scroll">${pairs.map((p) => renderPreviewRight(p)).join('')}</div>
      </div>
    </div>`;

  // Sync scroll between panes
  const panes = el.querySelectorAll('.preview-scroll');
  if (panes.length === 2) {
    let syncing = false;
    const sync = (src, tgt) => {
      if (syncing) return;
      syncing = true;
      tgt.scrollTop = src.scrollTop;
      syncing = false;
    };
    panes[0].addEventListener('scroll', () => sync(panes[0], panes[1]));
    panes[1].addEventListener('scroll', () => sync(panes[1], panes[0]));
  }

  // Close preview button
  const closePreview = el.querySelector('#close-preview');
  if (closePreview) {
    closePreview.addEventListener('click', () => {
      el.innerHTML = '';
    });
  }
}

function renderPreviewLeft(pair) {
  if (pair.type === 'insert')
    return '<div class="diff-line diff-empty"><span class="diff-num"></span><span class="diff-code"></span></div>';
  const cls = pair.type === 'delete' ? 'diff-removed' : '';
  return `<div class="diff-line ${cls}"><span class="diff-num">${pair.leftNum}</span><span class="diff-code">${esc(pair.left)}</span></div>`;
}

function renderPreviewRight(pair) {
  if (pair.type === 'delete')
    return '<div class="diff-line diff-empty"><span class="diff-num"></span><span class="diff-code"></span></div>';
  const cls = pair.type === 'insert' ? 'diff-added' : '';
  const content = pair.right || '';
  return `<div class="diff-line ${cls}"><span class="diff-num">${pair.rightNum}</span><span class="diff-code">${esc(content)}</span></div>`;
}

async function loadSourcePreview(el, filePath, root) {
  try {
    const absPath = filePath.startsWith('/') ? filePath : root + '/' + filePath;
    const { content } = await api.readFile(absPath);
    const lines = content.split('\n');
    const maxLines = 30;
    const truncated = lines.length > maxLines;
    const shown = lines.slice(0, maxLines);
    el.innerHTML =
      `<pre class="source-code"><code>${shown.map((l, i) => `<span class="line-num">${String(i + 1).padStart(3)}</span> ${esc(l)}`).join('\n')}</code></pre>` +
      (truncated
        ? `<div class="source-truncated">${lines.length - maxLines} more lines\u2026</div>`
        : '');
  } catch (_err) {
    el.innerHTML =
      '<div class="text-secondary" style="font-size:12px;padding:8px">Unable to load source file</div>';
  }
}

export function reset() {
  report = null;
  selectedFile = null;
  searchQuery = '';
  typeFilter = '';
  sortCol = 'path';
  sortAsc = true;
}
