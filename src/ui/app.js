/* global document, window, history, location, navigator, alert */
import * as api from './api.js';
import * as analyzeView from './analyze.js';
import * as runView from './run.js';
import * as diffView from './diffview.js';
import { esc, emptyState, frameworkBadge } from './components.js';

const app = document.getElementById('app');

const state = {
  root: '.',
  route: { path: '/analyze', params: {} },
  convertAll: false,
  analysisReport: null,
  diffSource: null,
  diffOutput: null,
};

function setState(updates) {
  Object.assign(state, updates);
}

function navigate(path) {
  history.pushState(null, '', path);
  route(path);
}

function route(pathname) {
  pathname = pathname || location.pathname;
  const params = {};

  if (pathname.startsWith('/runs/')) {
    params.id = pathname.split('/runs/')[1];
    state.route = { path: '/runs/:id', params };
  } else if (pathname === '/diff') {
    state.route = { path: '/diff', params };
  } else if (pathname === '/convert-all') {
    state.route = { path: '/convert-all', params };
  } else {
    state.route = { path: '/analyze', params };
  }

  render();
}

function render() {
  // Update active nav link
  document.querySelectorAll('.nav-link').forEach((a) => {
    a.classList.toggle(
      'active',
      a.getAttribute('href') === state.route.path ||
        (a.getAttribute('href') === '/analyze' &&
          state.route.path === '/analyze')
    );
  });

  const actions = { setState, navigate };

  switch (state.route.path) {
    case '/analyze':
      runView.cleanup();
      analyzeView.render(app, state, actions);
      break;
    case '/runs/:id':
      analyzeView.reset();
      runView.render(app, state, actions);
      break;
    case '/diff':
      diffView.render(app, state, actions);
      break;
    case '/convert-all':
      renderConvertAll(app, state, actions);
      break;
    default:
      app.innerHTML = emptyState('Page not found');
  }
}

async function renderConvertAll(container, state, _actions) {
  const report = state.analysisReport;
  if (!report) {
    navigate('/analyze');
    return;
  }

  const testFiles = report.files.filter(
    (f) => f.type === 'test' && f.framework
  );
  const frameworks = [...new Set(testFiles.map((f) => f.framework))];
  const dirs = report.summary.directionsSupported;

  container.innerHTML = `
    <div class="view-header">
      <h1>Batch Convert</h1>
      <div class="view-actions">
        <button class="btn btn-ghost" id="back-btn">\u2190 Back</button>
      </div>
    </div>
    <div class="batch-config">
      <p>${testFiles.length} test files across ${frameworks.length} framework(s): ${frameworks.map((f) => frameworkBadge(f)).join(' ')}</p>
      <div class="form-group">
        <label>Source Framework</label>
        <select class="select-input" id="batch-from">
          ${frameworks.map((f) => `<option value="${esc(f)}">${esc(f)}</option>`).join('')}
        </select>
      </div>
      <div class="form-group">
        <label>Target Framework</label>
        <select class="select-input" id="batch-to">
          ${dirs
            .filter((d) => d.from === frameworks[0])
            .map((d) => `<option value="${esc(d.to)}">${esc(d.to)}</option>`)
            .join('')}
        </select>
      </div>
      <div class="form-group">
        <label>Output Directory</label>
        <input type="text" class="text-input" id="batch-outdir" value="./hamlet-out" />
      </div>
      <button class="btn btn-primary" id="batch-run">Run Conversion</button>
    </div>`;

  // Update target options when source changes
  const fromSel = container.querySelector('#batch-from');
  const toSel = container.querySelector('#batch-to');
  fromSel.addEventListener('change', () => {
    const from = fromSel.value;
    const targets = dirs.filter((d) => d.from === from);
    toSel.innerHTML = targets
      .map((d) => `<option value="${esc(d.to)}">${esc(d.to)}</option>`)
      .join('');
  });

  container
    .querySelector('#back-btn')
    .addEventListener('click', () => navigate('/analyze'));

  container.querySelector('#batch-run').addEventListener('click', async () => {
    const from = fromSel.value;
    const to = toSel.value;
    const outdir =
      container.querySelector('#batch-outdir').value || './hamlet-out';
    try {
      const { jobId } = await api.startConvert({
        root: state.root,
        direction: { from, to },
        outputMode: 'out-dir',
        outputDir: outdir,
      });
      navigate(`/runs/${jobId}`);
    } catch (err) {
      alert('Conversion failed: ' + err.message);
    }
  });
}

// ── Init ─────────────────────────────────────────────────────────────

async function init() {
  try {
    const health = await api.getHealth();
    state.root = health.root || '.';
  } catch (_e) {
    state.root = '.';
  }

  // Nav click handlers
  document.querySelectorAll('.nav-link').forEach((a) => {
    a.addEventListener('click', (e) => {
      e.preventDefault();
      navigate(a.getAttribute('href'));
    });
  });

  window.addEventListener('popstate', () => route());

  // Global copy-button delegation
  document.addEventListener('click', (e) => {
    const copyBtn = e.target.closest('.copy-btn');
    if (copyBtn) {
      navigator.clipboard.writeText(copyBtn.dataset.copy);
    }
  });

  route(location.pathname);
}

init();
