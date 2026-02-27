/* global EventSource, document */
const BASE = '';

async function request(method, path, body) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
  };
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(`${BASE}${path}`, opts);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  return data;
}

export async function getHealth() {
  return request('GET', '/api/health');
}

export async function analyze(root, options = {}) {
  return request('POST', '/api/analyze', { root, ...options });
}

export async function startConvert(params) {
  return request('POST', '/api/convert', params);
}

export async function getJob(id) {
  return request('GET', `/api/jobs/${id}`);
}

export async function getArtifacts(jobId) {
  return request('GET', `/api/artifacts/${jobId}`);
}

export async function readFile(filePath) {
  const res = await fetch(
    `${BASE}/api/file?path=${encodeURIComponent(filePath)}`
  );
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  return data;
}

export async function openPath(filePath) {
  return request('POST', '/api/open', { path: filePath });
}

/**
 * Subscribe to SSE job events. Returns a close function.
 */
export function subscribeJob(jobId, { onLog, onStatus, onDone }) {
  const es = new EventSource(`${BASE}/api/jobs/${jobId}/stream`);
  es.addEventListener('log', (e) => onLog && onLog(JSON.parse(e.data)));
  es.addEventListener(
    'status',
    (e) => onStatus && onStatus(JSON.parse(e.data))
  );
  es.addEventListener('done', (e) => {
    onDone && onDone(JSON.parse(e.data));
    es.close();
  });
  es.onerror = () => es.close();
  return () => es.close();
}

export function downloadJson(data, filename) {
  const blob = new Blob([JSON.stringify(data, null, 2)], {
    type: 'application/json',
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}
