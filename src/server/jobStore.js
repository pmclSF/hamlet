import { randomUUID } from 'node:crypto';

/** @type {Map<string, Object>} */
const jobs = new Map();

/** @type {Map<string, Set<Function>>} */
const listeners = new Map();

/**
 * Create a new job and store it.
 * @param {Object} params - Conversion parameters for the job
 * @returns {Object} The created job
 */
export function createJob(params) {
  const id = randomUUID();
  const job = {
    id,
    status: 'queued',
    params,
    createdAt: new Date().toISOString(),
    startedAt: null,
    finishedAt: null,
    log: [],
    result: null,
    error: null,
  };
  jobs.set(id, job);
  return job;
}

/**
 * Retrieve a job by ID.
 * @param {string} id
 * @returns {Object|null}
 */
export function getJob(id) {
  return jobs.get(id) || null;
}

/**
 * Merge fields into an existing job and emit a status event.
 * @param {string} id
 * @param {Object} fields
 */
export function updateJob(id, fields) {
  const job = jobs.get(id);
  if (!job) return;
  Object.assign(job, fields);
  emit(id, { type: 'status', data: { status: job.status } });
}

/**
 * Append a log line to a job and emit a log event.
 * @param {string} id
 * @param {string} line
 */
export function appendLog(id, line) {
  const job = jobs.get(id);
  if (!job) return;
  job.log.push(line);
  emit(id, { type: 'log', data: line });
}

/**
 * Subscribe to events for a job.
 * @param {string} id
 * @param {Function} fn - Receives { type, data }
 */
export function onJobEvent(id, fn) {
  if (!listeners.has(id)) {
    listeners.set(id, new Set());
  }
  listeners.get(id).add(fn);
}

/**
 * Unsubscribe from events for a job.
 * @param {string} id
 * @param {Function} fn
 */
export function offJobEvent(id, fn) {
  const set = listeners.get(id);
  if (set) {
    set.delete(fn);
    if (set.size === 0) listeners.delete(id);
  }
}

function emit(id, event) {
  const set = listeners.get(id);
  if (!set) return;
  for (const fn of set) {
    fn(event);
  }
}
