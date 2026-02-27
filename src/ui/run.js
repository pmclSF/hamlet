/* global document */
import * as api from './api.js';
import { esc, card, statusBadge, spinner, emptyState } from './components.js';

let job = null;
let logs = [];
let closeStream = null;

export async function render(container, state, actions) {
  const jobId = state.route.params.id;
  if (!jobId) {
    container.innerHTML = emptyState('No job ID specified');
    return;
  }

  if (!job || job.id !== jobId) {
    logs = [];
    job = null;
    container.innerHTML = `<div class="loading">${spinner()} Loading job...</div>`;
    try {
      job = await api.getJob(jobId);
      logs = [...job.log];
    } catch (err) {
      container.innerHTML = `<div class="error-msg">Job not found: ${esc(err.message)}</div>`;
      return;
    }
  }

  renderContent(container, state, actions);

  // Subscribe to SSE if job is still running
  if (job.status === 'queued' || job.status === 'running') {
    if (closeStream) closeStream();
    closeStream = api.subscribeJob(jobId, {
      onLog(line) {
        logs.push(line);
        appendLogLine(container, line);
      },
      onStatus(data) {
        job.status = data.status;
        updateStatusBadge(container);
      },
      onDone() {
        closeStream = null;
        // Re-fetch full job for result data
        api.getJob(jobId).then((j) => {
          job = j;
          logs = [...j.log];
          renderContent(container, state, actions);
        });
      },
    });
  }
}

function renderContent(container, state, actions) {
  const isDone = job.status === 'completed' || job.status === 'failed';
  const result = job.result || {};
  const files = result.files || [];
  const totalTodos = files.reduce((s, f) => s + (f.todosAdded || 0), 0);

  container.innerHTML = `
    <div class="view-header">
      <h1>Conversion Run</h1>
      <div class="view-actions">
        <button class="btn btn-ghost" id="back-btn">\u2190 Back to Analysis</button>
        ${isDone ? `<button class="btn btn-ghost" id="dl-result">Download Report</button>` : ''}
      </div>
    </div>
    <div class="cards-row">
      ${card('Status', '')}
      ${job.result ? card('Converted', result.filesConverted || 0) : card('Converted', '--')}
      ${job.result ? card('Failed', result.filesFailed || 0) : card('Failed', '--')}
      ${job.result ? card('TODOs Added', totalTodos) : card('TODOs', '--')}
    </div>
    <div id="status-container" class="status-line">${statusBadge(job.status)} ${job.status === 'running' ? spinner() : ''}</div>
    <div class="logs-section">
      <h3>Logs</h3>
      <div class="logs-panel" id="logs-panel">
        ${logs.map((l) => `<div class="log-line">${esc(l)}</div>`).join('')}
      </div>
    </div>
    ${isDone && files.length > 0 ? renderResults(files, state, actions) : ''}`;

  attachEvents(container, state, actions);
  scrollLogs(container);
}

function renderResults(files, _state, _actions) {
  return `<div class="results-section">
    <h3>Results</h3>
    <div class="table-wrap"><table class="data-table">
      <thead><tr>
        <th>Source</th>
        <th>Status</th>
        <th>Output</th>
        <th>TODOs</th>
        <th>Actions</th>
      </tr></thead>
      <tbody>
        ${files
          .map(
            (f) => `<tr>
          <td class="mono">${esc(f.source)}</td>
          <td>${statusBadge(f.status)}</td>
          <td class="mono">${f.outputPath ? esc(f.outputPath) : '--'}</td>
          <td>${f.todosAdded || 0}</td>
          <td>${f.status === 'converted' && f.outputPath ? `<button class="btn btn-sm btn-ghost view-diff-btn" data-source="${esc(f.source)}" data-output="${esc(f.outputPath)}">View Diff</button> <button class="btn btn-sm btn-ghost open-btn" data-path="${esc(f.outputPath)}">Open</button>` : f.error ? `<span class="text-danger">${esc(f.error)}</span>` : ''}</td>
        </tr>`
          )
          .join('')}
      </tbody>
    </table></div>
    ${
      job.result
        ? `<div class="result-actions">
      <button class="btn btn-ghost" id="open-outdir">Open Output Folder</button>
    </div>`
        : ''
    }
  </div>`;
}

function appendLogLine(container, line) {
  const panel = container.querySelector('#logs-panel');
  if (!panel) return;
  const div = document.createElement('div');
  div.className = 'log-line';
  div.textContent = line;
  panel.appendChild(div);
  scrollLogs(container);
}

function updateStatusBadge(container) {
  const sc = container.querySelector('#status-container');
  if (sc)
    sc.innerHTML = `${statusBadge(job.status)} ${job.status === 'running' ? spinner() : ''}`;
}

function scrollLogs(container) {
  const panel = container.querySelector('#logs-panel');
  if (panel) panel.scrollTop = panel.scrollHeight;
}

function attachEvents(container, state, actions) {
  const back = container.querySelector('#back-btn');
  if (back) back.addEventListener('click', () => actions.navigate('/analyze'));

  const dl = container.querySelector('#dl-result');
  if (dl && job.result)
    dl.addEventListener('click', () =>
      api.downloadJson(job, 'hamlet-conversion.json')
    );

  // View diff buttons
  container.querySelectorAll('.view-diff-btn').forEach((b) => {
    b.addEventListener('click', () => {
      const source = b.dataset.source;
      const output = b.dataset.output;
      actions.setState({
        diffSource: state.root + '/' + source,
        diffOutput: output,
      });
      actions.navigate('/diff');
    });
  });

  // Open file buttons
  container.querySelectorAll('.open-btn').forEach((b) => {
    b.addEventListener('click', () => api.openPath(b.dataset.path));
  });

  // Open output folder
  const openDir = container.querySelector('#open-outdir');
  if (openDir && job.params) {
    openDir.addEventListener('click', () => {
      const dir = job.params.outputDir || './hamlet-out';
      api.openPath(dir);
    });
  }
}

export function cleanup() {
  if (closeStream) {
    closeStream();
    closeStream = null;
  }
  job = null;
  logs = [];
}
