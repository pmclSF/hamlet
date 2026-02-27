import { randomUUID } from 'node:crypto';

/** Maximum number of completed/failed jobs to retain. */
export const MAX_COMPLETED_JOBS = 100;

/** Maximum log lines per job before truncation. */
export const MAX_LOG_LINES = 5000;

/** Time in ms after completion before a job is evicted (30 minutes). */
export const JOB_TTL_MS = 30 * 60 * 1000;

/** @type {Map<string, Object>} */
const jobs = new Map();

/** @type {Map<string, Set<Function>>} */
const listeners = new Map();

/** @type {Set<ReturnType<typeof setTimeout>>} */
const ttlTimers = new Set();

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
 * When a job transitions to completed or failed, schedule TTL eviction
 * and enforce the MAX_COMPLETED_JOBS cap.
 * @param {string} id
 * @param {Object} fields
 */
export function updateJob(id, fields) {
  const job = jobs.get(id);
  if (!job) return;
  Object.assign(job, fields);
  emit(id, { type: 'status', data: { status: job.status } });

  if (job.status === 'completed' || job.status === 'failed') {
    _scheduleEviction(id);
    _enforceCompletedCap();
  }
}

/**
 * Append a log line to a job and emit a log event.
 * Truncates old entries when MAX_LOG_LINES is exceeded.
 * @param {string} id
 * @param {string} line
 */
export function appendLog(id, line) {
  const job = jobs.get(id);
  if (!job) return;
  job.log.push(line);

  if (job.log.length > MAX_LOG_LINES) {
    const dropped = job.log.length - MAX_LOG_LINES;
    job.log = job.log.slice(dropped);
    job.log[0] = `[...${dropped} earlier lines truncated]`;
  }

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

/**
 * Remove all jobs, listeners, and timers. Used in tests.
 */
export function _resetStore() {
  jobs.clear();
  listeners.clear();
  for (const timer of ttlTimers) clearTimeout(timer);
  ttlTimers.clear();
}

function emit(id, event) {
  const set = listeners.get(id);
  if (!set) return;
  for (const fn of set) {
    fn(event);
  }
}

/**
 * Schedule removal of a completed/failed job after JOB_TTL_MS.
 */
function _scheduleEviction(id) {
  const timer = setTimeout(() => {
    const job = jobs.get(id);
    // Never evict running/queued jobs (safety check)
    if (job && (job.status === 'completed' || job.status === 'failed')) {
      jobs.delete(id);
      listeners.delete(id);
    }
    ttlTimers.delete(timer);
  }, JOB_TTL_MS);
  // Prevent the timer from keeping the process alive
  timer.unref();
  ttlTimers.add(timer);
}

/**
 * If total completed/failed jobs exceed MAX_COMPLETED_JOBS, evict the oldest.
 */
function _enforceCompletedCap() {
  const completed = [];
  for (const [id, job] of jobs) {
    if (job.status === 'completed' || job.status === 'failed') {
      completed.push({ id, finishedAt: job.finishedAt || job.createdAt });
    }
  }

  if (completed.length <= MAX_COMPLETED_JOBS) return;

  // Sort oldest first
  completed.sort((a, b) => a.finishedAt.localeCompare(b.finishedAt));

  const toEvict = completed.length - MAX_COMPLETED_JOBS;
  for (let i = 0; i < toEvict; i++) {
    jobs.delete(completed[i].id);
    listeners.delete(completed[i].id);
  }
}
